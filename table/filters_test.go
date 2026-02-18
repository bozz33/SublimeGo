package table

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectFilterDefaultLabel(t *testing.T) {
	f := Select("role")
	assert.Equal(t, "role", f.GetLabel())
}

func TestBooleanFilterDefaultLabel(t *testing.T) {
	f := Boolean("published")
	assert.Equal(t, "published", f.GetLabel())
}

func TestTableWithFiltersIntegration(t *testing.T) {
	tbl := New([]any{}).
		WithFilters(
			Select("status").Label("Status"),
			Boolean("active").Label("Active"),
		)

	assert.Len(t, tbl.Filters, 2)
	assert.Equal(t, "select", tbl.Filters[0].GetType())
	assert.Equal(t, "boolean", tbl.Filters[1].GetType())
}
