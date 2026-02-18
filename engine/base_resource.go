package engine

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/a-h/templ"
)

// BaseResource provides default implementations for the Resource interface.
// Embed this in your resources to inherit defaults and override as needed.
type BaseResource struct {
	slug        string
	label       string
	pluralLabel string
	icon        string
	group       string
	sort        int

	// Table configuration
	tableColumns       []Column
	tableFilters       []FilterDef
	tableBulkActions   []BulkActionDef
	tableHeaderActions []HeaderAction
	tableExportURL     string
	tableImportURL     string
}

// NewBaseResource creates a BaseResource with required values.
func NewBaseResource(slug, label, pluralLabel string) *BaseResource {
	return &BaseResource{
		slug:        slug,
		label:       label,
		pluralLabel: pluralLabel,
		icon:        "folder",
		group:       "",
		sort:        100,
	}
}

// ResourceMeta implementation

func (b *BaseResource) Slug() string        { return b.slug }
func (b *BaseResource) Label() string       { return b.label }
func (b *BaseResource) PluralLabel() string { return b.pluralLabel }
func (b *BaseResource) Icon() string        { return b.icon }
func (b *BaseResource) Group() string       { return b.group }
func (b *BaseResource) Sort() int           { return b.sort }

// Fluent setters for configuration
func (b *BaseResource) SetSlug(slug string) *BaseResource {
	b.slug = slug
	return b
}

func (b *BaseResource) SetLabel(label string) *BaseResource {
	b.label = label
	return b
}

func (b *BaseResource) SetPluralLabel(label string) *BaseResource {
	b.pluralLabel = label
	return b
}

func (b *BaseResource) SetIcon(icon string) *BaseResource {
	b.icon = icon
	return b
}

func (b *BaseResource) SetGroup(group string) *BaseResource {
	b.group = group
	return b
}

func (b *BaseResource) SetSort(sort int) *BaseResource {
	b.sort = sort
	return b
}

// SetTableColumns sets the columns for BuildTableState.
func (b *BaseResource) SetTableColumns(cols ...Column) *BaseResource {
	b.tableColumns = cols
	return b
}

// SetTableFilters sets the filters for BuildTableState.
func (b *BaseResource) SetTableFilters(filters ...FilterDef) *BaseResource {
	b.tableFilters = filters
	return b
}

// SetTableBulkActions sets the bulk actions for BuildTableState.
func (b *BaseResource) SetTableBulkActions(actions ...BulkActionDef) *BaseResource {
	b.tableBulkActions = actions
	return b
}

// SetExportURL enables the export button with the given URL.
func (b *BaseResource) SetExportURL(url string) *BaseResource {
	b.tableExportURL = url
	return b
}

// SetImportURL enables the import button with the given URL.
func (b *BaseResource) SetImportURL(url string) *BaseResource {
	b.tableImportURL = url
	return b
}

// SetHeaderActions sets always-visible action buttons in the table header.
func (b *BaseResource) SetHeaderActions(actions ...HeaderAction) *BaseResource {
	b.tableHeaderActions = actions
	return b
}

// BuildTableState constructs a TableState from the resource's list data.
// Resolution order: ResourceQueryable > ResourceSearchable > ResourceFilterable > List.
func (b *BaseResource) BuildTableState(ctx context.Context, canCreate, canDelete bool) (TableState, error) {
	lq := GetListQuery(ctx)
	activeFilters := GetActiveFilters(ctx)

	items, total, err := b.fetchItems(ctx, lq, activeFilters)
	if err != nil {
		return TableState{}, err
	}

	rows := b.buildRows(items)
	pagination := buildPagination(lq, total)
	search, sortKey, sortDir := extractSortSearch(lq)

	return TableState{
		Title:         b.pluralLabel,
		Slug:          b.slug,
		Columns:       b.tableColumns,
		Rows:          rows,
		CanCreate:     canCreate,
		CanDelete:     canDelete,
		NewURL:        "/" + b.slug + "/create",
		BaseURL:       "/" + b.slug,
		Filters:       b.tableFilters,
		ActiveFilters: activeFilters,
		BulkActions:   b.tableBulkActions,
		HeaderActions: b.tableHeaderActions,
		ExportURL:     b.tableExportURL,
		ImportURL:     b.tableImportURL,
		Pagination:    pagination,
		Search:        search,
		SortKey:       sortKey,
		SortDir:       sortDir,
	}, nil
}

// fetchItems resolves which data-fetching strategy to use based on available interfaces.
func (b *BaseResource) fetchItems(ctx context.Context, lq *ListQuery, activeFilters map[string]string) ([]any, int, error) {
	self := interface{}(b)
	if lq == nil {
		items, err := b.List(ctx)
		return items, len(items), err
	}
	if q, ok := self.(ResourceQueryable); ok {
		return q.ListQuery(ctx, *lq)
	}
	if lq.Search != "" {
		if s, ok := self.(ResourceSearchable); ok {
			items, err := s.Search(ctx, lq.Search)
			return items, len(items), err
		}
	}
	if len(activeFilters) > 0 {
		if f, ok := self.(ResourceFilterable); ok {
			items, err := f.ListFiltered(ctx, activeFilters)
			return items, len(items), err
		}
	}
	items, err := b.List(ctx)
	return items, len(items), err
}

// buildRows converts items to table rows using reflection.
func (b *BaseResource) buildRows(items []any) []Row {
	rows := make([]Row, 0, len(items))
	for _, item := range items {
		row := Row{ID: getItemID(item)}
		for _, col := range b.tableColumns {
			row.Cells = append(row.Cells, getColumnValue(col, item))
		}
		rows = append(rows, row)
	}
	return rows
}

// buildPagination constructs a Pagination struct from ListQuery + total count.
func buildPagination(lq *ListQuery, total int) *Pagination {
	if lq == nil || lq.PerPage <= 0 {
		return nil
	}
	lastPage := (total + lq.PerPage - 1) / lq.PerPage
	if lastPage < 1 {
		lastPage = 1
	}
	return &Pagination{
		CurrentPage: lq.Page,
		PerPage:     lq.PerPage,
		Total:       total,
		LastPage:    lastPage,
	}
}

// extractSortSearch pulls search/sort values from ListQuery for template use.
func extractSortSearch(lq *ListQuery) (search, sortKey, sortDir string) {
	if lq == nil {
		return
	}
	return lq.Search, lq.SortKey, lq.SortDir
}

// Badge returns an empty badge by default.
func (b *BaseResource) Badge(ctx context.Context) string {
	return ""
}

// BadgeColor returns the default badge color.
func (b *BaseResource) BadgeColor(ctx context.Context) string {
	return "primary"
}

// Default permissions - all allowed

func (b *BaseResource) CanCreate(ctx context.Context) bool { return true }
func (b *BaseResource) CanRead(ctx context.Context) bool   { return true }
func (b *BaseResource) CanUpdate(ctx context.Context) bool { return true }
func (b *BaseResource) CanDelete(ctx context.Context) bool { return true }

// Table returns an empty table component. Override in concrete resources.
func (b *BaseResource) Table(ctx context.Context) templ.Component {
	return emptyComponent()
}

// Form returns an empty form component. Override in concrete resources.
func (b *BaseResource) Form(ctx context.Context, item any) templ.Component {
	return emptyComponent()
}

// List returns an empty list by default.
func (b *BaseResource) List(ctx context.Context) ([]any, error) {
	return []any{}, nil
}

// Get returns nil by default.
func (b *BaseResource) Get(ctx context.Context, id string) (any, error) {
	return nil, nil
}

// Create is a no-op by default.
func (b *BaseResource) Create(ctx context.Context, r *http.Request) error {
	return nil
}

// Update is a no-op by default.
func (b *BaseResource) Update(ctx context.Context, id string, r *http.Request) error {
	return nil
}

// Delete is a no-op by default.
func (b *BaseResource) Delete(ctx context.Context, id string) error {
	return nil
}

// BulkDelete is a no-op by default.
func (b *BaseResource) BulkDelete(ctx context.Context, ids []string) error {
	return nil
}

// emptyComponent returns an empty templ component.
func emptyComponent() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return nil
	})
}

// getItemID extracts the ID from an item using reflection.
// Looks for fields named "ID", "Id", or "id" (int or string).
func getItemID(item any) string {
	if item == nil {
		return ""
	}
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Sprintf("%v", item)
	}
	for _, name := range []string{"ID", "Id", "id"} {
		f := v.FieldByName(name)
		if f.IsValid() {
			return fmt.Sprintf("%v", f.Interface())
		}
	}
	return ""
}

// getColumnValue returns the string value of a column field from an item.
func getColumnValue(col Column, item any) string {
	if item == nil {
		return ""
	}
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Sprintf("%v", item)
	}
	f := v.FieldByName(col.Key)
	if !f.IsValid() {
		return ""
	}
	return fmt.Sprintf("%v", f.Interface())
}

// SimpleResource is a minimal resource that only requires metadata.
// Useful for simple resources without custom CRUD logic.
type SimpleResource struct {
	*BaseResource
	listFunc   func(ctx context.Context) ([]any, error)
	getFunc    func(ctx context.Context, id string) (any, error)
	createFunc func(ctx context.Context, r *http.Request) error
	updateFunc func(ctx context.Context, id string, r *http.Request) error
	deleteFunc func(ctx context.Context, id string) error
	tableFunc  func(ctx context.Context) templ.Component
	formFunc   func(ctx context.Context, item any) templ.Component
}

// NewSimpleResource creates a simple resource.
func NewSimpleResource(slug, label, pluralLabel string) *SimpleResource {
	return &SimpleResource{
		BaseResource: NewBaseResource(slug, label, pluralLabel),
	}
}

// WithIcon sets the icon.
func (s *SimpleResource) WithIcon(icon string) *SimpleResource {
	s.BaseResource.SetIcon(icon)
	return s
}

// WithGroup sets the navigation group.
func (s *SimpleResource) WithGroup(group string) *SimpleResource {
	s.BaseResource.SetGroup(group)
	return s
}

// WithSort sets the sort order.
func (s *SimpleResource) WithSort(sort int) *SimpleResource {
	s.BaseResource.SetSort(sort)
	return s
}

// WithList sets the list function.
func (s *SimpleResource) WithList(fn func(ctx context.Context) ([]any, error)) *SimpleResource {
	s.listFunc = fn
	return s
}

// WithGet sets the get function.
func (s *SimpleResource) WithGet(fn func(ctx context.Context, id string) (any, error)) *SimpleResource {
	s.getFunc = fn
	return s
}

// WithCreate sets the create function.
func (s *SimpleResource) WithCreate(fn func(ctx context.Context, r *http.Request) error) *SimpleResource {
	s.createFunc = fn
	return s
}

// WithUpdate sets the update function.
func (s *SimpleResource) WithUpdate(fn func(ctx context.Context, id string, r *http.Request) error) *SimpleResource {
	s.updateFunc = fn
	return s
}

// WithDelete sets the delete function.
func (s *SimpleResource) WithDelete(fn func(ctx context.Context, id string) error) *SimpleResource {
	s.deleteFunc = fn
	return s
}

// WithTable sets the table component function.
func (s *SimpleResource) WithTable(fn func(ctx context.Context) templ.Component) *SimpleResource {
	s.tableFunc = fn
	return s
}

// WithForm sets the form component function.
func (s *SimpleResource) WithForm(fn func(ctx context.Context, item any) templ.Component) *SimpleResource {
	s.formFunc = fn
	return s
}

// CRUD method overrides
func (s *SimpleResource) List(ctx context.Context) ([]any, error) {
	if s.listFunc != nil {
		return s.listFunc(ctx)
	}
	return s.BaseResource.List(ctx)
}

func (s *SimpleResource) Get(ctx context.Context, id string) (any, error) {
	if s.getFunc != nil {
		return s.getFunc(ctx, id)
	}
	return s.BaseResource.Get(ctx, id)
}

func (s *SimpleResource) Create(ctx context.Context, r *http.Request) error {
	if s.createFunc != nil {
		return s.createFunc(ctx, r)
	}
	return s.BaseResource.Create(ctx, r)
}

func (s *SimpleResource) Update(ctx context.Context, id string, r *http.Request) error {
	if s.updateFunc != nil {
		return s.updateFunc(ctx, id, r)
	}
	return s.BaseResource.Update(ctx, id, r)
}

func (s *SimpleResource) Delete(ctx context.Context, id string) error {
	if s.deleteFunc != nil {
		return s.deleteFunc(ctx, id)
	}
	return s.BaseResource.Delete(ctx, id)
}

func (s *SimpleResource) Table(ctx context.Context) templ.Component {
	if s.tableFunc != nil {
		return s.tableFunc(ctx)
	}
	return s.BaseResource.Table(ctx)
}

func (s *SimpleResource) Form(ctx context.Context, item any) templ.Component {
	if s.formFunc != nil {
		return s.formFunc(ctx, item)
	}
	return s.BaseResource.Form(ctx, item)
}
