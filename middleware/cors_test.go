package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig()

	assert.NotNil(t, config)
	assert.Contains(t, config.AllowedOrigins, "*")
	assert.Contains(t, config.AllowedMethods, "GET")
	assert.Contains(t, config.AllowedMethods, "POST")
	assert.False(t, config.AllowCredentials) // false avec wildcard
	assert.Equal(t, 3600, config.MaxAge)
}

func TestProductionCORSConfig(t *testing.T) {
	origins := []string{"https://example.com", "https://app.example.com"}
	config := ProductionCORSConfig(origins)

	assert.Equal(t, origins, config.AllowedOrigins)
	assert.True(t, config.AllowCredentials)
	assert.Equal(t, 7200, config.MaxAge)
}

func TestCORS(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORS()
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSWithSpecificOrigin(t *testing.T) {
	config := &CORSConfig{
		AllowedOrigins:   []string{"https://example.com"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowCredentials: true,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORSWithConfig(config)
	wrappedHandler := middleware(handler)

	// Allowed origin
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, "https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "Origin", rec.Header().Get("Vary"))
}

func TestCORSWithUnauthorizedOrigin(t *testing.T) {
	config := &CORSConfig{
		AllowedOrigins: []string{"https://example.com"},
		AllowedMethods: []string{"GET"},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORSWithConfig(config)
	wrappedHandler := middleware(handler)

	// Disallowed origin
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	// No CORS header
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSPreflightRequest(t *testing.T) {
	config := &CORSConfig{
		AllowedOrigins:     []string{"*"},
		AllowedMethods:     []string{"GET", "POST", "PUT"},
		AllowedHeaders:     []string{"Content-Type", "Authorization"},
		MaxAge:             3600,
		OptionsPassthrough: false,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Handler should not be called for preflight")
	})

	middleware := CORSWithConfig(config)
	wrappedHandler := middleware(handler)

	// Preflight request
	req := httptest.NewRequest("OPTIONS", "/api/users", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	// Verify CORS headers
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
	assert.Equal(t, "3600", rec.Header().Get("Access-Control-Max-Age"))

	// Preflight should return 204
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestCORSOptionsPassthrough(t *testing.T) {
	var handlerCalled bool

	config := &CORSConfig{
		AllowedOrigins:     []string{"*"},
		AllowedMethods:     []string{"GET", "OPTIONS"},
		OptionsPassthrough: true, // Let OPTIONS pass through to handler
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORSWithConfig(config)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	// Handler should be called
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		want           bool
	}{
		{
			name:           "wildcard allows all",
			origin:         "https://example.com",
			allowedOrigins: []string{"*"},
			want:           true,
		},
		{
			name:           "exact match",
			origin:         "https://example.com",
			allowedOrigins: []string{"https://example.com"},
			want:           true,
		},
		{
			name:           "no match",
			origin:         "https://evil.com",
			allowedOrigins: []string{"https://example.com"},
			want:           false,
		},
		{
			name:           "subdomain wildcard match",
			origin:         "https://app.example.com",
			allowedOrigins: []string{"*.example.com"},
			want:           true,
		},
		{
			name:           "protocol wildcard match",
			origin:         "https://app.example.com",
			allowedOrigins: []string{"//*.example.com"},
			want:           true,
		},
		{
			name:           "empty origin (same-origin)",
			origin:         "",
			allowedOrigins: []string{"https://example.com"},
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isOriginAllowed(tt.origin, tt.allowedOrigins)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMatchOriginPattern(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		pattern string
		want    bool
	}{
		{
			name:    "subdomain wildcard match",
			origin:  "https://app.example.com",
			pattern: "*.example.com",
			want:    true,
		},
		{
			name:    "subdomain wildcard no match",
			origin:  "https://other.com",
			pattern: "*.example.com",
			want:    false,
		},
		{
			name:    "protocol wildcard match",
			origin:  "http://app.example.com",
			pattern: "//*.example.com",
			want:    true,
		},
		{
			name:    "exact pattern (not a wildcard)",
			origin:  "https://example.com",
			pattern: "https://example.com",
			want:    false, // matchOriginPattern only handles wildcards
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchOriginPattern(tt.origin, tt.pattern)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNormalizeConfig(t *testing.T) {
	config := &CORSConfig{
		AllowedOrigins:   []string{" https://example.com ", "https://example.com", " http://test.com "},
		AllowedMethods:   []string{" get ", "POST", "get"},
		AllowedHeaders:   []string{" Content-Type ", " authorization "},
		AllowCredentials: true,
	}

	normalizeConfig(config)

	// Verify trim
	assert.Equal(t, "https://example.com", config.AllowedOrigins[0])
	assert.Equal(t, "http://test.com", config.AllowedOrigins[1])

	// Verify uppercase methods
	assert.Contains(t, config.AllowedMethods, "GET")
	assert.Contains(t, config.AllowedMethods, "POST")

	// Verify uniq (get was duplicated)
	count := 0
	for _, m := range config.AllowedMethods {
		if m == "GET" {
			count++
		}
	}
	assert.Equal(t, 1, count, "GET should appear only once")

	// Verify trim headers
	assert.Equal(t, "Content-Type", config.AllowedHeaders[0])
	assert.Equal(t, "authorization", config.AllowedHeaders[1])
}

func TestNormalizeConfigWildcardCredentials(t *testing.T) {
	config := &CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true, // Invalid with wildcard
	}

	normalizeConfig(config)

	// Credentials should be forced to false with wildcard
	assert.False(t, config.AllowCredentials)
}

func TestAllowOrigins(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AllowOrigins("https://example.com", "https://app.example.com")
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, "https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
}

func TestAllowAll(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AllowAll()
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://anywhere.com")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
}

func BenchmarkCORS(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORS()
	wrappedHandler := middleware(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Origin", "https://example.com")
		rec := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, req)
	}
}
