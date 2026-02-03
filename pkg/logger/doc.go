// Package logger provides structured logging using Go's slog package.
//
// It wraps slog with additional features like request ID tracking,
// context-aware logging, and HTTP middleware for request logging.
//
// Features:
//   - Structured logging with slog
//   - Multiple output formats (text, JSON)
//   - Log levels (debug, info, warn, error)
//   - Request ID tracking
//   - Context-aware logging
//   - HTTP request logging middleware
//   - Source file information
//
// Basic usage:
//
//	logger := logger.New(&logger.Config{
//		Level:       slog.LevelInfo,
//		Environment: "production",
//		AddSource:   true,
//	})
//
//	// Log messages
//	logger.Info("server started", "port", 8080)
//	logger.Error("failed to connect", "error", err)
//
//	// With context
//	logger.With("user_id", userID).Info("user logged in")
//
//	// HTTP middleware
//	router.Use(logger.Middleware(logger))
package logger
