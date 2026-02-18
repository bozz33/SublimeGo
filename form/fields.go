package form

import (
	"fmt"
	"html/template"
	"strings"
)

// BaseField contains common logic.
type BaseField struct {
	fieldName        string
	LabelStr         string
	fieldValue       any
	fieldPlaceholder string
	HelpText         string
	Required         bool
	Disabled         bool
	Hidden           bool
	fieldRules       []string
}

func (b *BaseField) Name() string                  { return b.fieldName }
func (b *BaseField) Label() string                 { return b.LabelStr }
func (b *BaseField) Value() any                    { return b.fieldValue }
func (b *BaseField) Placeholder() string           { return b.fieldPlaceholder }
func (b *BaseField) Help() string                  { return b.HelpText }
func (b *BaseField) IsRequired() bool              { return b.Required }
func (b *BaseField) IsDisabled() bool              { return b.Disabled }
func (b *BaseField) IsVisible() bool               { return !b.Hidden }
func (b *BaseField) ComponentType() string         { return "field" }
func (b *BaseField) Attributes() template.HTMLAttr { return "" }
func (b *BaseField) Rules() []string               { return b.fieldRules }

// RulesString returns the rules as a pipe-separated string for validation.
func (b *BaseField) RulesString() string {
	return strings.Join(b.fieldRules, "|")
}

// HasValue returns true if the field has a non-nil value.
func (b *BaseField) HasValue() bool { return b.fieldValue != nil }

// ValueString returns the value as a string.
func (b *BaseField) ValueString() string {
	if b.fieldValue == nil {
		return ""
	}
	return fmt.Sprintf("%v", b.fieldValue)
}

// IsChecked returns true if the value is a bool true (for checkbox).
func (b *BaseField) IsChecked() bool {
	if b.fieldValue == nil {
		return false
	}
	if val, ok := b.fieldValue.(bool); ok {
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
		BaseField: BaseField{fieldName: name, LabelStr: name},
		Type:      "text",
	}
}

// Email creates an email field.
func Email(name string) *TextInput {
	t := Text(name)
	t.Type = "email"
	t.fieldRules = append(t.fieldRules, "email")
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

// WithPlaceholder sets the placeholder.
func (f *TextInput) WithPlaceholder(text string) *TextInput {
	f.fieldPlaceholder = text
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
	f.fieldRules = append(f.fieldRules, "required")
	return f
}

// Disabled disables the field.
func (f *TextInput) Disabled() *TextInput {
	f.BaseField.Disabled = true
	return f
}

// Default sets the default value.
func (f *TextInput) Default(val any) *TextInput {
	f.fieldValue = val
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
		BaseField: BaseField{fieldName: name, LabelStr: name},
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
	t.fieldRules = append(t.fieldRules, "required")
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
	selectOptions []SelectOption
}

// NewSelect creates a select field.
func NewSelect(name string) *Select {
	return &Select{
		BaseField:     BaseField{fieldName: name, LabelStr: name},
		selectOptions: make([]SelectOption, 0),
	}
}

// Options sets the select options.
func (s *Select) Options(options map[string]string) *Select {
	for v, l := range options {
		s.selectOptions = append(s.selectOptions, SelectOption{Value: v, Label: l})
	}
	return s
}

// Label sets the label.
func (s *Select) Label(label string) *Select {
	s.LabelStr = label
	return s
}

// SelectOptions returns the available options.
func (s *Select) SelectOptions() []SelectOption { return s.selectOptions }

// Required makes the field required.
func (s *Select) Required() *Select {
	s.BaseField.Required = true
	s.fieldRules = append(s.fieldRules, "required")
	return s
}

// Default sets the default value.
func (s *Select) Default(val any) *Select {
	s.fieldValue = val
	return s
}

// Checkbox represents a checkbox field.
type Checkbox struct {
	BaseField
}

// NewCheckbox creates a checkbox field.
func NewCheckbox(name string) *Checkbox {
	return &Checkbox{
		BaseField: BaseField{fieldName: name, LabelStr: name},
	}
}

// Label sets the label.
func (c *Checkbox) Label(label string) *Checkbox {
	c.LabelStr = label
	return c
}

// Default sets the default value.
func (c *Checkbox) Default(val bool) *Checkbox {
	c.fieldValue = val
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
		BaseField: BaseField{fieldName: name, LabelStr: name},
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
	f.fieldRules = append(f.fieldRules, "required")
	return f
}

// DatePicker represents a date/datetime input field.
type DatePicker struct {
	BaseField
	Type    string // "date", "datetime-local", "time", "month", "week"
	MinDate string
	MaxDate string
	Format  string
}

// Date creates a date picker field (YYYY-MM-DD).
func Date(name string) *DatePicker {
	return &DatePicker{
		BaseField: BaseField{fieldName: name, LabelStr: name},
		Type:      "date",
	}
}

// DateTime creates a datetime-local picker field.
func DateTime(name string) *DatePicker {
	return &DatePicker{
		BaseField: BaseField{fieldName: name, LabelStr: name},
		Type:      "datetime-local",
	}
}

// Time creates a time picker field.
func Time(name string) *DatePicker {
	return &DatePicker{
		BaseField: BaseField{fieldName: name, LabelStr: name},
		Type:      "time",
	}
}

// Label sets the label.
func (d *DatePicker) Label(label string) *DatePicker {
	d.LabelStr = label
	return d
}

// Min sets the minimum date (YYYY-MM-DD).
func (d *DatePicker) Min(date string) *DatePicker {
	d.MinDate = date
	return d
}

// Max sets the maximum date (YYYY-MM-DD).
func (d *DatePicker) Max(date string) *DatePicker {
	d.MaxDate = date
	return d
}

// Required makes the field required.
func (d *DatePicker) Required() *DatePicker {
	d.BaseField.Required = true
	d.fieldRules = append(d.fieldRules, "required")
	return d
}

// Default sets the default value.
func (d *DatePicker) Default(val any) *DatePicker {
	d.fieldValue = val
	return d
}

// HiddenField represents a hidden input field.
type HiddenField struct {
	BaseField
}

// Hidden creates a hidden field with a fixed value.
func Hidden(name string, value any) *HiddenField {
	return &HiddenField{
		BaseField: BaseField{fieldName: name, LabelStr: name, fieldValue: value, Hidden: true},
	}
}

// Toggle represents a toggle switch (boolean, rendered differently from Checkbox).
type Toggle struct {
	BaseField
	OnLabel  string
	OffLabel string
}

// NewToggle creates a toggle switch field.
func NewToggle(name string) *Toggle {
	return &Toggle{
		BaseField: BaseField{fieldName: name, LabelStr: name},
		OnLabel:   "Yes",
		OffLabel:  "No",
	}
}

// Label sets the label.
func (t *Toggle) Label(label string) *Toggle {
	t.LabelStr = label
	return t
}

// Labels sets the on/off labels.
func (t *Toggle) Labels(on, off string) *Toggle {
	t.OnLabel = on
	t.OffLabel = off
	return t
}

// Default sets the default boolean value.
func (t *Toggle) Default(val bool) *Toggle {
	t.fieldValue = val
	return t
}

// RepeaterField represents a dynamic multi-value field (list of sub-fields).
// Each entry in the repeater is a map of field name -> value.
type RepeaterField struct {
	BaseField
	SubFields []Field
	MinItems  int
	MaxItems  int
	AddLabel  string
}

// Repeater creates a repeater field with the given sub-fields.
func Repeater(name string, subFields ...Field) *RepeaterField {
	return &RepeaterField{
		BaseField: BaseField{fieldName: name, LabelStr: name},
		SubFields: subFields,
		MinItems:  0,
		MaxItems:  0,
		AddLabel:  "Add item",
	}
}

// Label sets the label.
func (r *RepeaterField) Label(label string) *RepeaterField {
	r.LabelStr = label
	return r
}

// Min sets the minimum number of items.
func (r *RepeaterField) Min(n int) *RepeaterField {
	r.MinItems = n
	return r
}

// Max sets the maximum number of items.
func (r *RepeaterField) Max(n int) *RepeaterField {
	r.MaxItems = n
	return r
}

// AddButtonLabel sets the label for the "add item" button.
func (r *RepeaterField) AddButtonLabel(label string) *RepeaterField {
	r.AddLabel = label
	return r
}

// ---------------------------------------------------------------------------
// RichEditor — renders a WYSIWYG editor (e.g. Trix, TipTap, Quill).
// ---------------------------------------------------------------------------

// RichEditor represents a rich-text / WYSIWYG editor field.
type RichEditor struct {
	BaseField
	Toolbar   []string // e.g. ["bold","italic","link","heading","list","image"]
	MaxLength int
}

// NewRichEditor creates a rich editor field.
func NewRichEditor(name string) *RichEditor {
	return &RichEditor{
		BaseField: BaseField{fieldName: name, LabelStr: name},
		Toolbar:   []string{"bold", "italic", "underline", "link", "heading", "list", "image", "code"},
	}
}

// Label sets the label.
func (r *RichEditor) Label(label string) *RichEditor {
	r.LabelStr = label
	return r
}

// WithToolbar overrides the default toolbar buttons.
func (r *RichEditor) WithToolbar(items ...string) *RichEditor {
	r.Toolbar = items
	return r
}

// WithMaxLength sets the maximum character count.
func (r *RichEditor) WithMaxLength(n int) *RichEditor {
	r.MaxLength = n
	return r
}

// Required makes the field required.
func (r *RichEditor) Required() *RichEditor {
	r.BaseField.Required = true
	r.fieldRules = append(r.fieldRules, "required")
	return r
}

// Default sets the default HTML value.
func (r *RichEditor) Default(val string) *RichEditor {
	r.fieldValue = val
	return r
}

// ComponentType returns the component type identifier.
func (r *RichEditor) ComponentType() string { return "rich_editor" }

// ---------------------------------------------------------------------------
// MarkdownEditor — renders a Markdown editor with preview.
// ---------------------------------------------------------------------------

// MarkdownEditor represents a Markdown editor field with live preview.
type MarkdownEditor struct {
	BaseField
	RowCount int
}

// NewMarkdownEditor creates a Markdown editor field.
func NewMarkdownEditor(name string) *MarkdownEditor {
	return &MarkdownEditor{
		BaseField: BaseField{fieldName: name, LabelStr: name},
		RowCount:  10,
	}
}

// Label sets the label.
func (m *MarkdownEditor) Label(label string) *MarkdownEditor {
	m.LabelStr = label
	return m
}

// Rows sets the number of visible rows.
func (m *MarkdownEditor) Rows(rows int) *MarkdownEditor {
	m.RowCount = rows
	return m
}

// Required makes the field required.
func (m *MarkdownEditor) Required() *MarkdownEditor {
	m.BaseField.Required = true
	m.fieldRules = append(m.fieldRules, "required")
	return m
}

// Default sets the default Markdown value.
func (m *MarkdownEditor) Default(val string) *MarkdownEditor {
	m.fieldValue = val
	return m
}

// ComponentType returns the component type identifier.
func (m *MarkdownEditor) ComponentType() string { return "markdown_editor" }

// ---------------------------------------------------------------------------
// TagsInput — multi-value tag/chip input.
// ---------------------------------------------------------------------------

// TagsInput represents a tag/chip input field that stores multiple string values.
type TagsInput struct {
	BaseField
	Suggestions []string
	MaxTags     int
	Separator   string // delimiter for form submission, default ","
}

// NewTagsInput creates a tags input field.
func NewTagsInput(name string) *TagsInput {
	return &TagsInput{
		BaseField: BaseField{fieldName: name, LabelStr: name},
		Separator: ",",
	}
}

// Label sets the label.
func (t *TagsInput) Label(label string) *TagsInput {
	t.LabelStr = label
	return t
}

// WithSuggestions sets the autocomplete suggestions.
func (t *TagsInput) WithSuggestions(suggestions ...string) *TagsInput {
	t.Suggestions = suggestions
	return t
}

// WithMaxTags limits the number of tags.
func (t *TagsInput) WithMaxTags(n int) *TagsInput {
	t.MaxTags = n
	return t
}

// WithSeparator sets the delimiter used in form submission (default ",").
func (t *TagsInput) WithSeparator(sep string) *TagsInput {
	t.Separator = sep
	return t
}

// Required makes the field required.
func (t *TagsInput) Required() *TagsInput {
	t.BaseField.Required = true
	t.fieldRules = append(t.fieldRules, "required")
	return t
}

// Default sets the default tags.
func (t *TagsInput) Default(tags []string) *TagsInput {
	t.fieldValue = tags
	return t
}

// ComponentType returns the component type identifier.
func (t *TagsInput) ComponentType() string { return "tags_input" }

// Tags returns the current value as a string slice.
func (t *TagsInput) Tags() []string {
	if v, ok := t.fieldValue.([]string); ok {
		return v
	}
	return nil
}

// ---------------------------------------------------------------------------
// KeyValue — key-value pair input.
// ---------------------------------------------------------------------------

// KeyValuePair represents a single key-value entry.
type KeyValuePair struct {
	Key   string
	Value string
}

// KeyValue represents a dynamic key-value pair input field.
type KeyValue struct {
	BaseField
	KeyLabel   string
	ValueLabel string
	MaxPairs   int
	AddLabel   string
}

// NewKeyValue creates a key-value input field.
func NewKeyValue(name string) *KeyValue {
	return &KeyValue{
		BaseField:  BaseField{fieldName: name, LabelStr: name},
		KeyLabel:   "Key",
		ValueLabel: "Value",
		AddLabel:   "Add pair",
	}
}

// Label sets the label.
func (kv *KeyValue) Label(label string) *KeyValue {
	kv.LabelStr = label
	return kv
}

// WithLabels sets the key and value column labels.
func (kv *KeyValue) WithLabels(keyLabel, valueLabel string) *KeyValue {
	kv.KeyLabel = keyLabel
	kv.ValueLabel = valueLabel
	return kv
}

// WithMaxPairs limits the number of pairs.
func (kv *KeyValue) WithMaxPairs(n int) *KeyValue {
	kv.MaxPairs = n
	return kv
}

// AddButtonLabel sets the label for the "add pair" button.
func (kv *KeyValue) AddButtonLabel(label string) *KeyValue {
	kv.AddLabel = label
	return kv
}

// Default sets the default pairs.
func (kv *KeyValue) Default(pairs []KeyValuePair) *KeyValue {
	kv.fieldValue = pairs
	return kv
}

// ComponentType returns the component type identifier.
func (kv *KeyValue) ComponentType() string { return "key_value" }

// ---------------------------------------------------------------------------
// ColorPicker — color selection input.
// ---------------------------------------------------------------------------

// ColorPicker represents a color picker input field.
type ColorPicker struct {
	BaseField
	Swatches []string // predefined color swatches (hex)
}

// NewColorPicker creates a color picker field.
func NewColorPicker(name string) *ColorPicker {
	return &ColorPicker{
		BaseField: BaseField{fieldName: name, LabelStr: name},
	}
}

// Label sets the label.
func (c *ColorPicker) Label(label string) *ColorPicker {
	c.LabelStr = label
	return c
}

// WithSwatches sets predefined color swatches.
func (c *ColorPicker) WithSwatches(colors ...string) *ColorPicker {
	c.Swatches = colors
	return c
}

// Required makes the field required.
func (c *ColorPicker) Required() *ColorPicker {
	c.BaseField.Required = true
	c.fieldRules = append(c.fieldRules, "required")
	return c
}

// Default sets the default color (hex string, e.g. "#22c55e").
func (c *ColorPicker) Default(hex string) *ColorPicker {
	c.fieldValue = hex
	return c
}

// ComponentType returns the component type identifier.
func (c *ColorPicker) ComponentType() string { return "color_picker" }

// ---------------------------------------------------------------------------
// Slider — range slider input.
// ---------------------------------------------------------------------------

// Slider represents a range slider input field.
type Slider struct {
	BaseField
	Min  float64
	Max  float64
	Step float64
	Unit string // optional display unit (e.g. "%", "px", "kg")
}

// NewSlider creates a slider field.
func NewSlider(name string) *Slider {
	return &Slider{
		BaseField: BaseField{fieldName: name, LabelStr: name},
		Min:       0,
		Max:       100,
		Step:      1,
	}
}

// Label sets the label.
func (s *Slider) Label(label string) *Slider {
	s.LabelStr = label
	return s
}

// Range sets the min and max values.
func (s *Slider) Range(min, max float64) *Slider {
	s.Min = min
	s.Max = max
	return s
}

// WithStep sets the step increment.
func (s *Slider) WithStep(step float64) *Slider {
	s.Step = step
	return s
}

// WithUnit sets the display unit suffix.
func (s *Slider) WithUnit(unit string) *Slider {
	s.Unit = unit
	return s
}

// Default sets the default value.
func (s *Slider) Default(val float64) *Slider {
	s.fieldValue = val
	return s
}

// ComponentType returns the component type identifier.
func (s *Slider) ComponentType() string { return "slider" }
