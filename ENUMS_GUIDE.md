# Enums Guide

> Complete reference for SublimeGo typed enumerations — labels, colors, icons, and integration with forms and tables.

---

## Overview

SublimeGo enums are plain Go types that implement one or more interfaces:

| Interface | Method | Used for |
|-----------|--------|----------|
| `HasLabel` | `Label() string` | Human-readable display name |
| `HasColor` | `Color() string` | Tailwind color name (badge, table cell) |
| `HasIcon` | `Icon() string` | Material Icons name |
| `HasDescription` | `Description() string` | Longer tooltip/help text |

---

## Quick Start

```bash
# Generate a new enum
sublimego make:enum Status
```

This creates `internal/enums/status.go` with a ready-to-use scaffold.

---

## Defining an Enum

```go
package enums

type Status int

const (
    StatusDraft     Status = iota
    StatusPublished Status = iota
    StatusArchived  Status = iota
)

func (s Status) Label() string {
    switch s {
    case StatusDraft:     return "Draft"
    case StatusPublished: return "Published"
    case StatusArchived:  return "Archived"
    }
    return "Unknown"
}

func (s Status) Color() string {
    switch s {
    case StatusDraft:     return "gray"
    case StatusPublished: return "green"
    case StatusArchived:  return "yellow"
    }
    return "gray"
}

func (s Status) Icon() string {
    switch s {
    case StatusDraft:     return "edit_note"
    case StatusPublished: return "check_circle"
    case StatusArchived:  return "archive"
    }
    return "help"
}

func (s Status) String() string { return s.Label() }

// AllStatusValues is used with enum helpers.
var AllStatusValues = []Status{StatusDraft, StatusPublished, StatusArchived}
```

---

## Generic Helpers

All helpers are in the `enum` package and use Go generics.

### `enum.Options` — for form selects

```go
import "github.com/bozz33/sublimego/enum"

// Returns []form.SelectOption{Label, Value}
opts := enum.Options(AllStatusValues)

// Use in a form field:
form.Select("status").Label("Status").Options(opts)
```

### `enum.Labels` — map of value → label

```go
labels := enum.Labels(AllStatusValues)
// map["Draft":"Draft", "Published":"Published", ...]
```

### `enum.Colors` — map of value → Tailwind color

```go
colors := enum.Colors(AllStatusValues)
// map["Draft":"gray", "Published":"green", ...]

// Use in a table badge column:
table.Badge("status").Colors(colors)
```

### `enum.Icons` — map of value → Material Icon

```go
icons := enum.Icons(AllStatusValues)
// map["Draft":"edit_note", "Published":"check_circle", ...]
```

### `enum.BadgeColor` — single value lookup with fallback

```go
color := enum.BadgeColor(AllStatusValues, "Published", "gray")
// returns "green"

color = enum.BadgeColor(AllStatusValues, "unknown", "gray")
// returns "gray" (fallback)
```

### `enum.OptionsFromStringer` — when enum has `String()` method

```go
opts := enum.OptionsFromStringer(AllStatusValues)
// Same as Options but uses String() for the Value field
```

---

## Integration with Forms

```go
func (r *PostResource) Form(ctx context.Context, item any) []form.Field {
    return []form.Field{
        form.Text("title").Label("Title").Required(),
        form.Select("status").
            Label("Status").
            Options(enum.Options(enums.AllStatusValues)),
    }
}
```

---

## Integration with Tables

```go
func (r *PostResource) Columns() []table.Column {
    return []table.Column{
        table.Text("title").Label("Title").Sortable(),
        table.Badge("status").
            Label("Status").
            Colors(enum.Colors(enums.AllStatusValues)),
    }
}
```

---

## Integration with Infolist

```go
func (r *PostResource) View(ctx context.Context, item any) templ.Component {
    il := infolist.New().AddSection(&infolist.Section{
        Entries: []*infolist.Entry{
            infolist.TextEntry("title", "Title", getField(item, "Title")),
            infolist.BadgeEntry(
                "status", "Status",
                getField(item, "Status"),
                enum.Colors(enums.AllStatusValues)[fmt.Sprintf("%v", getField(item, "Status"))],
            ),
        },
    })
    return generics.Infolist(il)
}
```

---

## Available Tailwind Colors

These color names are recognized by SublimeGo badge rendering:

| Color | Badge classes |
|-------|--------------|
| `green` | `bg-green-100 text-green-700` |
| `red` | `bg-red-100 text-red-700` |
| `yellow` | `bg-yellow-100 text-yellow-700` |
| `blue` | `bg-blue-100 text-blue-700` |
| `purple` | `bg-purple-100 text-purple-700` |
| `orange` | `bg-orange-100 text-orange-700` |
| `gray` | `bg-gray-100 text-gray-700` |
| `indigo` | `bg-indigo-100 text-indigo-700` |
| `teal` | `bg-teal-100 text-teal-700` |
| `pink` | `bg-pink-100 text-pink-700` |

---

## Best Practices

1. **Always define `String()`** — required for `OptionsFromStringer`, `Labels`, `Colors`, `Icons`, `BadgeColor`.
2. **Export `AllXxxValues`** — makes it easy to pass to helpers without repeating the list.
3. **Keep enums in `internal/enums/`** — one file per enum type.
4. **Use `iota`** — for integer-backed enums stored in the database.
5. **Match DB values** — if your DB stores `"draft"`, make `String()` return `"draft"` (lowercase).
