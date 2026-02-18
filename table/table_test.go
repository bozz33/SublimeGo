package table

import (
	"testing"
)

func TestNewTable(t *testing.T) {
	data := []any{"item1", "item2"}
	tbl := New(data)

	if tbl == nil {
		t.Error("Expected table to be created")
	}
	if len(tbl.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(tbl.Rows))
	}
	if !tbl.Searchable {
		t.Error("Expected Searchable to be true by default")
	}
	if !tbl.Pagination {
		t.Error("Expected Pagination to be true by default")
	}
	if tbl.PerPage != 15 {
		t.Errorf("Expected PerPage 15, got %d", tbl.PerPage)
	}
}

func TestTableAddColumn(t *testing.T) {
	tbl := New(nil).
		AddColumn("id", "ID").
		AddColumn("name", "Name")

	if len(tbl.Columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(tbl.Columns))
	}
}

func TestTableWithColumns(t *testing.T) {
	col1 := Text("id").WithLabel("ID")
	col2 := Text("name").WithLabel("Name")

	tbl := New(nil).WithColumns(col1, col2)

	if len(tbl.Columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(tbl.Columns))
	}
}

func TestTableSetBaseURL(t *testing.T) {
	tbl := New(nil).SetBaseURL("/users")

	if tbl.BaseURL != "/users" {
		t.Errorf("Expected BaseURL '/users', got '%s'", tbl.BaseURL)
	}
}

func TestTableSearch(t *testing.T) {
	tbl := New(nil).Search(false)

	if tbl.Searchable {
		t.Error("Expected Searchable to be false")
	}
}

func TestTablePaginate(t *testing.T) {
	tbl := New(nil).Paginate(false)

	if tbl.Pagination {
		t.Error("Expected Pagination to be false")
	}
}

func TestTextColumn(t *testing.T) {
	col := Text("name").
		WithLabel("Full Name").
		Sortable().
		Searchable().
		Copyable()

	if col.Key() != "name" {
		t.Errorf("Expected key 'name', got '%s'", col.Key())
	}
	if col.LabelStr != "Full Name" {
		t.Errorf("Expected label 'Full Name', got '%s'", col.LabelStr)
	}
	if !col.SortableFlag {
		t.Error("Expected Sortable to be true")
	}
	if !col.SearchFlag {
		t.Error("Expected Searchable to be true")
	}
	if !col.CopyFlag {
		t.Error("Expected Copyable to be true")
	}
}

func TestTextColumnInterface(t *testing.T) {
	col := Text("test")

	if col.Key() != "test" {
		t.Errorf("Expected Key() 'test', got '%s'", col.Key())
	}
	if col.Type() != "text" {
		t.Errorf("Expected Type() 'text', got '%s'", col.Type())
	}
}

func TestBadgeColumn(t *testing.T) {
	col := Badge("status").
		WithLabel("Status").
		Colors(map[string]string{
			"active":   "success",
			"inactive": "danger",
		})

	if col.Key() != "status" {
		t.Errorf("Expected key 'status', got '%s'", col.Key())
	}
	if col.Type() != "badge" {
		t.Errorf("Expected type 'badge', got '%s'", col.Type())
	}
	if len(col.ColorMap) != 2 {
		t.Errorf("Expected 2 color mappings, got %d", len(col.ColorMap))
	}
}

func TestImageColumn(t *testing.T) {
	col := Image("avatar").
		WithLabel("Profile Picture").
		Round()

	if col.Key() != "avatar" {
		t.Errorf("Expected key 'avatar', got '%s'", col.Key())
	}
	if col.Type() != "image" {
		t.Errorf("Expected type 'image', got '%s'", col.Type())
	}
	if !col.Rounded {
		t.Error("Expected Rounded to be true")
	}
}

func TestSelectFilter(t *testing.T) {
	filter := Select("status").
		WithLabel("Status").
		WithOptions([]FilterOption{
			{Value: "active", Label: "Active"},
			{Value: "inactive", Label: "Inactive"},
		})

	if filter.Key() != "status" {
		t.Errorf("Expected key 'status', got '%s'", filter.Key())
	}
	if filter.Type() != "select" {
		t.Errorf("Expected type 'select', got '%s'", filter.Type())
	}
	if len(filter.FilterOptions()) != 2 {
		t.Errorf("Expected 2 options, got %d", len(filter.FilterOptions()))
	}
}

func TestBooleanFilter(t *testing.T) {
	filter := Boolean("active").WithLabel("Is Active")

	if filter.Key() != "active" {
		t.Errorf("Expected key 'active', got '%s'", filter.Key())
	}
	if filter.Type() != "boolean" {
		t.Errorf("Expected type 'boolean', got '%s'", filter.Type())
	}

	opts := filter.FilterOptions()
	if len(opts) != 2 {
		t.Errorf("Expected 2 options (Yes/No), got %d", len(opts))
	}
}
