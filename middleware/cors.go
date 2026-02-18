package middleware

import (
	"net/http"

	"github.com/rs/cors"
)

// CORSConfig configures the CORS middleware.
// This is a thin wrapper around github.com/rs/cors.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a permissive CORS config (dev).
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           3600,
	}
}

// ProductionCORSConfig returns a strict CORS config for production.
func ProductionCORSConfig(allowedOrigins []string) *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           7200,
	}
}

// CORS returns a CORS middleware with default (permissive) config.
func CORS() Middleware {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig returns a CORS middleware backed by github.com/rs/cors.
func CORSWithConfig(config *CORSConfig) Middleware {
	if config == nil {
		config = DefaultCORSConfig()
	}
	c := cors.New(cors.Options{
		AllowedOrigins:   config.AllowedOrigins,
		AllowedMethods:   config.AllowedMethods,
		AllowedHeaders:   config.AllowedHeaders,
		ExposedHeaders:   config.ExposedHeaders,
		AllowCredentials: config.AllowCredentials,
		MaxAge:           config.MaxAge,
	})
	return func(next http.Handler) http.Handler {
		return c.Handler(next)
	}
}

// AllowOrigins is a helper to allow specific origins.
func AllowOrigins(origins ...string) Middleware {
	return CORSWithConfig(&CORSConfig{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           3600,
	})
}

// AllowAll allows all origins (dev only).
func AllowAll() Middleware {
	return CORS()
}
