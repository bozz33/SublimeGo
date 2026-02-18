package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/bozz33/sublimego/errors"
	"github.com/bozz33/sublimego/logger"
)

// RecoveryConfig configures the recovery middleware.
type RecoveryConfig struct {
	ErrorHandler *errors.Handler
	Logger       *logger.Logger
	PrintStack   bool
	OnPanic      func(r *http.Request, rec any)
}

// DefaultRecoveryConfig returns a default configuration.
func DefaultRecoveryConfig(errorHandler *errors.Handler) *RecoveryConfig {
	return &RecoveryConfig{
		ErrorHandler: errorHandler,
		Logger:       logger.Default(),
		PrintStack:   false,
		OnPanic:      nil,
	}
}

// Recovery returns a middleware that captures panics.
func Recovery(errorHandler *errors.Handler) Middleware {
	return RecoveryWithConfig(DefaultRecoveryConfig(errorHandler))
}

// RecoveryWithConfig returns a recovery middleware with custom config.
func RecoveryWithConfig(config *RecoveryConfig) Middleware {
	if config == nil {
		config = DefaultRecoveryConfig(nil)
	}

	if config.Logger == nil {
		config.Logger = logger.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					// Capture the stack trace
					stack := debug.Stack()

					// Get logger from context if available
					log := logger.FromContext(r.Context())
					if log == nil {
						log = config.Logger
					}

					// Log the panic
					logPanic(log, r, rec, stack, config.PrintStack)

					// Optional callback
					if config.OnPanic != nil {
						config.OnPanic(r, rec)
					}

					// Create AppError
					err := errors.Internal(nil, "An internal error occurred")
					_ = err.WithField("panic", fmt.Sprint(rec))

					if config.PrintStack {
						err.Stack = string(stack)
					}

					// Display error page
					if config.ErrorHandler != nil {
						config.ErrorHandler.Handle(w, r, err)
					} else {
						// Simple fallback
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// logPanic logs the panic to the logs.
func logPanic(log *logger.Logger, r *http.Request, rec any, stack []byte, printStack bool) {
	attrs := []any{
		"panic", fmt.Sprint(rec),
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
	}

	if requestID := RequestID(r.Context()); requestID != "" {
		attrs = append(attrs, "request_id", requestID)
	}

	if ua := r.Header.Get("User-Agent"); ua != "" {
		attrs = append(attrs, "user_agent", ua)
	}

	if printStack {
		attrs = append(attrs, "stack", string(stack))
	}

	log.Error("panic recovered", attrs...)
}

// SafeHandler wrapper that never panics.
func SafeHandler(h http.Handler, errorHandler *errors.Handler) http.Handler {
	return Recovery(errorHandler)(h)
}

// SafeHandlerFunc wrapper that never panics.
func SafeHandlerFunc(fn http.HandlerFunc, errorHandler *errors.Handler) http.Handler {
	return SafeHandler(fn, errorHandler)
}

// MustNotPanic executes a function and panics if it panics (for tests).
func MustNotPanic(fn func()) {
	defer func() {
		if rec := recover(); rec != nil {
			panic(fmt.Sprintf("unexpected panic: %v\n%s", rec, debug.Stack()))
		}
	}()
	fn()
}

// CatchPanic executes a function and returns the panic as an error.
func CatchPanic(fn func()) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("panic: %v", rec)
		}
	}()
	fn()
	return nil
}

// RecoverToError converts a panic to an error.
func RecoverToError(rec any) error {
	if rec == nil {
		return nil
	}

	if err, ok := rec.(error); ok {
		return err
	}

	return fmt.Errorf("panic: %v", rec)
}

// Deprecated: Use Recovery instead.
// PanicRecovery is an alias for Recovery (compatibility).
func PanicRecovery(errorHandler *errors.Handler) Middleware {
	return Recovery(errorHandler)
}
