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

// SetActions sets the available actions.
func (t *Table) SetActions(acts ...*actions.Action) *Table {
	t.Actions = append(t.Actions, acts...)
	return t
}

// SetBaseURL sets the base URL.
func (t *Table) SetBaseURL(url string) *Table {
	t.BaseURL = url
	return t
}

// AddColumn adds a simple column.
func (t *Table) AddColumn(key, label string) *Table {
	col := Text(key).Label(label)
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
	GetKey() string
	GetLabel() string
	GetType() string
	IsSortable() bool
	IsSearchable() bool
	IsCopyable() bool
	GetValue(item any) string
}

// Action represents an action on a row.
type Action interface {
	GetLabel() string
	GetIcon() string
	GetColor() string
	GetURL(item any) string
	IsVisible(ctx context.Context, item any) bool
}

// Filter represents a table filter.
type Filter interface {
	GetKey() string
	GetLabel() string
	GetType() string
	GetOptions() []FilterOption
}

// FilterOption represents a filter option.
type FilterOption struct {
	Value string
	Label string
}
