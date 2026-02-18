package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Rule represents a validation rule that can be applied to a field.
type Rule interface {
	// Validate validates a value and returns an error message if invalid.
	Validate(value any) string
	// GetName returns the rule name.
	GetName() string
}

// RuleSet is a collection of rules for a field.
type RuleSet struct {
	FieldName string
	Rules     []Rule
}

// NewRuleSet creates a new rule set for a field.
func NewRuleSet(fieldName string) *RuleSet {
	return &RuleSet{
		FieldName: fieldName,
		Rules:     make([]Rule, 0),
	}
}

// Add adds a rule to the set.
func (rs *RuleSet) Add(rule Rule) *RuleSet {
	rs.Rules = append(rs.Rules, rule)
	return rs
}

// Required adds a required rule.
func (rs *RuleSet) Required() *RuleSet {
	return rs.Add(&RequiredRule{})
}

// Email adds an email rule.
func (rs *RuleSet) Email() *RuleSet {
	return rs.Add(&EmailRule{})
}

// Min adds a minimum length/value rule.
func (rs *RuleSet) Min(min int) *RuleSet {
	return rs.Add(&MinRule{Min: min})
}

// Max adds a maximum length/value rule.
func (rs *RuleSet) Max(max int) *RuleSet {
	return rs.Add(&MaxRule{Max: max})
}

// Between adds a between rule.
func (rs *RuleSet) Between(min, max int) *RuleSet {
	return rs.Add(&BetweenRule{Min: min, Max: max})
}

// Regex adds a regex pattern rule.
func (rs *RuleSet) Regex(pattern string) *RuleSet {
	return rs.Add(&RegexRule{Pattern: pattern})
}

// In adds an "in list" rule.
func (rs *RuleSet) In(values ...string) *RuleSet {
	return rs.Add(&InRule{Values: values})
}

// URL adds a URL validation rule.
func (rs *RuleSet) URL() *RuleSet {
	return rs.Add(&URLRule{})
}

// Numeric adds a numeric validation rule.
func (rs *RuleSet) Numeric() *RuleSet {
	return rs.Add(&NumericRule{})
}

// Alpha adds an alphabetic validation rule.
func (rs *RuleSet) Alpha() *RuleSet {
	return rs.Add(&AlphaRule{})
}

// AlphaNumeric adds an alphanumeric validation rule.
func (rs *RuleSet) AlphaNumeric() *RuleSet {
	return rs.Add(&AlphaNumericRule{})
}

// Validate validates a value against all rules.
func (rs *RuleSet) Validate(value any) []string {
	var errors []string
	for _, rule := range rs.Rules {
		if msg := rule.Validate(value); msg != "" {
			errors = append(errors, msg)
		}
	}
	return errors
}

// --- Rule Implementations ---

// RequiredRule validates that a value is not empty.
type RequiredRule struct{}

func (r *RequiredRule) GetName() string { return "required" }
func (r *RequiredRule) Validate(value any) string {
	if value == nil {
		return "This field is required"
	}
	if str, ok := value.(string); ok && strings.TrimSpace(str) == "" {
		return "This field is required"
	}
	return ""
}

// EmailRule validates email format.
type EmailRule struct{}

func (r *EmailRule) GetName() string { return "email" }
func (r *EmailRule) Validate(value any) string {
	str, ok := value.(string)
	if !ok || str == "" {
		return ""
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(str) {
		return "Invalid email format"
	}
	return ""
}

// MinRule validates minimum length or value.
type MinRule struct {
	Min int
}

func (r *MinRule) GetName() string { return "min" }
func (r *MinRule) Validate(value any) string {
	switch v := value.(type) {
	case string:
		if utf8.RuneCountInString(v) < r.Min {
			return fmt.Sprintf("Must be at least %d characters", r.Min)
		}
	case int:
		if v < r.Min {
			return fmt.Sprintf("Must be at least %d", r.Min)
		}
	case int64:
		if v < int64(r.Min) {
			return fmt.Sprintf("Must be at least %d", r.Min)
		}
	case float64:
		if v < float64(r.Min) {
			return fmt.Sprintf("Must be at least %d", r.Min)
		}
	}
	return ""
}

// MaxRule validates maximum length or value.
type MaxRule struct {
	Max int
}

func (r *MaxRule) GetName() string { return "max" }
func (r *MaxRule) Validate(value any) string {
	switch v := value.(type) {
	case string:
		if utf8.RuneCountInString(v) > r.Max {
			return fmt.Sprintf("Must be at most %d characters", r.Max)
		}
	case int:
		if v > r.Max {
			return fmt.Sprintf("Must be at most %d", r.Max)
		}
	case int64:
		if v > int64(r.Max) {
			return fmt.Sprintf("Must be at most %d", r.Max)
		}
	case float64:
		if v > float64(r.Max) {
			return fmt.Sprintf("Must be at most %d", r.Max)
		}
	}
	return ""
}

// BetweenRule validates that a value is between min and max.
type BetweenRule struct {
	Min, Max int
}

func (r *BetweenRule) GetName() string { return "between" }
func (r *BetweenRule) Validate(value any) string {
	minRule := &MinRule{Min: r.Min}
	maxRule := &MaxRule{Max: r.Max}
	if msg := minRule.Validate(value); msg != "" {
		return fmt.Sprintf("Must be between %d and %d", r.Min, r.Max)
	}
	if msg := maxRule.Validate(value); msg != "" {
		return fmt.Sprintf("Must be between %d and %d", r.Min, r.Max)
	}
	return ""
}

// RegexRule validates against a regex pattern.
type RegexRule struct {
	Pattern string
}

func (r *RegexRule) GetName() string { return "regex" }
func (r *RegexRule) Validate(value any) string {
	str, ok := value.(string)
	if !ok || str == "" {
		return ""
	}
	regex, err := regexp.Compile(r.Pattern)
	if err != nil {
		return "Invalid pattern"
	}
	if !regex.MatchString(str) {
		return "Invalid format"
	}
	return ""
}

// InRule validates that a value is in a list.
type InRule struct {
	Values []string
}

func (r *InRule) GetName() string { return "in" }
func (r *InRule) Validate(value any) string {
	str := fmt.Sprintf("%v", value)
	for _, v := range r.Values {
		if v == str {
			return ""
		}
	}
	return fmt.Sprintf("Must be one of: %s", strings.Join(r.Values, ", "))
}

// URLRule validates URL format.
type URLRule struct{}

func (r *URLRule) GetName() string { return "url" }
func (r *URLRule) Validate(value any) string {
	str, ok := value.(string)
	if !ok || str == "" {
		return ""
	}
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(str) {
		return "Invalid URL format"
	}
	return ""
}

// NumericRule validates that a value is numeric.
type NumericRule struct{}

func (r *NumericRule) GetName() string { return "numeric" }
func (r *NumericRule) Validate(value any) string {
	str, ok := value.(string)
	if !ok {
		return ""
	}
	if _, err := strconv.ParseFloat(str, 64); err != nil {
		return "Must be a number"
	}
	return ""
}

// AlphaRule validates that a value contains only letters.
type AlphaRule struct{}

func (r *AlphaRule) GetName() string { return "alpha" }
func (r *AlphaRule) Validate(value any) string {
	str, ok := value.(string)
	if !ok || str == "" {
		return ""
	}
	alphaRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
	if !alphaRegex.MatchString(str) {
		return "Must contain only letters"
	}
	return ""
}

// AlphaNumericRule validates that a value contains only letters and numbers.
type AlphaNumericRule struct{}

func (r *AlphaNumericRule) GetName() string { return "alphanumeric" }
func (r *AlphaNumericRule) Validate(value any) string {
	str, ok := value.(string)
	if !ok || str == "" {
		return ""
	}
	alphaNumRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !alphaNumRegex.MatchString(str) {
		return "Must contain only letters and numbers"
	}
	return ""
}

// --- Helper Functions ---

// ParseRules parses a string of rules (e.g., "required|email|min:5") into a RuleSet.
func ParseRules(fieldName string, rulesStr string) *RuleSet {
	rs := NewRuleSet(fieldName)
	if rulesStr == "" {
		return rs
	}

	parts := strings.Split(rulesStr, "|")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check for parameters (e.g., "min:5")
		ruleParts := strings.SplitN(part, ":", 2)
		ruleName := ruleParts[0]
		param := ""
		if len(ruleParts) > 1 {
			param = ruleParts[1]
		}

		switch ruleName {
		case "required":
			rs.Required()
		case "email":
			rs.Email()
		case "url":
			rs.URL()
		case "numeric":
			rs.Numeric()
		case "alpha":
			rs.Alpha()
		case "alphanumeric":
			rs.AlphaNumeric()
		case "min":
			if val, err := strconv.Atoi(param); err == nil {
				rs.Min(val)
			}
		case "max":
			if val, err := strconv.Atoi(param); err == nil {
				rs.Max(val)
			}
		case "between":
			parts := strings.Split(param, ",")
			if len(parts) == 2 {
				min, _ := strconv.Atoi(parts[0])
				max, _ := strconv.Atoi(parts[1])
				rs.Between(min, max)
			}
		case "in":
			values := strings.Split(param, ",")
			rs.In(values...)
		case "regex":
			rs.Regex(param)
		}
	}

	return rs
}

// ValidateMap validates a map of field values against a map of rule strings.
func ValidateMap(data map[string]any, rules map[string]string) map[string][]string {
	errors := make(map[string][]string)

	for field, ruleStr := range rules {
		rs := ParseRules(field, ruleStr)
		value := data[field]
		if fieldErrors := rs.Validate(value); len(fieldErrors) > 0 {
			errors[field] = fieldErrors
		}
	}

	return errors
}

// HasValidationErrors checks if there are any validation errors.
func HasValidationErrors(errors map[string][]string) bool {
	return len(errors) > 0
}

// FirstValidationError returns the first error for a field.
func FirstValidationError(errors map[string][]string, field string) string {
	if errs, ok := errors[field]; ok && len(errs) > 0 {
		return errs[0]
	}
	return ""
}
