package table

import (
	"fmt"
	"reflect"
)

// TextColumn represents a text column.
type TextColumn struct {
	Key          string
	LabelStr     string
	SortableFlag bool
	SearchFlag   bool
	CopyFlag     bool
}

// Text creates a new text column.
func Text(key string) *TextColumn {
	return &TextColumn{
		Key:      key,
		LabelStr: key,
	}
}

// Label sets the column label.
func (c *TextColumn) Label(label string) *TextColumn {
	c.LabelStr = label
	return c
}

// Sortable makes the column sortable.
func (c *TextColumn) Sortable() *TextColumn {
	c.SortableFlag = true
	return c
}

// Searchable makes the column searchable.
func (c *TextColumn) Searchable() *TextColumn {
	c.SearchFlag = true
	return c
}

// Copyable makes the column copyable.
func (c *TextColumn) Copyable() *TextColumn {
	c.CopyFlag = true
	return c
}

// Column interface implementation
func (c *TextColumn) GetKey() string     { return c.Key }
func (c *TextColumn) GetLabel() string   { return c.LabelStr }
func (c *TextColumn) GetType() string    { return "text" }
func (c *TextColumn) IsSortable() bool   { return c.SortableFlag }
func (c *TextColumn) IsSearchable() bool { return c.SearchFlag }
func (c *TextColumn) IsCopyable() bool   { return c.CopyFlag }
func (c *TextColumn) GetValue(item any) string {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(c.Key)
	if field.IsValid() {
		return fmt.Sprintf("%v", field.Interface())
	}
	return ""
}

// BadgeColumn represents a badge column.
type BadgeColumn struct {
	Key          string
	LabelStr     string
	SortableFlag bool
	ColorMap     map[string]string
}

// Badge creates a new badge column.
func Badge(key string) *BadgeColumn {
	return &BadgeColumn{
		Key:      key,
		LabelStr: key,
		ColorMap: make(map[string]string),
	}
}

// Label sets the column label.
func (c *BadgeColumn) Label(label string) *BadgeColumn {
	c.LabelStr = label
	return c
}

// Sortable makes the column sortable.
func (c *BadgeColumn) Sortable() *BadgeColumn {
	c.SortableFlag = true
	return c
}

// Colors sets the colors by value.
func (c *BadgeColumn) Colors(colors map[string]string) *BadgeColumn {
	c.ColorMap = colors
	return c
}

// GetColor returns the color for a value.
func (c *BadgeColumn) GetColor(value string) string {
	if color, ok := c.ColorMap[value]; ok {
		return color
	}
	return "primary"
}

// Column interface implementation
func (c *BadgeColumn) GetKey() string     { return c.Key }
func (c *BadgeColumn) GetLabel() string   { return c.LabelStr }
func (c *BadgeColumn) GetType() string    { return "badge" }
func (c *BadgeColumn) IsSortable() bool   { return c.SortableFlag }
func (c *BadgeColumn) IsSearchable() bool { return false }
func (c *BadgeColumn) IsCopyable() bool   { return false }
func (c *BadgeColumn) GetValue(item any) string {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(c.Key)
	if field.IsValid() {
		return fmt.Sprintf("%v", field.Interface())
	}
	return ""
}

// ImageColumn represents an image column.
type ImageColumn struct {
	Key      string
	LabelStr string
	Rounded  bool
}

// Image creates a new image column.
func Image(key string) *ImageColumn {
	return &ImageColumn{
		Key:      key,
		LabelStr: key,
	}
}

// Label sets the column label.
func (c *ImageColumn) Label(label string) *ImageColumn {
	c.LabelStr = label
	return c
}

// Round makes the image round.
func (c *ImageColumn) Round() *ImageColumn {
	c.Rounded = true
	return c
}

// Column interface implementation
func (c *ImageColumn) GetKey() string     { return c.Key }
func (c *ImageColumn) GetLabel() string   { return c.LabelStr }
func (c *ImageColumn) GetType() string    { return "image" }
func (c *ImageColumn) IsSortable() bool   { return false }
func (c *ImageColumn) IsSearchable() bool { return false }
func (c *ImageColumn) IsCopyable() bool   { return false }
func (c *ImageColumn) GetValue(item any) string {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(c.Key)
	if field.IsValid() {
		return fmt.Sprintf("%v", field.Interface())
	}
	return ""
}
