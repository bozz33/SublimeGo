package form

import (
	"html/template"
)

// Component is the base interface for any form element.
type Component interface {
	IsVisible() bool
	ComponentType() string
}

// Field defines the contract for an input field.
type Field interface {
	Component
	Name() string
	Label() string
	Value() any
	Placeholder() string
	Help() string
	IsRequired() bool
	IsDisabled() bool
	Attributes() template.HTMLAttr
	Rules() []string
}

// Layout defines the contract for layout.
type Layout interface {
	Component
	Schema() []Component
}
