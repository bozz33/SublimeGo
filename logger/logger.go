package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Config contains the logger configuration.
type Config struct {
	Environment    string
	Level          slog.Level
	OutputPath     string
	AddSource      bool
	EnableRotation bool
	MaxSizeMB      int
	MaxBackups     int
	MaxAgeDays     int
	Compress       bool
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		Environment:    "dev",
		Level:          slog.LevelDebug,
		OutputPath:     "",
		AddSource:      true,
		EnableRotation: false,
		MaxSizeMB:      100,
		MaxBackups:     3,
		MaxAgeDays:     28,
		Compress:       true,
	}
}

// Logger wraps slog.Logger with helper methods.
type Logger struct {
	*slog.Logger
	config *Config
}

// New creates a new configured logger.
func New(cfg *Config) *Logger {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	var writer io.Writer = os.Stdout

	if cfg.OutputPath != "" {
		var fileWriter io.Writer

		if cfg.EnableRotation {
			fileWriter = &lumberjack.Logger{
				Filename:   cfg.OutputPath,
				MaxSize:    cfg.MaxSizeMB,
				MaxBackups: cfg.MaxBackups,
				MaxAge:     cfg.MaxAgeDays,
				Compress:   cfg.Compress,
			}
		} else {
			file, err := os.OpenFile(cfg.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				fileWriter = os.Stdout
			} else {
				fileWriter = file
			}
		}

		writer = io.MultiWriter(os.Stdout, fileWriter)
	}

	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	}

	if cfg.Environment == "prod" || cfg.Environment == "production" {
		handler = slog.NewJSONHandler(writer, opts)
	} else {
		handler = slog.NewTextHandler(writer, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
		config: cfg,
	}
}

// With returns a new logger with default attributes.
func (l *Logger) With(attrs ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(attrs...),
		config: l.config,
	}
}

// WithGroup returns a new logger in a group.
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		Logger: l.Logger.WithGroup(name),
		config: l.config,
	}
}

// Request logs an HTTP request.
func (l *Logger) Request(method, path string, status int, duration time.Duration, attrs ...any) {
	level := slog.LevelInfo

	if status >= 500 {
		level = slog.LevelError
	} else if status >= 400 {
		level = slog.LevelWarn
	}

	baseAttrs := []any{
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", status),
		slog.Duration("duration", duration),
	}

	l.Log(context.Background(), level, "http request", append(baseAttrs, attrs...)...)
}

// Err is a helper to create an error attribute.
func Err(err error) slog.Attr {
	return slog.Any("error", err)
}

// ParseLevel converts a string to slog.Level.
func ParseLevel(level string) slog.Level {
	levelMap := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	if l, ok := levelMap[level]; ok {
		return l
	}
	return slog.LevelInfo
}

// Default global instance.
var defaultLogger = New(DefaultConfig())

// SetDefault changes the global logger.
func SetDefault(l *Logger) {
	defaultLogger = l
	slog.SetDefault(l.Logger)
}

// Default returns the global logger.
func Default() *Logger {
	return defaultLogger
}

// Global functions using the default logger.

// Debug logs a debug message.
func Debug(msg string, attrs ...any) {
	defaultLogger.Debug(msg, attrs...)
}

func Info(msg string, attrs ...any) {
	defaultLogger.Info(msg, attrs...)
}

func Warn(msg string, attrs ...any) {
	defaultLogger.Warn(msg, attrs...)
}

func Error(msg string, attrs ...any) {
	defaultLogger.Error(msg, attrs...)
}

func Request(method, path string, status int, duration time.Duration, attrs ...any) {
	defaultLogger.Request(method, path, status, duration, attrs...)
}
