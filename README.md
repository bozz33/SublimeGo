# SublimeGo

A modern, idiomatic Go framework for building admin panels.

[![Go Reference](https://pkg.go.dev/badge/github.com/bozz33/sublimego.svg)](https://pkg.go.dev/github.com/bozz33/sublimego)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Tech Stack

- **Language**: Go 1.24+
- **ORM**: Ent (Facebook) - Type-safe, code-first
- **Templating**: Templ - Type-safe HTML templates
- **UI**: Tailwind CSS v4, Alpine.js, HTMX
- **Database**: SQLite (default), PostgreSQL, MySQL
- **CLI**: Cobra for code generation

## Features

- **Resource System**: Full CRUD with automatic generation
- **Form Builder**: Fluent form builder with validation
- **Table Builder**: Interactive tables with sorting, filters, and pagination
- **Actions**: Customizable actions with confirmation modals
- **Widgets**: Stats cards and charts (ApexCharts)
- **Navigation**: Advanced navigation with groups and badges
- **Authentication**: Built-in auth with bcrypt and sessions
- **Multi-Panel**: Support for multiple admin panels


## Installation

### Quick Start

```bash
# Clone the repository
git clone https://github.com/bozz33/sublimego.git myproject
cd myproject

# Install dependencies
go mod download

# Initialize (DB, migrations, admin user)
go run cmd/sublimego/main.go init

# Start the server
go run cmd/sublimego/main.go serve
```

### Use as a Library

```bash
go get github.com/bozz33/sublimego@latest
```

```go
package main

import (
    "github.com/bozz33/sublimego/engine"
    "github.com/bozz33/sublimego/form"
    "github.com/bozz33/sublimego/table"
)

func main() {
    // Create your admin panel
    panel := engine.NewPanel("admin").
        SetPath("/admin").
        SetBrandName("My App")
    
    // Register resources
    panel.AddResources(
        &ProductResource{},
        &UserResource{},
    )
    
    // Start server
    panel.Serve(":8080")
}
```

Server starts at `http://localhost:8080`

## CLI Commands

### Code Generation

```bash
# Create a new resource
sublimego make:resource Product

# Create a migration
sublimego make:migration create_products_table

# List resources
sublimego resource list
```

### Database

```bash
# Run migrations
sublimego db migrate

# Rollback
sublimego db rollback

# Full reset
sublimego db reset
```

### Server

```bash
# Start in development mode
sublimego serve

# Production mode
sublimego serve --env=production
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Test specific package
go test ./actions/...
```

## Documentation

- **Architecture**: [ARCHITECTURE.md](ARCHITECTURE.md) - Project structure and design patterns
- **Resources**: [RESOURCES_GUIDE.md](RESOURCES_GUIDE.md) - Complete guide to creating resources
- **Templating**: [TEMPLATING.md](TEMPLATING.md) - Working with Templ templates
- **Configuration**: [PANEL_CONFIG.md](PANEL_CONFIG.md) - Panel configuration options

## Project Structure

```
sublimego/
├── actions/            # Action system
├── auth/               # Authentication
├── engine/             # Framework core
├── form/               # Form builder
├── table/              # Table builder
├── widget/             # Dashboard widgets
├── ui/                 # UI components and layouts
├── middleware/         # HTTP middlewares
├── validation/         # Validation rules
├── internal/           # Private packages
│   ├── ent/            # ORM schemas (generated)
│   ├── providers/      # Data providers
│   └── registry/       # Resource registration
├── cmd/                # CLI commands
│   └── sublimego/      # Main CLI entry point
├── views/              # Templ templates
└── config/             # Configuration files
```

## Usage Example

### Complete Resource Implementation

```go
package product

import (
    "context"
    "fmt"
    
    "github.com/bozz33/sublimego/engine"
    "github.com/bozz33/sublimego/form"
    "github.com/bozz33/sublimego/table"
    "github.com/bozz33/sublimego/actions"
)

type ProductResource struct {
    engine.BaseResource
    client *ent.Client
}

func NewProductResource(client *ent.Client) *ProductResource {
    return &ProductResource{client: client}
}

func (r *ProductResource) GetMeta() engine.ResourceMeta {
    return engine.ResourceMeta{
        Name:         "product",
        Label:        "Product",
        PluralLabel:  "Products",
        Icon:         "package",
        Description:  "Manage your product catalog",
    }
}

func (r *ProductResource) GetForm() *form.Form {
    return form.New().SetSchema(
        form.Text("name").Label("Product Name").Required(),
        form.Textarea("description").Label("Description").Rows(5),
        form.Number("price").Label("Price").Required().Min(0),
        form.Select("status").Label("Status").Options([]form.Option{
            {Value: "draft", Label: "Draft"},
            {Value: "published", Label: "Published"},
            {Value: "archived", Label: "Archived"},
        }).Default("draft"),
    )
}

func (r *ProductResource) GetTable() *table.Table {
    return table.New(r.GetData).
        WithColumns(
            table.ID("id"),
            table.Text("name").Sortable().Searchable(),
            table.Currency("price").Sortable(),
            table.Badge("status").Colors(map[string]string{
                "draft":     "gray",
                "published": "green",
                "archived":  "red",
            }),
            table.DateTime("created_at").Sortable(),
        ).
        SetDefaultSort("created_at", "desc").
        SetActions(
            actions.EditAction("/admin/products"),
            actions.DeleteAction("/admin/products"),
        )
}
```

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Development Status

This project is under active development. We're working on:
- Improving test coverage
- Simplifying architecture
- Better documentation
- Performance optimizations

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- Inspired by [Laravel Filament](https://filamentphp.com/)
- Built with [Ent](https://entgo.io/) and [Templ](https://templ.guide/)

---

**Built with care by the SublimeGo community**
