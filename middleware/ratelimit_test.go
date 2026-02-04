package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimiter(t *testing.T) {
	config := &RateLimitConfig{
		RequestsPerMinute: 60,
		Burst:             10,
	}

	rl := NewRateLimiter(config)
	require.NotNil(t, rl)
	assert.Equal(t, 60, rl.config.RequestsPerMinute)
	assert.Equal(t, 10, rl.config.Burst)
	assert.NotNil(t, rl.config.KeyFunc)

	rl.Stop()
}

func TestDefaultRateLimitConfig(t *testing.T) {
	config := DefaultRateLimitConfig()
	assert.Equal(t, 60, config.RequestsPerMinute)
	assert.Equal(t, 10, config.Burst)
	assert.NotNil(t, config.KeyFunc)
	assert.Equal(t, 5*time.Minute, config.CleanupInterval)
}

func TestRateLimiter_AllowRequests(t *testing.T) {
	rl := NewRateLimiter(&RateLimitConfig{
		RequestsPerMinute: 60,
		Burst:             5,
	})
	defer rl.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrapped := rl.Middleware()(handler)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "OK", rec.Body.String())
	}
}

func TestRateLimiter_ExceedLimit(t *testing.T) {
	rl := NewRateLimiter(&RateLimitConfig{
		RequestsPerMinute: 60,
		Burst:             3,
	})
	defer rl.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware()(handler)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	// Check for rate limit message
	body := rec.Body.String()
	assert.True(t, strings.Contains(body, "Too many requests"), "Expected rate limit message")
	assert.Equal(t, "0", rec.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, rec.Header().Get("Retry-After"))
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(&RateLimitConfig{
		RequestsPerMinute: 60,
		Burst:             2,
	})
	defer rl.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware()(handler)

	ips := []string{"192.168.1.1:1234", "192.168.1.2:1234", "192.168.1.3:1234"}

	for _, ip := range ips {
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = ip
			rec := httptest.NewRecorder()
			wrapped.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code, "IP %s should be allowed", ip)
		}
	}
}

func TestRateLimiter_Whitelist(t *testing.T) {
	rl := NewRateLimiter(&RateLimitConfig{
		RequestsPerMinute: 60,
		Burst:             1,
		WhitelistIPs:      []string{"127.0.0.1", "192.168.1.100"},
	})
	defer rl.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware()(handler)

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "Whitelisted IP should never be rate limited")
	}
}

func TestRateLimiter_WhitelistCIDR(t *testing.T) {
	rl := NewRateLimiter(&RateLimitConfig{
		RequestsPerMinute: 60,
		Burst:             1,
		WhitelistIPs:      []string{"192.168.0.0/16"},
	})
	defer rl.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware()(handler)

	testIPs := []string{
		"192.168.1.1:1234",
		"192.168.50.100:5678",
		"192.168.255.255:9999",
	}

	for _, ip := range testIPs {
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = ip
			rec := httptest.NewRecorder()
			wrapped.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code, "IP %s in CIDR should be whitelisted", ip)
		}
	}
}

func TestRateLimiter_Headers(t *testing.T) {
	rl := NewRateLimiter(&RateLimitConfig{
		RequestsPerMinute: 60,
		Burst:             5,
	})
	defer rl.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware()(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	assert.Equal(t, "60", rec.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Reset"))
}

func TestRateLimiter_KeyByUser(t *testing.T) {
	rl := NewRateLimiter(&RateLimitConfig{
		RequestsPerMinute: 60,
		Burst:             2,
		KeyFunc:           KeyByUser,
	})
	defer rl.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware()(handler)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestRateLimiter_KeyByHeader(t *testing.T) {
	rl := NewRateLimiter(&RateLimitConfig{
		RequestsPerMinute: 60,
		Burst:             2,
		KeyFunc:           KeyByHeader("X-API-Key"),
	})
	defer rl.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware()(handler)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "test-key-123")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "test-key-123")
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
}

func TestRateLimiter_OnLimitExceeded(t *testing.T) {
	var called bool
	var capturedKey string

	rl := NewRateLimiter(&RateLimitConfig{
		RequestsPerMinute: 60,
		Burst:             1,
		OnLimitExceeded: func(r *http.Request, key string) {
			called = true
			capturedKey = key
		},
	})
	defer rl.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware()(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rec = httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Contains(t, capturedKey, "192.168.1.1")
}

func TestRateLimiter_Concurrent(t *testing.T) {
	rl := NewRateLimiter(&RateLimitConfig{
		RequestsPerMinute: 100,
		Burst:             50,
	})
	defer rl.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware()(handler)

	var wg sync.WaitGroup
	successCount := 0
	rateLimitCount := 0
	var mu sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = fmt.Sprintf("192.168.1.%d:1234", id%10)
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			mu.Lock()
			if rec.Code == http.StatusOK {
				successCount++
			} else if rec.Code == http.StatusTooManyRequests {
				rateLimitCount++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	assert.Greater(t, successCount, 0, "Should have some successful requests")
	t.Logf("Success: %d, Rate Limited: %d", successCount, rateLimitCount)
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := NewRateLimiter(&RateLimitConfig{
		RequestsPerMinute: 60,
		Burst:             5,
		CleanupInterval:   100 * time.Millisecond,
	})
	defer rl.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware()(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	count := 0
	rl.limiters.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	assert.Equal(t, 1, count, "Should have 1 limiter")

	time.Sleep(300 * time.Millisecond)

	count = 0
	rl.limiters.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	assert.Equal(t, 0, count, "Cleanup should have removed inactive limiters")
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
	req.RemoteAddr = "192.168.1.1:1234"

	ip := getClientIPFromRequest(req)
	assert.Equal(t, "203.0.113.1", ip)
}

func TestGetClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "203.0.113.1")
	req.RemoteAddr = "192.168.1.1:1234"

	ip := getClientIPFromRequest(req)
	assert.Equal(t, "203.0.113.1", ip)
}

func TestGetClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	ip := getClientIPFromRequest(req)
	assert.Equal(t, "192.168.1.1", ip)
}

func TestIsIPInCIDR(t *testing.T) {
	tests := []struct {
		ip       string
		cidr     string
		expected bool
	}{
		{"192.168.1.1", "192.168.0.0/16", true},
		{"192.168.255.255", "192.168.0.0/16", true},
		{"192.169.1.1", "192.168.0.0/16", false},
		{"10.0.0.1", "10.0.0.0/8", true},
		{"11.0.0.1", "10.0.0.0/8", false},
		{"127.0.0.1", "127.0.0.0/8", true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s in %s", tt.ip, tt.cidr), func(t *testing.T) {
			result := isIPInCIDR(tt.ip, tt.cidr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWithRateLimitConfig(t *testing.T) {
	rl := WithRateLimitConfig(120, 20, KeyByIP)
	require.NotNil(t, rl)
	assert.Equal(t, 120, rl.config.RequestsPerMinute)
	assert.Equal(t, 20, rl.config.Burst)
	assert.NotNil(t, rl.config.KeyFunc)

	rl.Stop()
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	rl := NewRateLimiter(&RateLimitConfig{
		RequestsPerMinute: 10000,
		Burst:             1000,
	})
	defer rl.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware()(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = fmt.Sprintf("192.168.1.%d:1234", i%256)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
	}
}

func BenchmarkRateLimiter_Concurrent(b *testing.B) {
	rl := NewRateLimiter(&RateLimitConfig{
		RequestsPerMinute: 10000,
		Burst:             1000,
	})
	defer rl.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware()(handler)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = fmt.Sprintf("192.168.1.%d:1234", i%256)
			rec := httptest.NewRecorder()
			wrapped.ServeHTTP(rec, req)
			i++
		}
	})
}
