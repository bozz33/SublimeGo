package validation

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structures

type User struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Age      int    `json:"age" validate:"required,gte=18"`
}

type Product struct {
	Name  string  `json:"name" validate:"required,min=3,max=100"`
	Price float64 `json:"price" validate:"required,gt=0"`
	Slug  string  `json:"slug" validate:"required,slug"`
}

type FrenchContact struct {
	Phone      string `json:"phone" validate:"required,phone_fr"`
	PostalCode string `json:"postal_code" validate:"required,postal_code_fr"`
}

type Company struct {
	SIRET string `json:"siret" validate:"required,siret"`
	SIREN string `json:"siren" validate:"required,siren"`
}

// Tests basiques

func TestNew(t *testing.T) {
	v := New()

	assert.NotNil(t, v)
	assert.NotNil(t, v.validate)
	assert.NotNil(t, v.messages)
}

func TestValidateStruct_Valid(t *testing.T) {
	user := User{
		Email:    "test@example.com",
		Password: "password123",
		Age:      25,
	}

	errors := ValidateStruct(user)
	assert.Nil(t, errors)
}

func TestValidateStruct_Invalid(t *testing.T) {
	user := User{
		Email:    "invalid-email",
		Password: "short",
		Age:      16,
	}

	errors := ValidateStruct(user)
	require.NotNil(t, errors)

	// Verify expected errors
	assert.Contains(t, errors, "email")
	assert.Contains(t, errors, "password")
	assert.Contains(t, errors, "age")

	// Verify French messages
	assert.Contains(t, errors["email"], "adresse email valide")
	assert.Contains(t, errors["password"], "au minimum 8 caractères")
	assert.Contains(t, errors["age"], "supérieur ou égal à 18")
}

func TestValidateStruct_Empty(t *testing.T) {
	user := User{}

	errors := ValidateStruct(user)
	require.NotNil(t, errors)

	assert.Contains(t, errors, "email")
	assert.Contains(t, errors, "password")
	assert.Contains(t, errors, "age")

	// Messages "obligatoire"
	assert.Contains(t, errors["email"], "obligatoire")
	assert.Contains(t, errors["password"], "obligatoire")
	assert.Contains(t, errors["age"], "obligatoire")
}

func TestValidateForm_Valid(t *testing.T) {
	form := url.Values{
		"email":    {"test@example.com"},
		"password": {"password123"},
		"age":      {"25"},
	}

	req := httptest.NewRequest("POST", "/test", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ParseForm()

	var user User
	errors := ValidateForm(req, &user)

	assert.Nil(t, errors)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "password123", user.Password)
	assert.Equal(t, 25, user.Age)
}

func TestValidateForm_Invalid(t *testing.T) {
	form := url.Values{
		"email":    {"invalid-email"},
		"password": {"short"},
		"age":      {"16"},
	}

	req := httptest.NewRequest("POST", "/test", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ParseForm()

	var user User
	errors := ValidateForm(req, &user)

	require.NotNil(t, errors)
	assert.Contains(t, errors, "email")
	assert.Contains(t, errors, "password")
	assert.Contains(t, errors, "age")
}

func TestValidateJSON_Valid(t *testing.T) {
	json := `{"email":"test@example.com","password":"password123","age":25}`

	req := httptest.NewRequest("POST", "/test", strings.NewReader(json))
	req.Header.Set("Content-Type", "application/json")

	var user User
	errors := ValidateJSON(req, &user)

	assert.Nil(t, errors)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "password123", user.Password)
	assert.Equal(t, 25, user.Age)
}

func TestValidateJSON_Invalid(t *testing.T) {
	json := `{"email":"invalid-email","password":"short","age":16}`

	req := httptest.NewRequest("POST", "/test", strings.NewReader(json))
	req.Header.Set("Content-Type", "application/json")

	var user User
	errors := ValidateJSON(req, &user)

	require.NotNil(t, errors)
	assert.Contains(t, errors, "email")
	assert.Contains(t, errors, "password")
	assert.Contains(t, errors, "age")
}

func TestValidateJSON_InvalidJSON(t *testing.T) {
	json := `{"email":"test@example.com","password":"password123","age":25` // Manque }

	req := httptest.NewRequest("POST", "/test", strings.NewReader(json))
	req.Header.Set("Content-Type", "application/json")

	var user User
	errors := ValidateJSON(req, &user)

	require.NotNil(t, errors)
	assert.Contains(t, errors, "json")
	assert.Contains(t, errors["json"], "Invalid JSON format")
}

// Tests helpers

func TestHasErrors(t *testing.T) {
	assert.False(t, HasErrors(nil))
	assert.False(t, HasErrors(map[string]string{}))
	assert.True(t, HasErrors(map[string]string{"field": "error"}))
}

func TestCountErrors(t *testing.T) {
	assert.Equal(t, 0, CountErrors(nil))
	assert.Equal(t, 0, CountErrors(map[string]string{}))
	assert.Equal(t, 1, CountErrors(map[string]string{"field": "error"}))
	assert.Equal(t, 2, CountErrors(map[string]string{"field1": "error1", "field2": "error2"}))
}

func TestGetError(t *testing.T) {
	errors := map[string]string{
		"email":    "Email invalide",
		"password": "Mot de passe trop court",
	}

	assert.Equal(t, "Email invalide", GetError(errors, "email"))
	assert.Equal(t, "Mot de passe trop court", GetError(errors, "password"))
	assert.Equal(t, "", GetError(errors, "name"))
}

func TestFirstError(t *testing.T) {
	errors := map[string]string{
		"email":    "Email invalide",
		"password": "Mot de passe trop court",
	}

	// Order is not guaranteed, but we should get one of the errors
	first := FirstError(errors)
	assert.True(t, first == "Email invalide" || first == "Mot de passe trop court")

	assert.Equal(t, "", FirstError(nil))
	assert.Equal(t, "", FirstError(map[string]string{}))
}

func TestAllErrors(t *testing.T) {
	errors := map[string]string{
		"email":    "Email invalide",
		"password": "Mot de passe trop court",
	}

	all := AllErrors(errors)
	assert.Len(t, all, 2)
	assert.Contains(t, all, "Email invalide")
	assert.Contains(t, all, "Mot de passe trop court")
}

func TestErrorsAsString(t *testing.T) {
	errors := map[string]string{
		"email":    "Email invalide",
		"password": "Mot de passe trop court",
	}

	result := ErrorsAsString(errors, ", ")
	assert.Contains(t, result, "Email invalide")
	assert.Contains(t, result, "Mot de passe trop court")
	assert.Contains(t, result, ", ")
}

func TestMergeErrors(t *testing.T) {
	errors1 := map[string]string{
		"email": "Email invalide",
		"name":  "Nom requis",
	}

	errors2 := map[string]string{
		"password": "Mot de passe trop court",
		"email":    "Email déjà utilisé", // Should overwrite
	}

	merged := MergeErrors(errors1, errors2)

	assert.Len(t, merged, 3)
	assert.Equal(t, "Email déjà utilisé", merged["email"]) // Overwritten
	assert.Equal(t, "Nom requis", merged["name"])
	assert.Equal(t, "Mot de passe trop court", merged["password"])
}

func TestFilterErrors(t *testing.T) {
	errors := map[string]string{
		"email":    "Email invalide",
		"password": "Mot de passe trop court",
		"name":     "Nom requis",
	}

	filtered := FilterErrors(errors, "email", "password")

	assert.Len(t, filtered, 2)
	assert.Contains(t, filtered, "email")
	assert.Contains(t, filtered, "password")
	assert.NotContains(t, filtered, "name")
}

func TestOnlyErrors(t *testing.T) {
	errors := map[string]string{
		"email":    "Email invalide",
		"password": "Mot de passe trop court",
	}

	fields := OnlyErrors(errors)
	assert.Len(t, fields, 2)
	assert.Contains(t, fields, "email")
	assert.Contains(t, fields, "password")
}

// Tests validators custom

func TestValidatePhoneFR_Valid(t *testing.T) {
	validPhones := []string{
		"0612345678",
		"06 12 34 56 78",
		"06.12.34.56.78",
		"06-12-34-56-78",
		"+33612345678",
		"+33 6 12 34 56 78",
		"0712345678",
		"0123456789",
	}

	for _, phone := range validPhones {
		t.Run(phone, func(t *testing.T) {
			assert.True(t, IsValidPhoneFR(phone), "Phone should be valid: %s", phone)
		})
	}
}

func TestValidatePhoneFR_Invalid(t *testing.T) {
	invalidPhones := []string{
		"12345678",    // Too short
		"06123456789", // Too long
		"abcd1234",    // Letters
		"",            // Empty
	}

	for _, phone := range invalidPhones {
		t.Run(phone, func(t *testing.T) {
			assert.False(t, IsValidPhoneFR(phone), "Phone should be invalid: %s", phone)
		})
	}
}

func TestValidatePostalCodeFR_Valid(t *testing.T) {
	validCodes := []string{
		"75001",
		"13000",
		"69001",
		"2A000",
		"2B000",
		"97100", // DOM-TOM
	}

	for _, code := range validCodes {
		t.Run(code, func(t *testing.T) {
			assert.True(t, IsValidPostalCodeFR(code), "Postal code should be valid: %s", code)
		})
	}
}

func TestValidatePostalCodeFR_Invalid(t *testing.T) {
	invalidCodes := []string{
		"7500",   // Too short
		"750012", // Too long
		"75A01",  // Invalid format
		"ABCDE",  // Letters
		"",       // Empty
	}

	for _, code := range invalidCodes {
		t.Run(code, func(t *testing.T) {
			assert.False(t, IsValidPostalCodeFR(code), "Postal code should be invalid: %s", code)
		})
	}
}

func TestValidateSlug_Valid(t *testing.T) {
	validSlugs := []string{
		"mon-article",
		"article-123",
		"mon_article_123",
		"test",
		"a",
		"123",
	}

	for _, slug := range validSlugs {
		t.Run(slug, func(t *testing.T) {
			assert.True(t, IsValidSlug(slug), "Slug should be valid: %s", slug)
		})
	}
}

func TestValidateSlug_Invalid(t *testing.T) {
	invalidSlugs := []string{
		"-article",    // Starts with -
		"article-",    // Ends with -
		"_article",    // Starts with _
		"article_",    // Ends with _
		"Mon Article", // Spaces
		"",            // Empty
		"article@123", // Special character
	}

	for _, slug := range invalidSlugs {
		t.Run(slug, func(t *testing.T) {
			assert.False(t, IsValidSlug(slug), "Slug should be invalid: %s", slug)
		})
	}
}

func TestValidateSIRET_Valid(t *testing.T) {
	validSIRETs := []string{
		"73282932000074", // Real example with valid Luhn
	}

	for _, siret := range validSIRETs {
		t.Run(siret, func(t *testing.T) {
			assert.True(t, IsValidSIRET(siret), "SIRET should be valid: %s", siret)
		})
	}
}

func TestValidateSIRET_Invalid(t *testing.T) {
	invalidSIRETs := []string{
		"73282932000075",  // Same format but invalid Luhn
		"1234567890123",   // Too short
		"123456789012345", // Too long
		"ABCDEFGHIJKLMN",  // Letters
		"",                // Empty
	}

	for _, siret := range invalidSIRETs {
		t.Run(siret, func(t *testing.T) {
			assert.False(t, IsValidSIRET(siret), "SIRET should be invalid: %s", siret)
		})
	}
}

func TestValidateSIREN_Valid(t *testing.T) {
	validSIRENs := []string{
		"732829320", // Real example with valid Luhn
	}

	for _, siren := range validSIRENs {
		t.Run(siren, func(t *testing.T) {
			assert.True(t, IsValidSIREN(siren), "SIREN should be valid: %s", siren)
		})
	}
}

func TestValidateSIREN_Invalid(t *testing.T) {
	invalidSIRENs := []string{
		"732829321",  // Same format but invalid Luhn
		"12345678",   // Too short
		"1234567890", // Too long
		"ABCDEFGHI",  // Letters
		"",           // Empty
	}

	for _, siren := range invalidSIRENs {
		t.Run(siren, func(t *testing.T) {
			assert.False(t, IsValidSIREN(siren), "SIREN should be invalid: %s", siren)
		})
	}
}

func TestValidateStrongPassword_Valid(t *testing.T) {
	validPasswords := []string{
		"Password123",
		"StrongPass1",
		"MySecurePwd9",
		"Abcdefgh1",
	}

	for _, password := range validPasswords {
		t.Run(password, func(t *testing.T) {
			assert.True(t, IsStrongPassword(password), "Password should be valid: %s", password)
		})
	}
}

func TestValidateStrongPassword_Invalid(t *testing.T) {
	invalidPasswords := []string{
		"short",         // Too short
		"alllowercase1", // No uppercase
		"ALLUPPERCASE1", // No lowercase
		"NoNumbersHere", // No digit
		"Short1",        // Too short despite other criteria
		"",              // Empty
	}

	for _, password := range invalidPasswords {
		t.Run(password, func(t *testing.T) {
			assert.False(t, IsStrongPassword(password), "Password should be invalid: %s", password)
		})
	}
}

// Integration tests

func TestFrenchContact_Validation(t *testing.T) {
	contact := FrenchContact{
		Phone:      "06 12 34 56 78",
		PostalCode: "75001",
	}

	errors := ValidateStruct(contact)
	assert.Nil(t, errors)
}

func TestFrenchContact_Invalid(t *testing.T) {
	contact := FrenchContact{
		Phone:      "invalid-phone",
		PostalCode: "invalid-postal",
	}

	errors := ValidateStruct(contact)
	require.NotNil(t, errors)

	assert.Contains(t, errors, "phone")
	assert.Contains(t, errors, "postal_code")
	assert.Contains(t, errors["phone"], "téléphone français valide")
	assert.Contains(t, errors["postal_code"], "code postal français valide")
}

func TestCompany_Validation(t *testing.T) {
	company := Company{
		SIRET: "73282932000074",
		SIREN: "732829320",
	}

	errors := ValidateStruct(company)
	assert.Nil(t, errors)
}

func TestCompany_Invalid(t *testing.T) {
	company := Company{
		SIRET: "invalid-siret",
		SIREN: "invalid-siren",
	}

	errors := ValidateStruct(company)
	require.NotNil(t, errors)

	assert.Contains(t, errors, "siret")
	assert.Contains(t, errors, "siren")
	assert.Contains(t, errors["siret"], "numéro SIRET valide")
	assert.Contains(t, errors["siren"], "numéro SIREN valide")
}

// Tests helpers globaux

func TestCheck(t *testing.T) {
	validUser := User{
		Email:    "test@example.com",
		Password: "password123",
		Age:      25,
	}

	invalidUser := User{
		Email:    "invalid-email",
		Password: "short",
		Age:      16,
	}

	assert.True(t, Check(validUser))
	assert.False(t, Check(invalidUser))
}

func TestMust(t *testing.T) {
	validUser := User{
		Email:    "test@example.com",
		Password: "password123",
		Age:      25,
	}

	// Ne doit pas panic
	assert.NotPanics(t, func() {
		Must(validUser)
	})

	// Doit panic avec utilisateur invalide
	assert.Panics(t, func() {
		invalidUser := User{
			Email:    "invalid-email",
			Password: "short",
			Age:      16,
		}
		Must(invalidUser)
	})
}

// Benchmarks

func BenchmarkValidateStruct_Valid(b *testing.B) {
	user := User{
		Email:    "test@example.com",
		Password: "password123",
		Age:      25,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateStruct(user)
	}
}

func BenchmarkValidateStruct_Invalid(b *testing.B) {
	user := User{
		Email:    "invalid-email",
		Password: "short",
		Age:      16,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateStruct(user)
	}
}

func BenchmarkIsValidPhoneFR(b *testing.B) {
	phone := "06 12 34 56 78"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsValidPhoneFR(phone)
	}
}

func BenchmarkIsValidSIRET(b *testing.B) {
	siret := "73282932000074"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsValidSIRET(siret)
	}
}
