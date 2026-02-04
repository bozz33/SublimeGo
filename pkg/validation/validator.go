package validation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"
)

// Validator wraps the go-playground validator.
type Validator struct {
	validate *validator.Validate
	messages map[string]string
}

// New creates a new validator.
func New() *Validator {
	v := &Validator{
		validate: validator.New(),
		messages: defaultMessages(),
	}

	v.validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		if name == "" {
			return fld.Name
		}
		return name
	})

	v.registerCustomValidators()

	return v
}

// Validate validates a struct.
func (v *Validator) Validate(s interface{}) error {
	return v.validate.Struct(s)
}

// ValidateVar validates a single variable.
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}

// ValidateStruct validates a struct and returns formatted errors.
func ValidateStruct(s interface{}) map[string]string {
	v := New()
	err := v.Validate(s)
	if err == nil {
		return nil
	}

	return formatErrors(err, v.messages)
}

// ValidateForm validates an HTTP form and binds to a struct.
func ValidateForm(r *http.Request, dest interface{}) map[string]string {
	if err := r.ParseForm(); err != nil {
		return map[string]string{"form": "Failed to parse form"}
	}

	if err := decoder.Decode(dest, r.Form); err != nil {
		return map[string]string{"form": "Failed to bind data"}
	}

	return ValidateStruct(dest)
}

// ValidateJSON validates JSON and binds to a struct.
func ValidateJSON(r *http.Request, dest interface{}) map[string]string {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		return map[string]string{"json": "Invalid JSON format"}
	}
	return ValidateStruct(dest)
}

// Check quickly checks if a struct is valid.
func Check(s interface{}) bool {
	return ValidateStruct(s) == nil
}

// Must validates a struct and panics on error (useful for tests).
func Must(s interface{}) {
	if errors := ValidateStruct(s); errors != nil {
		panic(fmt.Sprintf("Validation failed: %v", errors))
	}
}

// formatErrors formats validation errors.
func formatErrors(err error, messages map[string]string) map[string]string {
	result := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()
			param := e.Param()

			message, exists := messages[tag]
			if !exists {
				message = fmt.Sprintf("Field %s is invalid", field)
			}

			message = strings.ReplaceAll(message, "{field}", field)
			message = strings.ReplaceAll(message, "{param}", param)
			message = strings.ReplaceAll(message, "{value}", fmt.Sprintf("%v", e.Value()))

			result[field] = message
		}
	}

	return result
}

// HasErrors checks if there are any errors.
func HasErrors(errors map[string]string) bool {
	return len(errors) > 0
}

// CountErrors returns the number of errors.
func CountErrors(errors map[string]string) int {
	return len(errors)
}

// GetError retrieves the error for a specific field.
func GetError(errors map[string]string, field string) string {
	return errors[field]
}

// FirstError returns the first error.
func FirstError(errors map[string]string) string {
	for _, msg := range errors {
		return msg
	}
	return ""
}

// AllErrors returns all errors as a slice.
func AllErrors(errors map[string]string) []string {
	return lo.Values(errors)
}

// ErrorsAsString concatenates errors into a string.
func ErrorsAsString(errors map[string]string, separator string) string {
	return strings.Join(AllErrors(errors), separator)
}

// MergeErrors merges multiple error maps.
func MergeErrors(errorMaps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, errMap := range errorMaps {
		for k, v := range errMap {
			result[k] = v
		}
	}
	return result
}

// FilterErrors filters errors for specific fields.
func FilterErrors(errors map[string]string, fields ...string) map[string]string {
	return lo.PickBy(errors, func(key string, value string) bool {
		return lo.Contains(fields, key)
	})
}

// OnlyErrors returns only the field names with errors.
func OnlyErrors(errors map[string]string) []string {
	return lo.Keys(errors)
}

// RegisterCustomMessage registers a custom message.
func RegisterCustomMessage(tag, message string) {
	customMessages[tag] = message
}

// RegisterValidation registers a custom validator.
func RegisterValidation(tag string, fn validator.Func) {
	validator.New().RegisterValidation(tag, fn)
}

// Global variables for custom messages.
var (
	customMessages = make(map[string]string)
	decoder        = NewFormDecoder()
)

// NewFormDecoder creates a form decoder.
func NewFormDecoder() *FormDecoder {
	return &FormDecoder{}
}

// FormDecoder decodes form data to a struct.
type FormDecoder struct{}

// Decode decodes form data to a struct.
func (d *FormDecoder) Decode(dest interface{}, data map[string][]string) error {
	val := reflect.ValueOf(dest).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		jsonTag := fieldType.Tag.Get("json")
		fieldName := strings.SplitN(jsonTag, ",", 2)[0]
		if fieldName == "" || fieldName == "-" {
			fieldName = fieldType.Name
		}

		if values, exists := data[fieldName]; exists && len(values) > 0 {
			value := values[0]

			// Assign value based on type
			if field.CanSet() {
				switch field.Kind() {
				case reflect.String:
					field.SetString(value)
				case reflect.Int:
					if intVal, err := parseInt(value); err == nil {
						field.SetInt(int64(intVal))
					}
				case reflect.Float64:
					if floatVal, err := parseFloat(value); err == nil {
						field.SetFloat(floatVal)
					}
				}
			}
		}
	}

	return nil
}

// parseInt converts string to int
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// parseFloat converts string to float64
func parseFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}
