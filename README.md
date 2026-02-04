# SublimeGo Starter

A complete starter kit for building admin panels with Go, powered by [Sublime Admin](https://github.com/bozz33/sublime-admin).

> **Note**: This is the **starter project**. For the reusable framework, see [sublime-admin](https://github.com/bozz33/sublime-admin).

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

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    SUBLIMEGO ECOSYSTEM                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  sublime-admin (Framework)         SublimeGo (This Repo)   │
│  ─────────────────────────         ─────────────────────   │
│  github.com/bozz33/sublime-admin   Starter project         │
│                                                             │
│  • engine/    • form/              • Your resources        │
│  • table/     • auth/              • Your views            │
│  • middleware/• validation/        • Your config           │
│  • ui/        • widget/            • CLI tools             │
│                                                             │
│  go get sublime-admin@v1.0.0       git clone SublimeGo     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Installation

### Option 1: Start a new project (Recommended)

```bash
# Clone this starter
git clone https://github.com/bozz33/SublimeGo.git myproject
cd myproject

# Install dependencies
go mod download

# Initialize (DB, migrations, admin user)
go run cmd/sublimego/main.go init

# Start the server
go run cmd/sublimego/main.go serve
```

### Option 2: Use the framework in an existing project

```bash
# Install the framework
go get github.com/bozz33/sublime-admin@v1.0.0
```

```go
package main

import (
    "github.com/bozz33/sublime-admin/engine"
    "github.com/bozz33/sublime-admin/form"
    "github.com/bozz33/sublime-admin/table"
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

## Updating the Framework

To receive updates from the `sublime-admin` framework:

```bash
go get -u github.com/bozz33/sublime-admin@latest
```

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
