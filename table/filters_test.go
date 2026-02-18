package table

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectFilterDefaultLabel(t *testing.T) {
	f := Select("role")
	assert.Equal(t, "role", f.Label())
}

func TestBooleanFilterDefaultLabel(t *testing.T) {
	f := Boolean("published")
	assert.Equal(t, "published", f.Label())
}

func TestTableWithFiltersIntegration(t *testing.T) {
	tbl := New([]any{}).
		WithFilters(
			Select("status").WithLabel("Status"),
			Boolean("active").WithLabel("Active"),
		)

	assert.Len(t, tbl.Filters, 2)
	assert.Equal(t, "select", tbl.Filters[0].Type())
	assert.Equal(t, "boolean", tbl.Filters[1].Type())
}
