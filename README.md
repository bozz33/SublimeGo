# SublimeGo Framework

SublimeGo is a modern Go framework inspired by Laravel Filament, designed to accelerate web application development with a clean, maintainable architecture.

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

```bash
# Install SublimeGo in your Go project
go get github.com/bozz33/SublimeGo@v1.0.0
```

## Quick Start

### Option 1: Use as a library in your existing project

```go
package main

import (
    "github.com/bozz33/SublimeGo/pkg/engine"
    "github.com/bozz33/SublimeGo/pkg/form"
    "github.com/bozz33/SublimeGo/pkg/table"
)

func main() {
    // Create your admin panel
    panel := engine.NewPanel("admin", "/admin")
    
    // Register resources
    panel.RegisterResources(
        &ProductResource{},
        &UserResource{},
    )
    
    // Start server
    panel.Serve(":8080")
}
```

### Option 2: Clone and customize

```bash
# Clone the repository
git clone https://github.com/bozz33/SublimeGo.git
cd SublimeGo

# Install dependencies
go mod download

# Initialize project (DB, migrations, admin user)
go run cmd/sublimego/main.go init

# Start the server
go run cmd/sublimego/main.go serve
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
go test ./pkg/actions/...
```

## Documentation

- **Architecture**: [ARCHITECTURE.md](docs/ARCHITECTURE.md)
- **Resources**: [RESOURCES_GUIDE.md](docs/RESOURCES_GUIDE.md)
- **Templating**: [TEMPLATING.md](docs/TEMPLATING.md)
- **Configuration**: [PANEL_CONFIG.md](docs/PANEL_CONFIG.md)

## Project Structure

```
SublimeGo/
├── cmd/sublimego/          # CLI entry point
├── internal/
│   ├── ent/                # ORM schemas and entities
│   ├── resources/          # CRUD resources
│   ├── providers/          # Data providers
│   └── registry/           # Resource registration
├── pkg/
│   ├── actions/            # Action system
│   ├── auth/               # Authentication
│   ├── engine/             # Framework core
│   ├── form/               # Form builder
│   ├── table/              # Table builder
│   ├── widget/             # Dashboard widgets
│   └── ui/                 # Templates and layouts
├── views/                  # Templ templates
└── public/                 # Static assets
```

## Usage Example

### Creating a Resource

```go
package product

import (
    "github.com/bozz33/SublimeGo/pkg/engine"
    "github.com/bozz33/SublimeGo/pkg/form"
    "github.com/bozz33/SublimeGo/pkg/table"
)

type ProductResource struct {
    engine.BaseResource
}

func (r *ProductResource) Schema() *form.Form {
    return form.New().SetSchema(
        form.Text("name").Label("Name").Required(),
        form.Textarea("description").Label("Description"),
        form.Number("price").Label("Price").Required(),
    )
}

func (r *ProductResource) Table() *table.Table {
    return table.New(r.GetData()).
        WithColumns(
            table.Text("name").Label("Name").Sortable(),
            table.Badge("status").Label("Status"),
            table.Text("price").Label("Price"),
        ).
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
