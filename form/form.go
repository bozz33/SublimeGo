package form

import (
	"context"

	"github.com/bozz33/sublimego/validation"
)

// Form is the main container.
type Form struct {
	Schema []Component
	Model  any
	State  map[string]any
	Errors map[string][]string
}

// New creates a new form.
func New() *Form {
	return &Form{
		Schema: make([]Component, 0),
		State:  make(map[string]any),
		Errors: make(map[string][]string),
	}
}

// SetSchema sets the form structure.
func (f *Form) SetSchema(components ...Component) *Form {
	f.Schema = components
	return f
}

// Bind binds a model to the form for pre-filling.
func (f *Form) Bind(model any) *Form {
	f.Model = model
	return f
}

// SaveProcessing handles logic before saving.
func (f *Form) SaveProcessing(ctx context.Context) error {
	return nil
}

// Validate validates the form data against all field rules.
func (f *Form) Validate(data map[string]any) bool {
	f.Errors = make(map[string][]string)

	for _, component := range f.Schema {
		field, ok := component.(interface {
			GetName() string
			GetRules() []string
		})
		if !ok {
			continue
		}

		fieldName := field.GetName()
		rules := field.GetRules()

		if len(rules) == 0 {
			continue
		}

		// Build rule set from field rules
		ruleSet := validation.NewRuleSet(fieldName)
		for _, rule := range rules {
			parsed := validation.ParseRules(fieldName, rule)
			for _, r := range parsed.Rules {
				ruleSet.Add(r)
			}
		}

		// Validate the field value
		value := data[fieldName]
		if errors := ruleSet.Validate(value); len(errors) > 0 {
			f.Errors[fieldName] = errors
		}
	}

	return len(f.Errors) == 0
}

// GetValidationRules returns all validation rules as a map for use with validation.ValidateMap.
func (f *Form) GetValidationRules() map[string]string {
	rules := make(map[string]string)

	for _, component := range f.Schema {
		field, ok := component.(interface {
			GetName() string
			GetRulesString() string
		})
		if !ok {
			continue
		}

		rulesStr := field.GetRulesString()
		if rulesStr != "" {
			rules[field.GetName()] = rulesStr
		}
	}

	return rules
}

// HasErrors returns true if the form has validation errors.
func (f *Form) HasErrors() bool {
	return len(f.Errors) > 0
}

// GetError returns the first error for a field.
func (f *Form) GetError(fieldName string) string {
	if errors, ok := f.Errors[fieldName]; ok && len(errors) > 0 {
		return errors[0]
	}
	return ""
}

// GetAllErrors returns all errors for a field.
func (f *Form) GetAllErrors(fieldName string) []string {
	return f.Errors[fieldName]
}
