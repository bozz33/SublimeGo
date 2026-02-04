# SublimeGo v0.2 - A Laravel Filament-inspired Admin Framework for Go

Hey r/golang!

A few weeks ago I shared SublimeGo, a framework inspired by Laravel Filament for building admin panels in Go. Thanks to the amazing feedback from this community, I've made significant improvements.

## What Changed Since v0.1

Based on your recommendations:

- **Complete English translation** - All comments, labels, and documentation now in English
- **Removed unnecessary packages** - Cleaned up cache, mail, storage, utils (use standard library or dedicated packages instead)
- **Fixed all tests** - 14 test packages now pass
- **Cleaner architecture** - Removed placeholder code and redundant abstractions
- **Better Go idioms** - Following community best practices

## Tech Stack

- **Go 1.24+** with modules
- **Ent** (Facebook) - Type-safe ORM
- **Templ** - Type-safe HTML templates
- **Tailwind CSS v4** + Alpine.js + HTMX
- **Cobra** - CLI for code generation

## Core Features

```go
// Resource-based CRUD with fluent API
type ProductResource struct {
    engine.BaseResource
}

func (r *ProductResource) Schema() *form.Form {
    return form.New().SetSchema(
        form.Text("name").Label("Name").Required(),
        form.Number("price").Label("Price").Required(),
    )
}

func (r *ProductResource) Table() *table.Table {
    return table.New(r.GetData()).
        WithColumns(
            table.Text("name").Sortable(),
            table.Badge("status"),
        ).
        SetActions(
            actions.EditAction("/products"),
            actions.DeleteAction("/products"),
        )
}
```

## What's Working

- Resource system with automatic CRUD
- Form builder with validation
- Table builder with sorting, filters, pagination
- Dashboard widgets (stats, charts)
- Authentication with sessions
- Rate limiting middleware
- Job queue for async tasks

## What I'm Looking For

1. **Architecture feedback** - Is the package structure idiomatic?
2. **Testing approach** - Currently at ~60% coverage, aiming for 80%+
3. **Missing features** - What would make this useful for your projects?

## Links

- GitHub: https://github.com/bozz33/SublimeGo
- Branch: `main` (stable)

## Honest Assessment

This is a learning project that grew into something more. The codebase follows Go conventions but there's always room for improvement. I'm particularly interested in:

- Whether the fluent API pattern makes sense in Go
- If the middleware stack is properly designed
- Suggestions for the job queue implementation

Thanks for any feedback!

---

**TL;DR**: Admin panel framework for Go, inspired by Laravel Filament. v0.2 addresses community feedback with English translation, cleaner code, and passing tests.
