package form

import (
	"html/template"
)

// Component is the base interface for any form element.
type Component interface {
	IsVisible() bool
	GetComponentType() string
}

// Field defines the contract for an input field.
type Field interface {
	Component
	GetName() string
	GetLabel() string
	GetValue() any
	GetPlaceholder() string
	GetHelp() string
	IsRequired() bool
	IsDisabled() bool
	GetAttributes() template.HTMLAttr
	GetRules() []string
}

// Layout defines the contract for layout.
type Layout interface {
	Component
	GetSchema() []Component
}
