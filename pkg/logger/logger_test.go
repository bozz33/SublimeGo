package logger

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	cfg := DefaultConfig()
	logger := New(cfg)

	require.NotNil(t, logger)
	assert.Equal(t, cfg, logger.config)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "dev", cfg.Environment)
	assert.Equal(t, slog.LevelDebug, cfg.Level)
	assert.True(t, cfg.AddSource)
}

func TestLoggerOutput(t *testing.T) {
	var buf bytes.Buffer

	cfg := &Config{
		Environment: "dev",
		Level:       slog.LevelDebug,
		AddSource:   false,
	}

	// Create a handler that writes to the buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	})

	logger := &Logger{
		Logger: slog.New(handler),
		config: cfg,
	}

	logger.Info("test message", slog.String("key", "value"))

	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "key=value")
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := &Logger{
		Logger: slog.New(handler),
		config: &Config{Level: slog.LevelInfo},
	}

	// Debug should not appear (level = Info)
	logger.Debug("debug message")
	assert.Empty(t, buf.String())

	// Info should appear
	logger.Info("info message")
	assert.Contains(t, buf.String(), "info message")

	buf.Reset()

	// Warn should appear
	logger.Warn("warn message")
	assert.Contains(t, buf.String(), "warn message")

	buf.Reset()

	// Error should appear
	logger.Error("error message")
	assert.Contains(t, buf.String(), "error message")
}

func TestRequest(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := &Logger{
		Logger: slog.New(handler),
		config: &Config{Level: slog.LevelInfo},
	}

	logger.Request("GET", "/test", 200, 100*time.Millisecond)

	output := buf.String()
	assert.Contains(t, output, "GET")
	assert.Contains(t, output, "/test")
	assert.Contains(t, output, "200")
}

func TestRequestWithErrorStatus(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := &Logger{
		Logger: slog.New(handler),
		config: &Config{Level: slog.LevelInfo},
	}

	// Status 500 -> Error level
	logger.Request("POST", "/api", 500, 50*time.Millisecond)

	output := buf.String()
	assert.Contains(t, output, "level=ERROR")

	buf.Reset()

	// Status 404 -> Warn level
	logger.Request("GET", "/missing", 404, 10*time.Millisecond)

	output = buf.String()
	assert.Contains(t, output, "level=WARN")
}

func TestWith(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := &Logger{
		Logger: slog.New(handler),
		config: &Config{Level: slog.LevelInfo},
	}

	// Logger with default attributes
	loggerWithAttrs := logger.With(
		slog.String("user_id", "123"),
		slog.String("session", "abc"),
	)

	loggerWithAttrs.Info("test")

	output := buf.String()
	assert.Contains(t, output, "user_id=123")
	assert.Contains(t, output, "session=abc")
}

func TestMiddleware(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := &Logger{
		Logger: slog.New(handler),
		config: &Config{Level: slog.LevelInfo},
	}

	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get logger from context
		l := FromContext(r.Context())
		assert.NotNil(t, l)

		// Get request ID
		requestID := RequestIDFromContext(r.Context())
		assert.NotEmpty(t, requestID)

		w.WriteHeader(http.StatusOK)
	})

	middleware := Middleware(logger)
	wrappedHandler := middleware(httpHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")

	rec := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestFromContext(t *testing.T) {
	logger := New(DefaultConfig())

	// Contexte sans logger -> retourne default
	ctx := context.Background()
	l := FromContext(ctx)
	assert.NotNil(t, l)

	// Contexte avec logger
	ctx = WithContext(ctx, logger)
	l = FromContext(ctx)
	assert.Equal(t, logger, l)
}

func TestWithRequestID(t *testing.T) {
	ctx := context.Background()

	ctx, requestID := WithRequestID(ctx)

	assert.NotEmpty(t, requestID)
	assert.Equal(t, requestID, RequestIDFromContext(ctx))
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"unknown", slog.LevelInfo}, // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseLevel(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestJSONOutput(t *testing.T) {
	var buf bytes.Buffer

	cfg := &Config{
		Environment: "prod",
		Level:       slog.LevelInfo,
	}

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: cfg.Level,
	})

	logger := &Logger{
		Logger: slog.New(handler),
		config: cfg,
	}

	logger.Info("test", slog.String("key", "value"))

	output := buf.String()
	// En production, output est JSON
	assert.True(t, strings.Contains(output, "{"))
	assert.True(t, strings.Contains(output, "\"msg\""))
}

func BenchmarkLogger(b *testing.B) {
	logger := New(DefaultConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark",
			slog.String("key", "value"),
			slog.Int("count", i),
		)
	}
}

func BenchmarkRequest(b *testing.B) {
	logger := New(DefaultConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Request("GET", "/api/test", 200, 10*time.Millisecond)
	}
}
