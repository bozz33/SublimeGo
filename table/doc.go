// Package table provides a fluent API for building data tables.
//
// It supports various column types, sorting, filtering, pagination,
// and row actions. Tables are rendered as Templ components with
// HTMX integration for dynamic updates.
//
// Features:
//   - Fluent column builder API
//   - Multiple column types (text, badge, boolean, date, etc.)
//   - Sortable columns
//   - Filters and search
//   - Pagination
//   - Row actions (edit, delete, custom)
//   - Bulk actions
//   - HTMX integration
//
// Basic usage:
//
//	table := table.New(data).
//		WithColumns(
//			table.Text("name").Label("Name").Sortable(),
//			table.Badge("status").Label("Status"),
//			table.Date("created_at").Label("Created").Format("2006-01-02"),
//		).
//		WithActions(
//			actions.EditAction("/users"),
//			actions.DeleteAction("/users"),
//		).
//		Paginate(25)
//
//	// Render table
//	table.Render(ctx)
package table
