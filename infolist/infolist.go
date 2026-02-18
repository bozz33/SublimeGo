package infolist

import "fmt"

// EntryType defines the display type of an infolist entry.
type EntryType string

const (
	EntryTypeText     EntryType = "text"
	EntryTypeBadge    EntryType = "badge"
	EntryTypeBoolean  EntryType = "boolean"
	EntryTypeDate     EntryType = "date"
	EntryTypeImage    EntryType = "image"
	EntryTypeColor    EntryType = "color"
	EntryTypeKeyValue EntryType = "keyvalue"
)

// Entry is a single read-only field in an Infolist.
type Entry struct {
	Name       string
	LabelStr   string
	Value      any
	Type       EntryType
	BadgeColor string // for EntryTypeBadge
	Format     string // for EntryTypeDate (Go time format)
	IsCopyable bool
	Hidden     bool
	HelpText   string
}

// GetLabel returns the display label.
func (e *Entry) GetLabel() string { return e.LabelStr }

// GetValueStr returns the value as a string.
func (e *Entry) GetValueStr() string {
	if e.Value == nil {
		return ""
	}
	return fmt.Sprintf("%v", e.Value)
}

// IsVisible returns true if the entry should be displayed.
func (e *Entry) IsVisible() bool { return !e.Hidden }

// Section groups entries under a heading.
type Section struct {
	Heading string
	Columns int // 1, 2, or 3 — default 2
	Entries []*Entry
}

// Infolist is the top-level container for a read-only detail view.
type Infolist struct {
	Sections []*Section
}

// New creates an empty Infolist.
func New() *Infolist {
	return &Infolist{}
}

// AddSection appends a section and returns the Infolist for chaining.
func (il *Infolist) AddSection(s *Section) *Infolist {
	if s.Columns == 0 {
		s.Columns = 2
	}
	il.Sections = append(il.Sections, s)
	return il
}

// --- Fluent entry builders ---

// TextEntry creates a plain text entry.
func TextEntry(name, label string, value any) *Entry {
	return &Entry{Name: name, LabelStr: label, Value: value, Type: EntryTypeText}
}

// BadgeEntry creates a badge entry with a color.
func BadgeEntry(name, label string, value any, color string) *Entry {
	return &Entry{Name: name, LabelStr: label, Value: value, Type: EntryTypeBadge, BadgeColor: color}
}

// BooleanEntry creates a boolean (✓/✗) entry.
func BooleanEntry(name, label string, value any) *Entry {
	return &Entry{Name: name, LabelStr: label, Value: value, Type: EntryTypeBoolean}
}

// DateEntry creates a date entry with a Go time format string.
func DateEntry(name, label string, value any, format string) *Entry {
	if format == "" {
		format = "2006-01-02"
	}
	return &Entry{Name: name, LabelStr: label, Value: value, Type: EntryTypeDate, Format: format}
}

// ImageEntry creates an image entry (renders an <img> tag).
func ImageEntry(name, label string, value any) *Entry {
	return &Entry{Name: name, LabelStr: label, Value: value, Type: EntryTypeImage}
}

// ColorEntry creates a color swatch entry.
func ColorEntry(name, label string, value any) *Entry {
	return &Entry{Name: name, LabelStr: label, Value: value, Type: EntryTypeColor}
}

// WithCopy marks the entry as copyable (adds a copy button).
func (e *Entry) WithCopy() *Entry {
	e.IsCopyable = true
	return e
}

// Help adds a help text below the entry.
func (e *Entry) Help(text string) *Entry {
	e.HelpText = text
	return e
}

// Hide hides the entry conditionally.
func (e *Entry) Hide(hidden bool) *Entry {
	e.Hidden = hidden
	return e
}
