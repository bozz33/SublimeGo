package table

import (
	"context"

	"github.com/bozz33/sublimego/actions"
)

// Table represents the complete configuration of a table.
type Table struct {
	Columns     []Column
	Rows        []any
	Actions     []*actions.Action
	BulkActions []*BulkAction
	Filters     []Filter
	Summaries   []Summary
	Groups      []Grouping
	Searchable  bool
	Pagination  bool
	PerPage     int
	BaseURL     string
}

// New creates a new Table instance.
func New(data []any) *Table {
	return &Table{
		Columns:     make([]Column, 0),
		Rows:        data,
		Actions:     make([]*actions.Action, 0),
		BulkActions: make([]*BulkAction, 0),
		Filters:     make([]Filter, 0),
		Searchable:  true,
		Pagination:  true,
		PerPage:     15,
	}
}

// WithColumns sets the table columns.
func (t *Table) WithColumns(cols ...Column) *Table {
	t.Columns = append(t.Columns, cols...)
	return t
}

// WithActions sets the available actions.
func (t *Table) WithActions(acts ...*actions.Action) *Table {
	t.Actions = append(t.Actions, acts...)
	return t
}

// WithBaseURL sets the base URL.
func (t *Table) WithBaseURL(url string) *Table {
	t.BaseURL = url
	return t
}

// AddColumn adds a simple column.
func (t *Table) AddColumn(key, label string) *Table {
	col := Text(key).WithLabel(label)
	t.Columns = append(t.Columns, col)
	return t
}

// WithBulkActions sets the bulk actions available on selected rows.
func (t *Table) WithBulkActions(actions ...*BulkAction) *Table {
	t.BulkActions = append(t.BulkActions, actions...)
	return t
}

// WithFilters sets the available filters.
func (t *Table) WithFilters(filters ...Filter) *Table {
	t.Filters = append(t.Filters, filters...)
	return t
}

// WithSummaries sets the column summaries shown in the table footer.
func (t *Table) WithSummaries(summaries ...Summary) *Table {
	t.Summaries = append(t.Summaries, summaries...)
	return t
}

// WithGroups sets the grouping definitions for the table.
func (t *Table) WithGroups(groups ...Grouping) *Table {
	t.Groups = append(t.Groups, groups...)
	return t
}

// Search enables/disables search.
func (t *Table) Search(enabled bool) *Table {
	t.Searchable = enabled
	return t
}

// Paginate enables/disables pagination.
func (t *Table) Paginate(enabled bool) *Table {
	t.Pagination = enabled
	return t
}

// Column is the common interface for all columns.
type Column interface {
	Key() string
	Label() string
	Type() string
	IsSortable() bool
	IsSearchable() bool
	IsCopyable() bool
	Value(item any) string
}

// Action represents an action on a row.
type Action interface {
	Label() string
	Icon() string
	Color() string
	URL(item any) string
	IsVisible(ctx context.Context, item any) bool
}

// Filter represents a table filter.
type Filter interface {
	Key() string
	Label() string
	Type() string
	FilterOptions() []FilterOption
}

// FilterOption represents a filter option.
type FilterOption struct {
	Value string
	Label string
}

// SummaryType defines the aggregation function for a column summary.
type SummaryType string

const (
	SummarySum     SummaryType = "sum"
	SummaryAverage SummaryType = "average"
	SummaryMin     SummaryType = "min"
	SummaryMax     SummaryType = "max"
	SummaryCount   SummaryType = "count"
)

// Summary defines a column-level aggregation shown in the table footer.
type Summary interface {
	ColumnKey() string
	Label() string
	Type() SummaryType
	Format() string
	Compute(rows []any) string
}

// SummaryDef is the concrete implementation of Summary.
type SummaryDef struct {
	colKey    string
	labelStr  string
	sumType   SummaryType
	formatStr string
	computeFn func(rows []any) string
}

// NewSummary creates a SummaryDef for a column.
func NewSummary(columnKey string, t SummaryType) *SummaryDef {
	return &SummaryDef{
		colKey:  columnKey,
		sumType: t,
	}
}

// WithLabel sets the summary label.
func (s *SummaryDef) WithLabel(label string) *SummaryDef {
	s.labelStr = label
	return s
}

// WithFormat sets a printf-style format string (e.g. "%.2f").
func (s *SummaryDef) WithFormat(format string) *SummaryDef {
	s.formatStr = format
	return s
}

// WithCompute sets a custom compute function.
func (s *SummaryDef) WithCompute(fn func(rows []any) string) *SummaryDef {
	s.computeFn = fn
	return s
}

func (s *SummaryDef) ColumnKey() string { return s.colKey }
func (s *SummaryDef) Label() string     { return s.labelStr }
func (s *SummaryDef) Type() SummaryType { return s.sumType }
func (s *SummaryDef) Format() string    { return s.formatStr }
func (s *SummaryDef) Compute(rows []any) string {
	if s.computeFn != nil {
		return s.computeFn(rows)
	}
	return ""
}

// ---------------------------------------------------------------------------
// Grouping — group rows by a column value (like Filament's ->groups()).
// ---------------------------------------------------------------------------

// Grouping defines how table rows are grouped by a column value.
type Grouping struct {
	columnKey   string
	labelStr    string
	collapsible bool
	collapsed   bool
	titleFn     func(value string) string // optional custom group title
}

// GroupBy creates a new Grouping for the given column key.
func GroupBy(columnKey string) *Grouping {
	return &Grouping{
		columnKey: columnKey,
		labelStr:  columnKey,
	}
}

// WithLabel sets the display label for the grouping selector.
func (g *Grouping) WithLabel(label string) *Grouping {
	g.labelStr = label
	return g
}

// Collapsible makes the groups collapsible.
func (g *Grouping) Collapsible() *Grouping {
	g.collapsible = true
	return g
}

// CollapsedByDefault makes groups collapsed by default.
func (g *Grouping) CollapsedByDefault() *Grouping {
	g.collapsible = true
	g.collapsed = true
	return g
}

// WithTitleFn sets a custom function to format the group title from the raw value.
func (g *Grouping) WithTitleFn(fn func(value string) string) *Grouping {
	g.titleFn = fn
	return g
}

// ColumnKey returns the column key used for grouping.
func (g *Grouping) ColumnKey() string { return g.columnKey }

// Label returns the display label.
func (g *Grouping) Label() string { return g.labelStr }

// IsCollapsible returns true if groups can be collapsed.
func (g *Grouping) IsCollapsible() bool { return g.collapsible }

// IsCollapsed returns true if groups are collapsed by default.
func (g *Grouping) IsCollapsed() bool { return g.collapsed }

// Title returns the formatted group title for a raw column value.
func (g *Grouping) Title(value string) string {
	if g.titleFn != nil {
		return g.titleFn(value)
	}
	return value
}

// GroupedRow represents a group of rows sharing the same column value.
type GroupedRow struct {
	Key         string // raw column value
	Title       string // formatted display title
	Rows        []any  // rows in this group
	Collapsible bool
	Collapsed   bool
}

// GroupRows groups the given rows by the specified column using the Column interface.
func GroupRows(rows []any, col Column, g *Grouping) []GroupedRow {
	order := make([]string, 0)
	groups := make(map[string][]any)

	for _, row := range rows {
		key := col.Value(row)
		if _, exists := groups[key]; !exists {
			order = append(order, key)
		}
		groups[key] = append(groups[key], row)
	}

	result := make([]GroupedRow, 0, len(order))
	for _, key := range order {
		result = append(result, GroupedRow{
			Key:         key,
			Title:       g.Title(key),
			Rows:        groups[key],
			Collapsible: g.IsCollapsible(),
			Collapsed:   g.IsCollapsed(),
		})
	}
	return result
}
