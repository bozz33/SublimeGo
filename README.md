# SublimeGo

**A modern, idiomatic Go framework for building admin panels.**
Inspired by Laravel Filament, written entirely in idiomatic Go.

[![Go Reference](https://pkg.go.dev/badge/github.com/bozz33/sublimego.svg)](https://pkg.go.dev/github.com/bozz33/sublimego)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Build](https://github.com/bozz33/sublimego/actions/workflows/ci.yml/badge.svg)](https://github.com/bozz33/sublimego/actions)

---

## Two Repositories

SublimeGo is split into two complementary repositories:

| | [SublimeGo](https://github.com/bozz33/SublimeGo) | [sublime-admin](https://github.com/bozz33/sublime-admin) |
|---|---|---|
| **Type** | Complete starter project | Core framework library |
| **Use when** | Starting a new admin panel project | Adding admin panel to an existing Go project |
| **Includes** | DB setup, Ent schemas, CLI, examples, views | Framework packages only (engine, form, table, auth…) |
| **Import** | `github.com/bozz33/sublimego` | `github.com/bozz33/sublimeadmin` |

```bash
# Use SublimeGo — clone and start building
git clone https://github.com/bozz33/SublimeGo.git myproject

# Use sublime-admin — add to an existing project
go get github.com/bozz33/sublimeadmin@latest
```

> **Not sure which to pick?** Start with **SublimeGo** — it includes everything you need out of the box.
> Switch to **sublime-admin** when you need to integrate the framework into an existing codebase.

---

## Tech Stack

| Component    | Technology                                      |
|--------------|-------------------------------------------------|
| Language     | Go 1.24+                                        |
| ORM          | [Ent](https://entgo.io/)  type-safe, code-first |
| Templates    | [Templ](https://templ.guide/)  compiled to Go  |
| UI           | Tailwind CSS v4  Alpine.js  HTMX              |
| Database     | SQLite (default)  PostgreSQL  MySQL           |
| Sessions     | [SCS](https://github.com/alexedwards/scs)       |
| CLI          | [Cobra](https://github.com/spf13/cobra)         |

---

## Features

### Forms
- Fields: Text, Email, Password, Number, Textarea, Select, Checkbox, Toggle, DatePicker, FileUpload
- Advanced fields: **RichEditor**, **MarkdownEditor**, **TagsInput**, **KeyValue**, **ColorPicker**, **Slider**
- Layouts: Section, Grid, Tabs, **Wizard/Steps**, Callout, Repeater

### Tables
- Columns: Text, Badge, Boolean, Date, Image
- Sorting, search, pagination, filters, bulk actions
- **Summaries** (sum, average, min, max, count) in the table footer
- **Grouping**  group rows by column value with optional collapse
- Export CSV/Excel, Import CSV/Excel/JSON

### Authentication & Security
- Bcrypt password hashing + secure sessions
- Middleware: RequireAuth, RequireRole, RequirePermission, login throttling
- **MFA/TOTP**  two-factor authentication (RFC 6238) + recovery codes
- Role and permission management

### Notifications
- In-memory store (development)
- **DatabaseStore**  full Ent-backed persistence
- SSE (Server-Sent Events) for real-time delivery
- **Broadcaster**  per-user SSE fan-out with 30s heartbeat and `BroadcastAll`

### Advanced Architecture
- **Multi-tenancy**  `SubdomainResolver`, `PathResolver`, `MultiPanelRouter`, `TenantAware`
- **Render Hooks**  10 named UI injection points (`BeforeContent`, `AfterFormFields`, `InHead`, )
- **Plugin system**  `Plugin` interface with `Boot()` and a thread-safe global registry
- **Nested Resources**  `RelationManager` (BelongsTo, HasMany, ManyToMany)
- Background jobs with SQLite persistence
- Structured logger (`log/slog`) with log rotation (lumberjack)
- Built-in mailer

### Deployment
- **Single binary**  Tailwind assets, templates and stubs embedded via `//go:embed`
- Cross-compilation: `make build-all`  Linux/macOS/Windows  amd64/arm64

---

## Quick Start

```bash
# Clone the repository
git clone https://github.com/bozz33/sublimego.git myproject
cd myproject

# Download dependencies
go mod download

# Download Tailwind CSS standalone CLI (once)
make tailwind-download

# Initialize (DB + migrations + admin user)
go run cmd/sublimego/main.go init

# Start in development mode (hot reload)
make dev
```

The admin panel is available at `http://localhost:8080/admin`.

---

## Use as a Library

```bash
go get github.com/bozz33/sublimego@latest
```

```go
package main

import (
    "net/http"

    "github.com/bozz33/sublimego/engine"
)

func main() {
    panel := engine.NewPanel("admin").
        WithPath("/admin").
        WithBrandName("My App").
        WithDatabase(db)

    panel.AddResources(
        NewProductResource(db),
        NewUserResource(db),
    )

    http.ListenAndServe(":8080", panel.Router())
}
```

---

## Resource Example

```go
type ProductResource struct {
    engine.BaseResource
    db *ent.Client
}

func (r *ProductResource) Slug() string        { return "products" }
func (r *ProductResource) Label() string       { return "Product" }
func (r *ProductResource) PluralLabel() string { return "Products" }
func (r *ProductResource) Icon() string        { return "package" }

func (r *ProductResource) Form(ctx context.Context, item any) templ.Component {
    f := form.New().SetSchema(
        form.NewText("name").Label("Name").Required(),
        form.NewRichEditor("description").Label("Description"),
        form.NewNumber("price").Label("Price").Required(),
        form.NewTagsInput("tags").Label("Tags").WithSuggestions("sale", "new", "featured"),
        form.NewSelect("status").Label("Status").WithOptions(
            form.Option{Value: "draft",     Label: "Draft"},
            form.Option{Value: "published", Label: "Published"},
        ),
        form.NewColorPicker("color").Label("Color"),
    )
    return views.GenericForm(f)
}

func (r *ProductResource) Table(ctx context.Context) templ.Component {
    t := table.New(nil).
        WithColumns(
            table.Text("name").WithLabel("Name").WithSortable(true).WithSearchable(true),
            table.Badge("status").WithLabel("Status"),
            table.Date("created_at").WithLabel("Created").WithSortable(true),
        ).
        WithSummaries(
            table.NewSummary("price", table.SummarySum).WithLabel("Total").WithFormat("$%.2f"),
        ).
        WithGroups(
            table.GroupBy("status").WithLabel("By status").Collapsible(),
        )
    return views.GenericTable(t)
}
```

---

## CLI Commands

```bash
# Code generation
sublimego make:resource Product
sublimego make:page Dashboard

# Database
sublimego db migrate
sublimego db rollback
sublimego db reset

# Server
sublimego serve
sublimego serve --env=production

# Utilities
sublimego doctor
sublimego routes
sublimego version
```

---

## Make Targets

```bash
make dev            # Development with hot reload (air + templ watch)
make build          # Production binary (CGO enabled)
make build-all      # Cross-compile for 5 targets (Linux/macOS/Windows)
make generate       # Regenerate Templ templates
make css            # Compile Tailwind CSS (production, minified)
make css-watch      # Compile Tailwind CSS in watch mode
make test           # Run all tests
make lint           # Run golangci-lint
make clean          # Remove bin/ tmp/ tools/
make install-tools  # Install air, templ, golangci-lint
```

---

## Project Structure

```
sublimego/
 actions/          # Row actions (edit, delete, custom, bulk)
 appconfig/        # Configuration loading (Viper)
 auth/             # Authentication, sessions, MFA/TOTP
 cmd/
    sublimego/    # Cobra CLI (serve, init, make:*, db, routes, doctor)
 engine/           # Framework core (Panel, CRUD, multi-tenancy, relations)
 errors/           # Structured errors  package apperrors
 export/           # CSV / Excel export
 flash/            # Flash messages
 form/             # Form builder (fields, layouts, validation)
 generator/        # Code generation (embedded .tmpl stubs)
 hooks/            # Render Hooks  named UI injection points
 import/           # CSV / Excel / JSON import
 infolist/         # Read-only detail view (Infolist)
 internal/
    ent/          # Ent ORM schemas + generated code
    registry/     # Auto-generated resource registry
    scanner/      # Source scanner for code generation
 jobs/             # Background job queue (SQLite)
 logger/           # Structured logger (slog + lumberjack)
 mailer/           # Email sending
 middleware/       # HTTP middlewares (auth, CORS, CSRF, recovery, throttle)
 notifications/    # Notifications (memory store, DatabaseStore, SSE Broadcaster)
 plugin/           # Plugin system
 search/           # Global fuzzy search
 table/            # Table builder (columns, filters, summaries, grouping)
 ui/
    assets/       # Static assets embedded via //go:embed
    atoms/        # Atomic Templ components (badge, button, modal, steps, tabs)
    components/   # Composite Templ components
    layouts/      # Layouts (base, sidebar, topbar, flash)
 validation/       # Validation (go-playground/validator + gorilla/schema)
 views/            # Application Templ templates (auth, dashboard, generics, widgets)
 widget/           # Dashboard widgets (stats cards, ApexCharts)
 generate.go       # //go:generate directive
 go.mod
 Makefile
```

---

## Testing

```bash
go test ./...              # All tests
go test -cover ./...       # With coverage
go test -race ./...        # Race condition detection
go test ./engine/... -v    # Specific package, verbose
```

---

## Documentation

| Document | Description |
|----------|-------------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | Project structure, patterns, design decisions |
| [RESOURCES_GUIDE.md](RESOURCES_GUIDE.md) | Complete guide to building resources |
| [PANEL_CONFIG.md](PANEL_CONFIG.md) | Panel configuration reference |
| [TEMPLATING.md](TEMPLATING.md) | Working with Templ templates |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contribution guidelines |

---

## License

MIT  see [LICENSE](LICENSE).

---

*Inspired by [Laravel Filament](https://filamentphp.com/)  Built with [Ent](https://entgo.io/) and [Templ](https://templ.guide/)*

