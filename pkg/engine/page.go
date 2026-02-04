package engine

import (
	"context"
	"net/http"

	"github.com/a-h/templ"
)

// Page represents a custom page in the admin panel.
// Unlike Resources which provide CRUD operations, Pages are standalone
// views that can display any content (reports, settings, analytics, etc.)
type Page interface {
	// Slug returns the URL path for this page (e.g., "settings", "reports/sales")
	Slug() string

	// Label returns the display name for navigation
	Label() string

	// Icon returns the icon name for the sidebar
	Icon() string

	// Group returns the navigation group (empty for root level)
	Group() string

	// Sort returns the sort order in navigation
	Sort() int

	// Render returns the page content as a templ component
	Render(ctx context.Context, r *http.Request) templ.Component

	// CanAccess checks if the current user can access this page
	CanAccess(ctx context.Context) bool
}

// BasePage provides default implementations for the Page interface.
// Embed this in your custom pages to inherit defaults.
type BasePage struct {
	slug  string
	label string
	icon  string
	group string
	sort  int
}

// NewBasePage creates a new BasePage with required values.
func NewBasePage(slug, label string) *BasePage {
	return &BasePage{
		slug:  slug,
		label: label,
		icon:  "document",
		group: "",
		sort:  100,
	}
}

// Page interface implementation
func (p *BasePage) Slug() string  { return p.slug }
func (p *BasePage) Label() string { return p.label }
func (p *BasePage) Icon() string  { return p.icon }
func (p *BasePage) Group() string { return p.group }
func (p *BasePage) Sort() int     { return p.sort }

// CanAccess returns true by default (all users can access)
func (p *BasePage) CanAccess(ctx context.Context) bool { return true }

// Fluent setters for configuration
func (p *BasePage) SetSlug(slug string) *BasePage {
	p.slug = slug
	return p
}

func (p *BasePage) SetLabel(label string) *BasePage {
	p.label = label
	return p
}

func (p *BasePage) SetIcon(icon string) *BasePage {
	p.icon = icon
	return p
}

func (p *BasePage) SetGroup(group string) *BasePage {
	p.group = group
	return p
}

func (p *BasePage) SetSort(sort int) *BasePage {
	p.sort = sort
	return p
}

// SimplePage is a minimal page implementation using a render function.
// Useful for quick page creation without defining a full struct.
type SimplePage struct {
	*BasePage
	renderFunc func(ctx context.Context, r *http.Request) templ.Component
	accessFunc func(ctx context.Context) bool
}

// NewSimplePage creates a simple page with a render function.
func NewSimplePage(slug, label string, render func(ctx context.Context, r *http.Request) templ.Component) *SimplePage {
	return &SimplePage{
		BasePage:   NewBasePage(slug, label),
		renderFunc: render,
	}
}

// WithIcon sets the icon.
func (p *SimplePage) WithIcon(icon string) *SimplePage {
	p.BasePage.SetIcon(icon)
	return p
}

// WithGroup sets the navigation group.
func (p *SimplePage) WithGroup(group string) *SimplePage {
	p.BasePage.SetGroup(group)
	return p
}

// WithSort sets the sort order.
func (p *SimplePage) WithSort(sort int) *SimplePage {
	p.BasePage.SetSort(sort)
	return p
}

// WithAccess sets a custom access check function.
func (p *SimplePage) WithAccess(fn func(ctx context.Context) bool) *SimplePage {
	p.accessFunc = fn
	return p
}

// Render returns the page content.
func (p *SimplePage) Render(ctx context.Context, r *http.Request) templ.Component {
	if p.renderFunc != nil {
		return p.renderFunc(ctx, r)
	}
	return emptyComponent()
}

// CanAccess checks if the user can access this page.
func (p *SimplePage) CanAccess(ctx context.Context) bool {
	if p.accessFunc != nil {
		return p.accessFunc(ctx)
	}
	return p.BasePage.CanAccess(ctx)
}
