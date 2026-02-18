package infolist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	il := New()
	assert.NotNil(t, il)
	assert.Empty(t, il.Sections)
}

func TestAddSection(t *testing.T) {
	il := New()
	s := &Section{Heading: "Details", Entries: []*Entry{}}
	il.AddSection(s)

	assert.Len(t, il.Sections, 1)
	assert.Equal(t, "Details", il.Sections[0].Heading)
}

func TestAddSectionDefaultColumns(t *testing.T) {
	il := New()
	s := &Section{Heading: "Test"}
	il.AddSection(s)
	assert.Equal(t, 2, il.Sections[0].Columns)
}

func TestAddSectionChaining(t *testing.T) {
	il := New().
		AddSection(&Section{Heading: "A"}).
		AddSection(&Section{Heading: "B"})
	assert.Len(t, il.Sections, 2)
}

func TestTextEntry(t *testing.T) {
	e := TextEntry("name", "Full Name", "John Doe")
	assert.Equal(t, "name", e.Name)
	assert.Equal(t, "Full Name", e.GetLabel())
	assert.Equal(t, "John Doe", e.GetValueStr())
	assert.Equal(t, EntryTypeText, e.Type)
}

func TestBadgeEntry(t *testing.T) {
	e := BadgeEntry("status", "Status", "active", "green")
	assert.Equal(t, EntryTypeBadge, e.Type)
	assert.Equal(t, "green", e.BadgeColor)
	assert.Equal(t, "active", e.GetValueStr())
}

func TestBooleanEntry(t *testing.T) {
	e := BooleanEntry("active", "Active", true)
	assert.Equal(t, EntryTypeBoolean, e.Type)
	assert.Equal(t, "true", e.GetValueStr())
}

func TestDateEntry(t *testing.T) {
	e := DateEntry("created_at", "Created", "2024-01-15", "")
	assert.Equal(t, EntryTypeDate, e.Type)
	assert.Equal(t, "2006-01-02", e.Format)
}

func TestDateEntryCustomFormat(t *testing.T) {
	e := DateEntry("created_at", "Created", "2024-01-15", "02/01/2006")
	assert.Equal(t, "02/01/2006", e.Format)
}

func TestImageEntry(t *testing.T) {
	e := ImageEntry("avatar", "Avatar", "/uploads/avatar.png")
	assert.Equal(t, EntryTypeImage, e.Type)
	assert.Equal(t, "/uploads/avatar.png", e.GetValueStr())
}

func TestColorEntry(t *testing.T) {
	e := ColorEntry("color", "Color", "#FF5733")
	assert.Equal(t, EntryTypeColor, e.Type)
	assert.Equal(t, "#FF5733", e.GetValueStr())
}

func TestEntryGetValueStrNil(t *testing.T) {
	e := TextEntry("x", "X", nil)
	assert.Equal(t, "", e.GetValueStr())
}

func TestEntryIsVisible(t *testing.T) {
	e := TextEntry("x", "X", "val")
	assert.True(t, e.IsVisible())

	e.Hide(true)
	assert.False(t, e.IsVisible())
}

func TestEntryWithCopy(t *testing.T) {
	e := TextEntry("x", "X", "val").WithCopy()
	assert.True(t, e.IsCopyable)
}

func TestEntryHelp(t *testing.T) {
	e := TextEntry("x", "X", "val").Help("Some help text")
	assert.Equal(t, "Some help text", e.HelpText)
}
