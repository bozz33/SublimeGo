// Package errors provides structured error handling for web applications.
//
// It defines AppError, a rich error type that includes HTTP status codes,
// error codes, messages, and optional stack traces. The package also provides
// factory functions for common HTTP errors and middleware for error recovery.
//
// Features:
//   - Structured errors with codes and status
//   - Stack trace capture
//   - HTTP error factories (NotFound, BadRequest, etc.)
//   - Error wrapping and unwrapping
//   - Panic recovery middleware
//   - Custom error pages
//
// Basic usage:
//
//	// Create specific errors
//	err := errors.NotFound("User not found")
//	err := errors.BadRequest("Invalid email format")
//
//	// Wrap existing errors
//	err := errors.Wrap(dbErr, "failed to fetch user")
//
//	// Check error type
//	if errors.IsAppError(err) {
//		appErr := errors.ToAppError(err)
//		// Handle based on status code
//	}
package errors
