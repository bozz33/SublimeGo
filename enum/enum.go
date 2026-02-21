// Package enum provides interfaces and generic helpers for typed enumerations
// in SublimeGo. Enums implement one or more of the Has* interfaces to expose
// their metadata to form fields, table columns, badges, and navigation items.
package enum

import "github.com/bozz33/sublimego/form"

// HasLabel is implemented by enums that have a human-readable label.
type HasLabel interface {
	Label() string
}

// HasColor is implemented by enums that carry a Tailwind color name
// (e.g. "green", "red", "yellow", "blue", "purple", "gray").
type HasColor interface {
	Color() string
}

// HasIcon is implemented by enums that carry a Material Icons Outlined name
// (e.g. "check_circle", "cancel", "warning").
type HasIcon interface {
	Icon() string
}

// HasDescription is implemented by enums that have a longer description.
type HasDescription interface {
	Description() string
}

// ---------------------------------------------------------------------------
// Generic helpers
// ---------------------------------------------------------------------------

// Options converts a slice of enum values that implement HasLabel into
// a slice of form.SelectOption, ready to pass to form.Select().Options().
func Options[T interface {
	comparable
	HasLabel
}](values []T) []form.SelectOption {
	opts := make([]form.SelectOption, len(values))
	for i, v := range values {
		opts[i] = form.SelectOption{
			Label: v.Label(),
			Value: any(v).(interface{ String() string }).String(),
		}
	}
	return opts
}

// OptionsFromStringer converts a slice of enum values that implement HasLabel
// and fmt.Stringer into form.SelectOption slice.
// Use this when your enum has a String() method (most common case).
func OptionsFromStringer[T interface {
	comparable
	HasLabel
	String() string
}](values []T) []form.SelectOption {
	opts := make([]form.SelectOption, len(values))
	for i, v := range values {
		opts[i] = form.SelectOption{
			Label: v.Label(),
			Value: v.String(),
		}
	}
	return opts
}

// Labels returns a map of value string → label string for all enum values.
// Useful for badge color maps and display lookups.
func Labels[T interface {
	comparable
	HasLabel
	String() string
}](values []T) map[string]string {
	m := make(map[string]string, len(values))
	for _, v := range values {
		m[v.String()] = v.Label()
	}
	return m
}

// Colors returns a map of value string → Tailwind color name for all enum values.
// Useful for table.Badge().Colors() and badge rendering.
func Colors[T interface {
	comparable
	HasColor
	String() string
}](values []T) map[string]string {
	m := make(map[string]string, len(values))
	for _, v := range values {
		m[v.String()] = v.Color()
	}
	return m
}

// Icons returns a map of value string → Material Icon name for all enum values.
func Icons[T interface {
	comparable
	HasIcon
	String() string
}](values []T) map[string]string {
	m := make(map[string]string, len(values))
	for _, v := range values {
		m[v.String()] = v.Icon()
	}
	return m
}

// BadgeColor returns the Tailwind color for a given enum value string,
// with a fallback default. Useful in table cell rendering.
func BadgeColor[T interface {
	comparable
	HasColor
	String() string
}](values []T, value string, defaultColor string) string {
	for _, v := range values {
		if v.String() == value {
			return v.Color()
		}
	}
	return defaultColor
}
