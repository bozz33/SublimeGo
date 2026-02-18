package engine

import (
	"context"
	"net/http"

	"github.com/a-h/templ"
)

// ResourceMeta defines resource metadata.
type ResourceMeta interface {
	Slug() string
	Label() string
	PluralLabel() string
	Icon() string
	Group() string
	Sort() int
}

// ResourceNavigation defines advanced navigation elements.
type ResourceNavigation interface {
	Badge(ctx context.Context) string
	BadgeColor(ctx context.Context) string
}

// ResourceViews defines resource views.
type ResourceViews interface {
	Table(ctx context.Context) templ.Component
	Form(ctx context.Context, item any) templ.Component
}

// ResourcePermissions defines resource permissions.
type ResourcePermissions interface {
	CanCreate(ctx context.Context) bool
	CanRead(ctx context.Context) bool
	CanUpdate(ctx context.Context) bool
	CanDelete(ctx context.Context) bool
}

// ResourceCRUD defines CRUD operations for a resource.
type ResourceCRUD interface {
	List(ctx context.Context) ([]any, error)
	Get(ctx context.Context, id string) (any, error)
	Create(ctx context.Context, r *http.Request) error
	Update(ctx context.Context, id string, r *http.Request) error
	Delete(ctx context.Context, id string) error
	BulkDelete(ctx context.Context, ids []string) error
}

// Resource is the complete interface a resource must implement.
type Resource interface {
	ResourceMeta
	ResourceViews
	ResourcePermissions
	ResourceCRUD
	ResourceNavigation
}

// ResourceViewable is an optional interface for resources that provide
// a read-only detail view (Infolist). Implement it to enable GET /{id}.
type ResourceViewable interface {
	View(ctx context.Context, item any) templ.Component
}

// Column defines a table column.
type Column struct {
	Key        string
	Label      string
	Type       string // "text", "boolean", "date", "badge", "image"
	Sortable   bool
	Searchable bool
}

// Row represents a table row.
type Row struct {
	ID    string
	Cells []string
}

// TableState contains the complete state of a table.
type TableState struct {
	Title         string
	Slug          string
	Columns       []Column
	Rows          []Row
	CanCreate     bool
	CanDelete     bool
	NewURL        string
	BaseURL       string
	Pagination    *Pagination
	Filters       []FilterDef       // available filter definitions
	ActiveFilters map[string]string // currently active filter values (key -> value)
	BulkActions   []BulkActionDef   // available bulk actions
	ExportURL     string            // non-empty = show export button
	ImportURL     string            // non-empty = show import button
}

// FilterDef describes a filter available on the table.
type FilterDef struct {
	Key     string
	Label   string
	Type    string // "select", "boolean", "text"
	Options []FilterOption
}

// FilterOption is a single option in a select filter.
type FilterOption struct {
	Value string
	Label string
}

// BulkActionDef describes a bulk action available on the table.
type BulkActionDef struct {
	Key   string
	Label string
	Icon  string
	Color string // "danger", "warning", "primary"
	URL   string // POST target URL
}

// Pagination contains pagination info.
type Pagination struct {
	CurrentPage int
	PerPage     int
	Total       int
	LastPage    int
}

// FieldType defines the type of a field.
type FieldType string

const (
	FieldText     FieldType = "text"
	FieldEmail    FieldType = "email"
	FieldPassword FieldType = "password"
	FieldNumber   FieldType = "number"
	FieldTextarea FieldType = "textarea"
	FieldSelect   FieldType = "select"
	FieldCheckbox FieldType = "checkbox"
	FieldDate     FieldType = "date"
	FieldFile     FieldType = "file"
)

// Field defines a form field.
type Field struct {
	Name        string
	Label       string
	Type        FieldType
	Value       string
	Placeholder string
	Required    bool
	Disabled    bool
	Options     []Option
	Validation  []string
	Error       string
}

// Option for select fields.
type Option struct {
	Value    string
	Label    string
	Selected bool
}

// FormState contains the complete state of a form.
type FormState struct {
	Title     string
	Slug      string
	ID        string
	ActionURL string
	Method    string
	Fields    []Field
	Errors    map[string]string
	OldValues map[string]string
}

// contextKey is the type for context keys.
type contextKey string

const (
	ContextKeyPanel    contextKey = "panel"
	ContextKeyResource contextKey = "resource"
	ContextKeyUser     contextKey = "user"
)

// GetPanelFromContext retrieves the Panel from context.
func GetPanelFromContext(ctx context.Context) *Panel {
	if p, ok := ctx.Value(ContextKeyPanel).(*Panel); ok {
		return p
	}
	return nil
}

// GetResourceFromContext retrieves the Resource from context.
func GetResourceFromContext(ctx context.Context) Resource {
	if r, ok := ctx.Value(ContextKeyResource).(Resource); ok {
		return r
	}
	return nil
}
