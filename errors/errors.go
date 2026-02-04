package errors

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/samber/lo"
)

// AppError represents a structured application error.
type AppError struct {
	Code       string
	Message    string
	StatusCode int
	Err        error
	Stack      string
	Fields     map[string]any
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithField adds an additional field.
func (e *AppError) WithField(key string, value any) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]any)
	}
	e.Fields[key] = value
	return e
}

// WithFields adds multiple fields.
func (e *AppError) WithFields(fields map[string]any) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]any)
	}
	for k, v := range fields {
		e.Fields[k] = v
	}
	return e
}

// WithStack captures the current stack trace.
func (e *AppError) WithStack() *AppError {
	e.Stack = string(debug.Stack())
	return e
}

// New creates a new AppError.
func New(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Fields:     make(map[string]any),
	}
}

// Wrap wraps an existing error.
func Wrap(err error, code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
		Stack:      string(debug.Stack()),
		Fields:     make(map[string]any),
	}
}

// NotFound creates a 404 error.
func NotFound(message string) *AppError {
	if message == "" {
		message = "Resource not found"
	}
	return New("NOT_FOUND", message, http.StatusNotFound)
}

// NotFoundf creates a formatted 404 error.
func NotFoundf(format string, args ...any) *AppError {
	return NotFound(fmt.Sprintf(format, args...))
}

// BadRequest creates a 400 error.
func BadRequest(message string) *AppError {
	if message == "" {
		message = "Invalid request"
	}
	return New("BAD_REQUEST", message, http.StatusBadRequest)
}

// BadRequestf creates a formatted 400 error.
func BadRequestf(format string, args ...any) *AppError {
	return BadRequest(fmt.Sprintf(format, args...))
}

// Unauthorized creates a 401 error.
func Unauthorized(message string) *AppError {
	if message == "" {
		message = "Authentication required"
	}
	return New("UNAUTHORIZED", message, http.StatusUnauthorized)
}

// Forbidden creates a 403 error.
func Forbidden(message string) *AppError {
	if message == "" {
		message = "Access denied"
	}
	return New("FORBIDDEN", message, http.StatusForbidden)
}

// Conflict creates a 409 error.
func Conflict(message string) *AppError {
	if message == "" {
		message = "A conflict occurred"
	}
	return New("CONFLICT", message, http.StatusConflict)
}

// ValidationError creates a validation error.
func ValidationError(fields map[string]string) *AppError {
	err := New("VALIDATION_ERROR", "Validation failed", http.StatusUnprocessableEntity)
	err.Fields = lo.MapEntries(fields, func(k string, v string) (string, any) {
		return k, v
	})
	return err
}

// Internal creates a 500 error.
func Internal(err error, message string) *AppError {
	if message == "" {
		message = "An internal error occurred"
	}

	appErr := Wrap(err, "INTERNAL_ERROR", message, http.StatusInternalServerError)
	appErr.Stack = string(debug.Stack())
	return appErr
}

// Internalf creates a formatted 500 error.
func Internalf(err error, format string, args ...any) *AppError {
	return Internal(err, fmt.Sprintf(format, args...))
}

// ServiceUnavailable creates a 503 error.
func ServiceUnavailable(message string) *AppError {
	if message == "" {
		message = "Service temporarily unavailable"
	}
	return New("SERVICE_UNAVAILABLE", message, http.StatusServiceUnavailable)
}

// ToAppError converts a standard error to AppError.
func ToAppError(err error) *AppError {
	if err == nil {
		return nil
	}
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return Internal(err, "An error occurred")
}

// IsAppError checks if an error is an AppError.
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// HasCode checks if the error has a specific code.
func HasCode(err error, code string) bool {
	if appErr := ToAppError(err); appErr != nil {
		return appErr.Code == code
	}
	return false
}

// IsNotFound checks if it's a 404 error.
func IsNotFound(err error) bool {
	if appErr := ToAppError(err); appErr != nil {
		return appErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsValidation checks if it's a validation error.
func IsValidation(err error) bool {
	return HasCode(err, "VALIDATION_ERROR")
}

// GetValidationErrors extracts validation errors from an AppError.
func GetValidationErrors(err error) map[string]string {
	appErr := ToAppError(err)
	if appErr == nil || !IsValidation(err) {
		return nil
	}
	return lo.MapValues(appErr.Fields, func(v any, _ string) string {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprint(v)
	})
}

// ErrorList represents multiple errors.
type ErrorList struct {
	Errors []*AppError
}

// Error implements the error interface.
func (e *ErrorList) Error() string {
	messages := lo.Map(e.Errors, func(err *AppError, _ int) string {
		return err.Message
	})
	return fmt.Sprintf("multiple errors: %v", messages)
}

// Add adds an error to the list.
func (e *ErrorList) Add(err *AppError) {
	e.Errors = append(e.Errors, err)
}

// HasErrors checks if there are any errors.
func (e *ErrorList) HasErrors() bool {
	return len(e.Errors) > 0
}

// First returns the first error.
func (e *ErrorList) First() *AppError {
	if len(e.Errors) == 0 {
		return nil
	}
	return e.Errors[0]
}

// NewErrorList creates a new error list.
func NewErrorList() *ErrorList {
	return &ErrorList{
		Errors: make([]*AppError, 0),
	}
}
