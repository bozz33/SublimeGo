// Package validation provides data validation using go-playground/validator.
//
// It wraps the validator library with English error messages by default and adds
// custom validators for specific data types (phone, postal code, SIRET/SIREN).
// The package supports struct validation, form validation, and JSON validation.
//
// Features:
//   - Struct validation with tags
//   - English error messages (French messages available)
//   - Custom validators (phone_fr, postal_code_fr, siret, siren, slug)
//   - Strong password validation
//   - Form and JSON validation helpers
//   - Error message helpers
//
// Basic usage:
//
//	type User struct {
//		Email    string `validate:"required,email"`
//		Password string `validate:"required,min=8"`
//		Phone    string `validate:"phone_fr"`
//	}
//
//	errors := validation.ValidateStruct(user)
//	if errors != nil {
//		// Handle validation errors
//		for field, msg := range errors {
//			fmt.Printf("%s: %s\n", field, msg)
//		}
//	}
//
//	// Validate JSON request
//	errors := validation.ValidateJSON(request, &user)
package validation
