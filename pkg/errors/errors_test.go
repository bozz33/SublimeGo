package errors

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	err := New("TEST_ERROR", "Test message", http.StatusBadRequest)

	require.NotNil(t, err)
	assert.Equal(t, "TEST_ERROR", err.Code)
	assert.Equal(t, "Test message", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
	assert.NotNil(t, err.Fields)
}

func TestWrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	err := Wrap(originalErr, "WRAP_ERROR", "Wrapped message", http.StatusInternalServerError)

	require.NotNil(t, err)
	assert.Equal(t, "WRAP_ERROR", err.Code)
	assert.Equal(t, "Wrapped message", err.Message)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode)
	assert.Equal(t, originalErr, err.Err)
	assert.NotEmpty(t, err.Stack)
}

func TestAppErrorError(t *testing.T) {
	// Without original error
	err := New("TEST", "Test message", 400)
	assert.Equal(t, "TEST: Test message", err.Error())

	// With original error
	originalErr := fmt.Errorf("db error")
	err = Wrap(originalErr, "TEST", "Test message", 500)
	assert.Contains(t, err.Error(), "TEST: Test message")
	assert.Contains(t, err.Error(), "db error")
}

func TestAppErrorUnwrap(t *testing.T) {
	originalErr := fmt.Errorf("original")
	err := Wrap(originalErr, "CODE", "message", 500)

	unwrapped := err.Unwrap()
	assert.Equal(t, originalErr, unwrapped)
}

func TestWithField(t *testing.T) {
	err := New("TEST", "message", 400)
	err.WithField("key", "value")

	assert.Equal(t, "value", err.Fields["key"])
}

func TestWithFields(t *testing.T) {
	err := New("TEST", "message", 400)

	fields := map[string]any{
		"field1": "value1",
		"field2": 123,
	}

	err.WithFields(fields)

	assert.Equal(t, "value1", err.Fields["field1"])
	assert.Equal(t, 123, err.Fields["field2"])
}

func TestWithStack(t *testing.T) {
	err := New("TEST", "message", 500)
	err.WithStack()

	assert.NotEmpty(t, err.Stack)
	assert.Contains(t, err.Stack, "goroutine")
}

func TestNotFound(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		wantCode int
	}{
		{
			name:     "with message",
			message:  "User not found",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "empty message",
			message:  "",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NotFound(tt.message)

			assert.Equal(t, "NOT_FOUND", err.Code)
			assert.Equal(t, tt.wantCode, err.StatusCode)
			assert.NotEmpty(t, err.Message)
		})
	}
}

func TestNotFoundf(t *testing.T) {
	err := NotFoundf("User %d not found", 123)

	assert.Equal(t, "NOT_FOUND", err.Code)
	assert.Contains(t, err.Message, "User 123 not found")
	assert.Equal(t, http.StatusNotFound, err.StatusCode)
}

func TestBadRequest(t *testing.T) {
	err := BadRequest("Invalid input")

	assert.Equal(t, "BAD_REQUEST", err.Code)
	assert.Equal(t, "Invalid input", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
}

func TestUnauthorized(t *testing.T) {
	err := Unauthorized("")

	assert.Equal(t, "UNAUTHORIZED", err.Code)
	assert.NotEmpty(t, err.Message)
	assert.Equal(t, http.StatusUnauthorized, err.StatusCode)
}

func TestForbidden(t *testing.T) {
	err := Forbidden("Not allowed")

	assert.Equal(t, "FORBIDDEN", err.Code)
	assert.Equal(t, "Not allowed", err.Message)
	assert.Equal(t, http.StatusForbidden, err.StatusCode)
}

func TestConflict(t *testing.T) {
	err := Conflict("Email already exists")

	assert.Equal(t, "CONFLICT", err.Code)
	assert.Equal(t, "Email already exists", err.Message)
	assert.Equal(t, http.StatusConflict, err.StatusCode)
}

func TestValidationError(t *testing.T) {
	fields := map[string]string{
		"email":    "Email invalide",
		"password": "Mot de passe trop court",
	}

	err := ValidationError(fields)

	assert.Equal(t, "VALIDATION_ERROR", err.Code)
	assert.Equal(t, http.StatusUnprocessableEntity, err.StatusCode)
	assert.Len(t, err.Fields, 2)
	assert.Equal(t, "Email invalide", err.Fields["email"])
	assert.Equal(t, "Mot de passe trop court", err.Fields["password"])
}

func TestInternal(t *testing.T) {
	originalErr := fmt.Errorf("db connection failed")
	err := Internal(originalErr, "Database error")

	assert.Equal(t, "INTERNAL_ERROR", err.Code)
	assert.Equal(t, "Database error", err.Message)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode)
	assert.Equal(t, originalErr, err.Err)
	assert.NotEmpty(t, err.Stack)
}

func TestInternalf(t *testing.T) {
	originalErr := fmt.Errorf("timeout")
	err := Internalf(originalErr, "Failed after %d attempts", 3)

	assert.Equal(t, "INTERNAL_ERROR", err.Code)
	assert.Contains(t, err.Message, "Failed after 3 attempts")
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode)
}

func TestServiceUnavailable(t *testing.T) {
	err := ServiceUnavailable("Maintenance")

	assert.Equal(t, "SERVICE_UNAVAILABLE", err.Code)
	assert.Equal(t, "Maintenance", err.Message)
	assert.Equal(t, http.StatusServiceUnavailable, err.StatusCode)
}

func TestToAppError(t *testing.T) {
	// nil error
	appErr := ToAppError(nil)
	assert.Nil(t, appErr)

	// Already an AppError
	original := NotFound("test")
	appErr = ToAppError(original)
	assert.Equal(t, original, appErr)

	// Standard error
	standardErr := fmt.Errorf("standard error")
	appErr = ToAppError(standardErr)
	assert.NotNil(t, appErr)
	assert.Equal(t, "INTERNAL_ERROR", appErr.Code)
	assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
}

func TestIsAppError(t *testing.T) {
	appErr := NotFound("test")
	assert.True(t, IsAppError(appErr))

	standardErr := fmt.Errorf("standard")
	assert.False(t, IsAppError(standardErr))
}

func TestHasCode(t *testing.T) {
	err := NotFound("test")

	assert.True(t, HasCode(err, "NOT_FOUND"))
	assert.False(t, HasCode(err, "INTERNAL_ERROR"))

	standardErr := fmt.Errorf("standard")
	assert.False(t, HasCode(standardErr, "NOT_FOUND"))
}

func TestIsNotFound(t *testing.T) {
	err := NotFound("test")
	assert.True(t, IsNotFound(err))

	err = BadRequest("test")
	assert.False(t, IsNotFound(err))
}

func TestIsValidation(t *testing.T) {
	err := ValidationError(map[string]string{"email": "invalid"})
	assert.True(t, IsValidation(err))

	err = NotFound("test")
	assert.False(t, IsValidation(err))
}

func TestGetValidationErrors(t *testing.T) {
	fields := map[string]string{
		"email":    "Invalid email",
		"password": "Too short",
	}

	err := ValidationError(fields)
	validationErrs := GetValidationErrors(err)

	assert.NotNil(t, validationErrs)
	assert.Equal(t, "Invalid email", validationErrs["email"])
	assert.Equal(t, "Too short", validationErrs["password"])

	// Non-validation error
	err = NotFound("test")
	validationErrs = GetValidationErrors(err)
	assert.Nil(t, validationErrs)
}

func TestErrorList(t *testing.T) {
	list := NewErrorList()

	assert.False(t, list.HasErrors())
	assert.Nil(t, list.First())

	err1 := NotFound("error 1")
	err2 := BadRequest("error 2")

	list.Add(err1)
	list.Add(err2)

	assert.True(t, list.HasErrors())
	assert.Len(t, list.Errors, 2)
	assert.Equal(t, err1, list.First())
}

func TestErrorListError(t *testing.T) {
	list := NewErrorList()
	list.Add(NotFound("error 1"))
	list.Add(BadRequest("error 2"))

	errMsg := list.Error()
	assert.Contains(t, errMsg, "multiple errors")
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New("TEST", "message", 400)
	}
}

func BenchmarkWrap(b *testing.B) {
	err := fmt.Errorf("original")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Wrap(err, "TEST", "message", 500)
	}
}

func BenchmarkToAppError(b *testing.B) {
	standardErr := fmt.Errorf("test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ToAppError(standardErr)
	}
}
