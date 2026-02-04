package form

import (
	"fmt"
	"html/template"
)

// BaseField contains common logic.
type BaseField struct {
	Name        string
	LabelStr    string
	Value       any
	Placeholder string
	HelpText    string
	Required    bool
	Disabled    bool
	Hidden      bool
	Rules       []string
}

func (b *BaseField) GetName() string                  { return b.Name }
func (b *BaseField) GetLabel() string                 { return b.LabelStr }
func (b *BaseField) GetValue() any                    { return b.Value }
func (b *BaseField) GetPlaceholder() string           { return b.Placeholder }
func (b *BaseField) GetHelp() string                  { return b.HelpText }
func (b *BaseField) IsRequired() bool                 { return b.Required }
func (b *BaseField) IsDisabled() bool                 { return b.Disabled }
func (b *BaseField) IsVisible() bool                  { return !b.Hidden }
func (b *BaseField) GetComponentType() string         { return "field" }
func (b *BaseField) GetAttributes() template.HTMLAttr { return "" }
func (b *BaseField) GetRules() []string               { return b.Rules }

// HasValue returns true if the field has a non-nil value.
func (b *BaseField) HasValue() bool { return b.Value != nil }

// GetValueString returns the value as a string.
func (b *BaseField) GetValueString() string {
	if b.Value == nil {
		return ""
	}
	return fmt.Sprintf("%v", b.Value)
}

// IsChecked returns true if the value is a bool true (for checkbox).
func (b *BaseField) IsChecked() bool {
	if b.Value == nil {
		return false
	}
	if val, ok := b.Value.(bool); ok {
		return val
	}
	return false
}

// TextInput represents a text input field.
type TextInput struct {
	BaseField
	Type string
}

// Text creates a standard text field.
func Text(name string) *TextInput {
	return &TextInput{
		BaseField: BaseField{Name: name, LabelStr: name},
		Type:      "text",
	}
}

// Email creates an email field.
func Email(name string) *TextInput {
	t := Text(name)
	t.Type = "email"
	t.Rules = append(t.Rules, "email")
	return t
}

// Password creates a password field.
func Password(name string) *TextInput {
	t := Text(name)
	t.Type = "password"
	return t
}

// Number creates a numeric field.
func Number(name string) *TextInput {
	t := Text(name)
	t.Type = "number"
	return t
}

// Label sets the field label.
func (f *TextInput) Label(label string) *TextInput {
	f.LabelStr = label
	return f
}

// Placeholder sets the placeholder.
func (f *TextInput) Placeholder(text string) *TextInput {
	f.BaseField.Placeholder = text
	return f
}

// HelperText sets the help text.
func (f *TextInput) HelperText(text string) *TextInput {
	f.HelpText = text
	return f
}

// Required makes the field required.
func (f *TextInput) Required() *TextInput {
	f.BaseField.Required = true
	f.Rules = append(f.Rules, "required")
	return f
}

// Disabled disables the field.
func (f *TextInput) Disabled() *TextInput {
	f.BaseField.Disabled = true
	return f
}

// Default sets the default value.
func (f *TextInput) Default(val any) *TextInput {
	f.Value = val
	return f
}

// Textarea represents a textarea field.
type Textarea struct {
	BaseField
	RowCount int
}

// NewTextarea creates a textarea field.
func NewTextarea(name string) *Textarea {
	return &Textarea{
		BaseField: BaseField{Name: name, LabelStr: name},
		RowCount:  3,
	}
}

// Label sets the label.
func (t *Textarea) Label(label string) *Textarea {
	t.LabelStr = label
	return t
}

// Rows sets the number of rows.
func (t *Textarea) Rows(rows int) *Textarea {
	t.RowCount = rows
	return t
}

// Required makes the field required.
func (t *Textarea) Required() *Textarea {
	t.BaseField.Required = true
	t.Rules = append(t.Rules, "required")
	return t
}

// SelectOption represents a select option.
type SelectOption struct {
	Label string
	Value string
}

// Select represents a select field.
type Select struct {
	BaseField
	Options []SelectOption
}

// NewSelect creates a select field.
func NewSelect(name string) *Select {
	return &Select{
		BaseField: BaseField{Name: name, LabelStr: name},
		Options:   make([]SelectOption, 0),
	}
}

// SetOptions sets the options.
func (s *Select) SetOptions(options map[string]string) *Select {
	for v, l := range options {
		s.Options = append(s.Options, SelectOption{Value: v, Label: l})
	}
	return s
}

// Label sets the label.
func (s *Select) Label(label string) *Select {
	s.LabelStr = label
	return s
}

// Required makes the field required.
func (s *Select) Required() *Select {
	s.BaseField.Required = true
	s.Rules = append(s.Rules, "required")
	return s
}

// Default sets the default value.
func (s *Select) Default(val any) *Select {
	s.Value = val
	return s
}

// Checkbox represents a checkbox field.
type Checkbox struct {
	BaseField
}

// NewCheckbox creates a checkbox field.
func NewCheckbox(name string) *Checkbox {
	return &Checkbox{
		BaseField: BaseField{Name: name, LabelStr: name},
	}
}

// Label sets the label.
func (c *Checkbox) Label(label string) *Checkbox {
	c.LabelStr = label
	return c
}

// Default sets the default value.
func (c *Checkbox) Default(val bool) *Checkbox {
	c.Value = val
	return c
}

// FileUpload represents a file upload field.
type FileUpload struct {
	BaseField
	AcceptTypes   string
	MaxFileSize   int64
	AllowMultiple bool
}

// NewFileUpload creates a file upload field.
func NewFileUpload(name string) *FileUpload {
	return &FileUpload{
		BaseField: BaseField{Name: name, LabelStr: name},
	}
}

// Label sets the label.
func (f *FileUpload) Label(label string) *FileUpload {
	f.LabelStr = label
	return f
}

// Accept sets the accepted file types.
func (f *FileUpload) Accept(accept string) *FileUpload {
	f.AcceptTypes = accept
	return f
}

// MaxSize sets the maximum size in bytes.
func (f *FileUpload) MaxSize(size int64) *FileUpload {
	f.MaxFileSize = size
	return f
}

// Multiple allows multiple files.
func (f *FileUpload) Multiple() *FileUpload {
	f.AllowMultiple = true
	return f
}

// Required makes the field required.
func (f *FileUpload) Required() *FileUpload {
	f.BaseField.Required = true
	f.Rules = append(f.Rules, "required")
	return f
}
