package logger

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Context keys
type contextKey string

const (
	loggerKey    contextKey = "logger"
	requestIDKey contextKey = "request_id"
)

// WithContext adds the logger to the context.
func WithContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext retrieves the logger from the context.
func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(loggerKey).(*Logger); ok {
		return l
	}
	return Default()
}

// WithRequestID adds a unique ID to the request.
func WithRequestID(ctx context.Context) (context.Context, string) {
	requestID := uuid.New().String()
	return context.WithValue(ctx, requestIDKey, requestID), requestID
}

// RequestIDFromContext retrieves the request ID.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// Middleware HTTP that adds the logger to each request's context.
func Middleware(l *Logger) func(http.Handler) http.Handler {
	if l == nil {
		l = Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ctx, requestID := WithRequestID(r.Context())

			reqLogger := l.With(
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
			)

			if ua := r.Header.Get("User-Agent"); ua != "" {
				reqLogger = reqLogger.With(slog.String("user_agent", ua))
			}

			ctx = WithContext(ctx, reqLogger)

			rw := &responseWriter{ResponseWriter: w, statusCode: 200}

			next.ServeHTTP(rw, r.WithContext(ctx))

			duration := time.Since(start)
			reqLogger.Request(r.Method, r.URL.Path, rw.statusCode, duration)
		})
	}
}

// responseWriter wrapper to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// L returns a logger from the context.
func L(ctx context.Context) *Logger {
	return FromContext(ctx)
}

// RequestID returns the request ID.
func RequestID(ctx context.Context) string {
	return RequestIDFromContext(ctx)
}
