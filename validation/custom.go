package validation

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Package-level compiled regexes — compiled once, zero alloc on hot path.
var (
	rePostalCodeFR = regexp.MustCompile(`^\d{5}$`)
	rePostalCorsFR = regexp.MustCompile(`^2[AB]\d{3}$`)
	reSlug         = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	reSIRET        = regexp.MustCompile(`^\d{14}$`)
	reSIREN        = regexp.MustCompile(`^\d{9}$`)
	reDigit        = regexp.MustCompile(`\d`)
	reUpper        = regexp.MustCompile(`[A-Z]`)
	reLower        = regexp.MustCompile(`[a-z]`)
)

// registerCustomValidators registers all custom validators.
func (v *Validator) registerCustomValidators() {
	_ = v.validate.RegisterValidation("phone_fr", validatePhoneFR)
	_ = v.validate.RegisterValidation("postal_code_fr", validatePostalCodeFR)
	_ = v.validate.RegisterValidation("slug", validateSlug)
	_ = v.validate.RegisterValidation("siret", validateSIRET)
	_ = v.validate.RegisterValidation("siren", validateSIREN)
	_ = v.validate.RegisterValidation("strong_password", validateStrongPassword)
}

// validatePhoneFR validates a French phone number.
// Accepted formats:
// - 0612345678
// - 06 12 34 56 78
// - 06.12.34.56.78
// - 06-12-34-56-78
// - +33612345678
// - +33 6 12 34 56 78
func validatePhoneFR(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, ".", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")

	if strings.HasPrefix(phone, "+33") {
		phone = phone[3:]
		if len(phone) != 9 {
			return false
		}
		// Must start with 6, 7 or 0 (for mobile)
		return strings.HasPrefix(phone, "6") || strings.HasPrefix(phone, "7") || strings.HasPrefix(phone, "0")
	}

	if len(phone) == 10 {
		return strings.HasPrefix(phone, "0")
	}

	if len(phone) == 9 {
		return strings.HasPrefix(phone, "6") || strings.HasPrefix(phone, "7")
	}

	return false
}

// validatePostalCodeFR validates a French postal code.
// Accepted formats: 75001, 13000, 69001, 2A000, 2B000
func validatePostalCodeFR(fl validator.FieldLevel) bool {
	postalCode := fl.Field().String()

	if len(postalCode) != 5 {
		return false
	}

	if rePostalCodeFR.MatchString(postalCode) {
		return true
	}
	return rePostalCorsFR.MatchString(postalCode)
}

// validateSlug validates a slug (URL-friendly).
// Accepts: mon-article, article-123, mon_article_123
// Rejects: Mon Article, -article, article-
func validateSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()

	if len(slug) == 0 {
		return false
	}

	if strings.HasPrefix(slug, "-") || strings.HasPrefix(slug, "_") ||
		strings.HasSuffix(slug, "-") || strings.HasSuffix(slug, "_") {
		return false
	}

	return reSlug.MatchString(slug)
}

// validateSIRET validates a SIRET number (14 digits).
// Uses the Luhn algorithm for validation.
func validateSIRET(fl validator.FieldLevel) bool {
	siret := fl.Field().String()

	siret = strings.ReplaceAll(siret, " ", "")
	siret = strings.ReplaceAll(siret, ".", "")

	if len(siret) != 14 {
		return false
	}

	if !reSIRET.MatchString(siret) {
		return false
	}

	sum := 0
	parity := len(siret) % 2

	for i, char := range siret {
		digit := int(char - '0')

		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
	}

	return sum%10 == 0
}

// validateSIREN validates a SIREN number (9 digits).
// Uses the Luhn algorithm for validation.
func validateSIREN(fl validator.FieldLevel) bool {
	siren := fl.Field().String()

	siren = strings.ReplaceAll(siren, " ", "")
	siren = strings.ReplaceAll(siren, ".", "")

	if len(siren) != 9 {
		return false
	}

	if !reSIREN.MatchString(siren) {
		return false
	}

	sum := 0
	parity := len(siren) % 2

	for i, char := range siren {
		digit := int(char - '0')

		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
	}

	return sum%10 == 0
}

// validateStrongPassword validates a strong password.
// Rules: 8+ characters, 1 uppercase, 1 lowercase, 1 digit
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	if !reUpper.MatchString(password) {
		return false
	}
	if !reLower.MatchString(password) {
		return false
	}
	return reDigit.MatchString(password)
}

// Standalone helpers for quick validation

// IsValidPhoneFR checks if a French phone number is valid.
func IsValidPhoneFR(phone string) bool {
	v := New()
	return v.validate.Var(phone, "phone_fr") == nil
}

// IsValidPostalCodeFR checks if a French postal code is valid.
func IsValidPostalCodeFR(postalCode string) bool {
	v := New()
	return v.validate.Var(postalCode, "postal_code_fr") == nil
}

// IsValidSlug checks if a slug is valid.
func IsValidSlug(slug string) bool {
	v := New()
	return v.validate.Var(slug, "slug") == nil
}

// IsValidSIRET checks if a SIRET is valid.
func IsValidSIRET(siret string) bool {
	v := New()
	return v.validate.Var(siret, "siret") == nil
}

// IsValidSIREN checks if a SIREN is valid.
func IsValidSIREN(siren string) bool {
	v := New()
	return v.validate.Var(siren, "siren") == nil
}

// IsStrongPassword checks if a password is strong.
func IsStrongPassword(password string) bool {
	v := New()
	return v.validate.Var(password, "strong_password") == nil
}
