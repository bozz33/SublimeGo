package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/bozz33/sublimego/logger"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

// LoggerConfig configures the logger middleware.
type LoggerConfig struct {
	Logger      *logger.Logger
	SkipPaths   []string
	SkipStatus  []int
	LogBody     bool
	MaxBodySize int
}

// DefaultLoggerConfig returns a default configuration.
func DefaultLoggerConfig(log *logger.Logger) *LoggerConfig {
	return &LoggerConfig{
		Logger:      log,
		SkipPaths:   []string{"/health", "/metrics", "/favicon.ico"},
		SkipStatus:  []int{},
		LogBody:     false,
		MaxBodySize: 1024, // 1KB
	}
}

// Logger returns a middleware that logs all HTTP requests.
func Logger(log *logger.Logger) Middleware {
	return LoggerWithConfig(DefaultLoggerConfig(log))
}

// LoggerWithConfig returns a logger middleware with custom config.
func LoggerWithConfig(config *LoggerConfig) Middleware {
	if config == nil {
		config = DefaultLoggerConfig(logger.Default())
	}

	if config.Logger == nil {
		config.Logger = logger.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if lo.Contains(config.SkipPaths, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			requestID := uuid.New().String()

			reqLogger := config.Logger.With(
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", getClientIP(r),
			)

			if ua := r.Header.Get("User-Agent"); ua != "" {
				reqLogger = reqLogger.With("user_agent", ua)
			}

			ctx := logger.WithContext(r.Context(), reqLogger)
			ctx = withRequestID(ctx, requestID)

			rw := NewResponseWriter(w)

			start := time.Now()

			next.ServeHTTP(rw, r.WithContext(ctx))

			duration := time.Since(start)

			if lo.Contains(config.SkipStatus, rw.Status()) {
				return
			}

			logRequest(reqLogger, r, rw, duration)
		})
	}
}

// logRequest logs the request details.
func logRequest(log *logger.Logger, r *http.Request, rw *responseWriter, duration time.Duration) {
	attrs := []any{
		"status", rw.Status(),
		"size", rw.Size(),
		"duration", duration.String(),
	}

	if r.URL.RawQuery != "" {
		attrs = append(attrs, "query", r.URL.RawQuery)
	}

	if referer := r.Header.Get("Referer"); referer != "" {
		attrs = append(attrs, "referer", referer)
	}

	msg := "http request"

	switch {
	case rw.Status() >= 500:
		log.Error(msg, attrs...)
	case rw.Status() >= 400:
		log.Warn(msg, attrs...)
	default:
		log.Info(msg, attrs...)
	}
}

// getClientIP retrieves the client IP address.
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := lo.Map(
			lo.Filter(
				lo.Map(
					splitAndTrim(xff, ","),
					func(ip string, _ int) string {
						return ip
					},
				),
				func(ip string, _ int) bool {
					return ip != ""
				},
			),
			func(ip string, _ int) string {
				return ip
			},
		)

		if len(ips) > 0 {
			return ips[0]
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	return r.RemoteAddr
}

// splitAndTrim splits a string and trims spaces.
func splitAndTrim(s, sep string) []string {
	var result []string
	var current string

	for _, char := range s {
		if string(char) == sep {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else if char != ' ' && char != '\t' {
			current += string(char)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}

type contextKey string

const requestIDKey contextKey = "request_id"

// withRequestID adds the request ID to the context.
func withRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestID retrieves the request ID from the context.
func RequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// GetRequestID is an alias for RequestID.
func GetRequestID(r *http.Request) string {
	return RequestID(r.Context())
}

// LoggerFromRequest retrieves the logger from the request.
func LoggerFromRequest(r *http.Request) *logger.Logger {
	return logger.FromContext(r.Context())
}
