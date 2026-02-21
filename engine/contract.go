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
	ID        string
	Cells     []string
	RecordURL string // optional: custom URL when clicking the row/first cell
}

// EmptyState configures the empty table placeholder.
type EmptyState struct {
	Icon        string // Material Icon name, default "inbox"
	Title       string // heading text
	Description string // sub-text
	ActionLabel string // optional CTA button label
	ActionURL   string // optional CTA button URL
}

// TableState contains the complete state of a table.
type TableState struct {
	Title          string
	Slug           string
	Columns        []Column
	Rows           []Row
	CanCreate      bool
	CanDelete      bool
	CanView        bool // show a view (eye) button per row
	NewURL         string
	BaseURL        string
	Pagination     *Pagination
	Filters        []FilterDef       // available filter definitions
	ActiveFilters  map[string]string // currently active filter values (key -> value)
	BulkActions    []BulkActionDef   // available bulk actions
	ExportURL      string            // non-empty = show export button
	ImportURL      string            // non-empty = show import button
	Search         string            // current ?search= value
	SortKey        string            // current ?sort= column key
	SortDir        string            // current ?dir= (asc|desc)
	HeaderActions  []HeaderAction    // always-visible action buttons in header
	PollInterval   int               // HTMX polling interval in seconds (0 = disabled)
	EmptyState     *EmptyState       // custom empty state (nil = default)
	HiddenColumns  []string          // column keys hidden by default (user can toggle)
	ColumnManager  bool              // show column visibility toggle button
	ColumnOrder    []string          // ordered list of column keys (empty = default order)
	ReorderColumns bool              // allow drag & drop column reordering
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

// HeaderAction describes a standalone action button shown in the table header.
// Unlike BulkActions, HeaderActions are always visible and not tied to row selection.
type HeaderAction struct {
	Label  string
	URL    string // GET link
	Icon   string
	Color  string // "primary", "secondary", "danger"
	Method string // "GET" (default) or "POST"
}

// Pagination contains pagination info.
type Pagination struct {
	CurrentPage int
	PerPage     int
	Total       int
	LastPage    int
}

// contextKey is the type for context keys.
type contextKey string

const (
	ContextKeyPanel         contextKey = "panel"
	ContextKeyResource      contextKey = "resource"
	ContextKeyUser          contextKey = "user"
	ContextKeyActiveFilters contextKey = "active_filters"
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

// GetActiveFilters retrieves the active filter map from context.
func GetActiveFilters(ctx context.Context) map[string]string {
	if f, ok := ctx.Value(ContextKeyActiveFilters).(map[string]string); ok {
		return f
	}
	return nil
}

// ResourceFilterable is an optional interface for resources that support
// server-side filtering. Implement ListFiltered to apply query filters at DB level.
type ResourceFilterable interface {
	// ListFiltered returns items matching the given filter key/value pairs.
	ListFiltered(ctx context.Context, filters map[string]string) ([]any, error)
}

// ListQuery holds all query parameters for list operations.
type ListQuery struct {
	Filters map[string]string // filter_* query params
	Search  string            // ?search=
	SortKey string            // ?sort=field
	SortDir string            // ?dir=asc|desc
	Page    int               // ?page=N (1-indexed)
	PerPage int               // ?per_page=N
}

// ResourceQueryable is an optional interface for resources that handle
// all list query parameters in a single call (filters + search + sort + pagination).
// Prefer this over ResourceFilterable when you need full control.
type ResourceQueryable interface {
	ListQuery(ctx context.Context, q ListQuery) (items []any, total int, err error)
}

// ResourceSearchable is an optional interface for resources that support
// full-text search via the ?search= query param.
type ResourceSearchable interface {
	Search(ctx context.Context, query string) ([]any, error)
}

// ResourceHookable is an optional interface for resources that need
// lifecycle hooks around CRUD operations.
type ResourceHookable interface {
	BeforeCreate(ctx context.Context, r *http.Request) error
	AfterCreate(ctx context.Context, item any) error
	BeforeUpdate(ctx context.Context, id string, r *http.Request) error
	AfterUpdate(ctx context.Context, id string, item any) error
	BeforeDelete(ctx context.Context, id string) error
	AfterDelete(ctx context.Context, id string) error
}
