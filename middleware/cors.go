package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

// CORSConfig configures the CORS middleware.
type CORSConfig struct {
	AllowedOrigins     []string
	AllowedMethods     []string
	AllowedHeaders     []string
	ExposedHeaders     []string
	AllowCredentials   bool
	MaxAge             int
	OptionsPassthrough bool
}

// DefaultCORSConfig returns a permissive CORS config (dev).
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:     []string{"*"},
		AllowedMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:     []string{"Link"},
		AllowCredentials:   false,
		MaxAge:             3600,
		OptionsPassthrough: false,
	}
}

// ProductionCORSConfig returns a strict CORS config (prod).
func ProductionCORSConfig(allowedOrigins []string) *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:     allowedOrigins,
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:     []string{},
		AllowCredentials:   true,
		MaxAge:             7200,
		OptionsPassthrough: false,
	}
}

// CORS returns a CORS middleware with default config.
func CORS() Middleware {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig returns a CORS middleware with custom config.
func CORSWithConfig(config *CORSConfig) Middleware {
	if config == nil {
		config = DefaultCORSConfig()
	}

	normalizeConfig(config)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if !isOriginAllowed(origin, config.AllowedOrigins) {
				if r.Method != "OPTIONS" || config.OptionsPassthrough {
					next.ServeHTTP(w, r)
				}
				return
			}

			setCORSHeaders(w, origin, config)

			if r.Method == "OPTIONS" {
				if !config.OptionsPassthrough {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// setCORSHeaders sets CORS headers.
func setCORSHeaders(w http.ResponseWriter, origin string, config *CORSConfig) {
	// Access-Control-Allow-Origin
	if lo.Contains(config.AllowedOrigins, "*") {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Add("Vary", "Origin")
	}

	// Access-Control-Allow-Methods
	if len(config.AllowedMethods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
	}

	// Access-Control-Allow-Headers
	if len(config.AllowedHeaders) > 0 {
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
	}

	// Access-Control-Expose-Headers
	if len(config.ExposedHeaders) > 0 {
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
	}

	// Access-Control-Allow-Credentials
	if config.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	// Access-Control-Max-Age
	if config.MaxAge > 0 {
		w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
	}
}

// isOriginAllowed checks if the origin is allowed.
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return true
	}

	if lo.Contains(allowedOrigins, "*") {
		return true
	}

	if lo.Contains(allowedOrigins, origin) {
		return true
	}

	return lo.SomeBy(allowedOrigins, func(pattern string) bool {
		return matchOriginPattern(origin, pattern)
	})
}

// matchOriginPattern checks if the origin matches a pattern.
func matchOriginPattern(origin, pattern string) bool {
	// Pattern wildcard subdomain: *.example.com
	if strings.HasPrefix(pattern, "*.") {
		domain := strings.TrimPrefix(pattern, "*.")
		return strings.HasSuffix(origin, domain)
	}

	// Pattern with protocol wildcard: //*.example.com
	if strings.HasPrefix(pattern, "//*.") {
		domain := strings.TrimPrefix(pattern, "//*.")
		// Accepts http:// and https://
		return strings.HasSuffix(origin, domain)
	}

	return false
}

// normalizeConfig normalizes the configuration.
func normalizeConfig(config *CORSConfig) {
	config.AllowedOrigins = lo.Map(config.AllowedOrigins, func(origin string, _ int) string {
		return strings.TrimSpace(origin)
	})

	config.AllowedMethods = lo.Map(config.AllowedMethods, func(method string, _ int) string {
		return strings.ToUpper(strings.TrimSpace(method))
	})

	config.AllowedHeaders = lo.Map(config.AllowedHeaders, func(header string, _ int) string {
		return strings.TrimSpace(header)
	})

	config.ExposedHeaders = lo.Map(config.ExposedHeaders, func(header string, _ int) string {
		return strings.TrimSpace(header)
	})

	config.AllowedOrigins = lo.Uniq(config.AllowedOrigins)
	config.AllowedMethods = lo.Uniq(config.AllowedMethods)
	config.AllowedHeaders = lo.Uniq(config.AllowedHeaders)
	config.ExposedHeaders = lo.Uniq(config.ExposedHeaders)

	if lo.Contains(config.AllowedOrigins, "*") && config.AllowCredentials {
		config.AllowCredentials = false
	}
}

// AllowOrigins is a helper to create a simple CORS config.
func AllowOrigins(origins ...string) Middleware {
	return CORSWithConfig(&CORSConfig{
		AllowedOrigins:     origins,
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials:   true,
		MaxAge:             3600,
		OptionsPassthrough: false,
	})
}

// AllowAll is a helper to allow all origins (dev only).
func AllowAll() Middleware {
	return CORS()
}
