# Architecture

This document describes the structure, patterns, and design decisions of the SublimeGo framework.

---

## Guiding Principles

1. **Idiomatic Go**  stdlib first; every dependency must be justified.
2. **Small interfaces**  each contract exposes the minimum required (Interface Segregation Principle).
3. **No magic**  everything is explicit, traceable, and testable.
4. **Single binary**  `//go:embed` for all assets; no external files required at runtime.
5. **Flat structure**  packages at the root level, no unnecessary `pkg/` wrapper.

---

## Package Layout

```
sublimego/
 actions/          # Row actions (edit, delete, custom, bulk)
 appconfig/        # Configuration loading (Viper + validation)
 auth/             # Authentication, sessions, roles, permissions, MFA/TOTP
 cmd/
    sublimego/    # Cobra CLI (serve, init, make:*, db, routes, doctor)
 config/           # YAML configuration files
 engine/           # Framework core: Panel, CRUD handlers, multi-tenancy, relations
 errors/           # Structured errors  package apperrors
 export/           # CSV / Excel export
 flash/            # Session-based flash messages
 form/             # Form builder: fields, layouts, validation
 generator/        # Code generation (embedded .tmpl stubs via //go:embed)
 hooks/            # Render Hooks  named UI injection points
 import/           # CSV / Excel / JSON import
 infolist/         # Read-only detail view (Infolist)
 internal/
    ent/          # Ent ORM schemas + generated code (do not edit manually)
    registry/     # Auto-generated resource registry (provider_gen.go)
    scanner/      # Source scanner for automatic resource discovery
 jobs/             # Background job queue with SQLite persistence
 logger/           # Structured logger (log/slog + lumberjack rotation)
 mailer/           # Email sending
 middleware/       # HTTP middlewares (auth, CORS, CSRF, recovery, throttle)
 notifications/    # Notifications: memory store, DatabaseStore, SSE Broadcaster
 plugin/           # Plugin system (Plugin interface + thread-safe global registry)
 search/           # Global fuzzy search (sahilm/fuzzy)
 table/            # Table builder: columns, filters, summaries, grouping, bulk actions
 ui/
    assets/       # Static assets embedded via //go:embed (CSS, JS)
    atoms/        # Atomic Templ components (badge, button, modal, steps, tabs)
    components/   # Composite Templ components (table, action buttons)
    layouts/      # Layout templates (base, sidebar, topbar, flash, navigation)
 validation/       # Validation (go-playground/validator + gorilla/schema)
 views/            # Application Templ templates (auth, dashboard, generics, widgets)
 widget/           # Dashboard widgets (stats cards, ApexCharts)
 generate.go       # //go:generate directive for the scanner
 go.mod
 Makefile
```

---

## Layer Diagram

```

                     HTTP Request                        

                         

                Middlewares (middleware/)                 
   RequireAuth  RequireRole  CORS  Recovery           
   Throttle  TenantMiddleware  CSRF                    

                         

            engine.Panel  (engine/panel.go)              
   Routing  Auth  Session  Resources  Pages          
   Notifications  Render Hooks  Multi-tenancy          

                                           
    
  Resource       Custom Page       Notifications  
  (CRUD)         (engine/page)     (SSE / DB)     
    
       

                    Builders                             
   form.Form  table.Table  infolist.Infolist           
   actions.Action  widget.Widget                        

       

             Templ Templates  (views/ + ui/)             
   views/generics/  ui/atoms/  ui/layouts/             

       

               Ent ORM  (internal/ent/)                  
   Schemas  Migrations  Type-safe queries              

```

---

## Key Modules

### `engine/`  Framework Core

| File | Role |
|------|------|
| `panel.go` | Panel configuration, route registration, middleware injection |
| `contract.go` | Interfaces: `Resource`, `ResourceCRUD`, `ResourceViews`, `ResourcePermissions`, `ResourceQueryable` |
| `base_resource.go` | Default implementations of the Resource interfaces |
| `crud_handler.go` | Generic HTTP handlers for List, Create, Edit, Delete, BulkDelete |
| `auth_handler.go` | Login / logout handlers |
| `page.go` + `page_handler.go` | Custom non-CRUD pages |
| `relations.go` | `RelationManager`  Nested Resources (BelongsTo, HasMany, ManyToMany) |
| `tenant.go` | Multi-tenancy: `Tenant`, `SubdomainResolver`, `PathResolver`, `MultiPanelRouter`, `TenantAware` |

### `form/`  Form Builder

```go
// Standard fields
form.NewText("name").Label("Name").Required()
form.NewEmail("email").Label("Email")
form.NewPassword("password").Label("Password")
form.NewNumber("price").Label("Price").Min(0)
form.NewTextarea("bio").Label("Bio").Rows(5)
form.NewSelect("status").Label("Status").WithOptions(...)
form.NewCheckbox("active").Label("Active")
form.NewToggle("featured").Label("Featured")
form.NewDatePicker("published_at").Label("Published At")
form.NewFileUpload("avatar").Label("Avatar").AcceptedTypes("image/*")

// Advanced fields
form.NewRichEditor("content").Label("Content").WithToolbar("bold", "italic", "link")
form.NewMarkdownEditor("notes").Label("Notes").Rows(10)
form.NewTagsInput("tags").Label("Tags").WithSuggestions("go", "web", "admin")
form.NewKeyValue("meta").Label("Metadata").WithLabels("Key", "Value")
form.NewColorPicker("color").Label("Color").WithSwatches("#ef4444", "#22c55e")
form.NewSlider("discount").Label("Discount").Range(0, 100).WithUnit("%")
form.NewRepeater("items").Label("Items")

// Layouts
form.NewSection("General").SetSchema(...)
form.NewGrid(2).SetSchema(...)
form.NewTabs().AddTab("General", ...).AddTab("Advanced", ...)
form.NewWizard().AddStep("Step 1", ...).AddStep("Step 2", ...)
form.NewCallout("Note").WithBody("Message").WithColor(form.CalloutInfo)
```

### `table/`  Table Builder

```go
table.New(data).
    WithColumns(
        table.Text("name").WithLabel("Name").WithSortable(true).WithSearchable(true),
        table.Badge("status").WithLabel("Status"),
        table.Boolean("active").WithLabel("Active"),
        table.Date("created_at").WithLabel("Created").WithSortable(true),
        table.Image("avatar").WithLabel("Avatar"),
    ).
    WithFilters(
        table.SelectFilter("status", "Status").WithOptions(...),
        table.BooleanFilter("active", "Active"),
    ).
    WithSummaries(
        table.NewSummary("price", table.SummarySum).WithLabel("Total").WithFormat("$%.2f"),
        table.NewSummary("price", table.SummaryAverage).WithLabel("Average"),
    ).
    WithGroups(
        table.GroupBy("status").WithLabel("By status").Collapsible(),
    ).
    WithBulkActions(
        table.NewBulkAction("delete", "Delete selected").WithColor("danger"),
    )
```

### `auth/`  Authentication

| File | Role |
|------|------|
| `manager.go` | `Manager`  Login, Logout, session, permissions, roles |
| `user.go` | `User` struct  permissions, roles, helpers |
| `mfa.go` | MFA/TOTP RFC 6238  `GenerateSecret`, `Validate`, `GenerateRecoveryCodes`, `UseRecoveryCode` |

### `notifications/`  Notifications

| File | Role |
|------|------|
| `notification.go` | `Notification`, `Level`, `Store` (in-memory + SSE) |
| `database_store.go` | `DatabaseStore`  Ent-backed persistence, implements `NotificationStore` |
| `handler.go` | HTTP handler + `NotificationStore` interface (swappable backend) |
| `broadcast.go` | `Broadcaster`  per-user SSE fan-out, 30s heartbeat, `BroadcastAll` |

### `hooks/`  Render Hooks

```go
// Available positions
hooks.BeforeContent    // Before the main content area
hooks.AfterContent     // After the main content area
hooks.BeforeFormFields // Before form fields
hooks.AfterFormFields  // After form fields
hooks.BeforeTableRows  // Before table rows
hooks.AfterTableRows   // After table rows
hooks.BeforeNavigation // Before the navigation sidebar
hooks.AfterNavigation  // After the navigation sidebar
hooks.InHead           // Inside <head>
hooks.InFooter         // Inside <footer>

// Usage
hooks.Register(hooks.AfterContent, func(ctx context.Context) templ.Component {
    return myBannerComponent()
})
```

### `plugin/`  Plugin System

```go
type MyPlugin struct{}

func (p *MyPlugin) Name() string { return "my-plugin" }
func (p *MyPlugin) Boot() error  { /* initialization */ return nil }

func init() {
    plugin.Register(&MyPlugin{})
}

// At panel startup
if err := plugin.Boot(); err != nil {
    log.Fatal(err)
}
```

---

## SQLite Drivers  Build Tags

The project supports two SQLite drivers selected at compile time:

| Build tag | Driver | Use case |
|-----------|--------|----------|
| `cgo` (default) | `github.com/mattn/go-sqlite3` | `make build`  maximum performance |
| `!cgo` | `modernc.org/sqlite` | `make build-all`  pure Go, cross-compilation |

```bash
make build      # CGO_ENABLED=1   mattn/go-sqlite3
make build-all  # CGO_ENABLED=0   modernc.org/sqlite (5 targets)
```

This is intentional and follows the standard Go pattern for optional CGO dependencies.

---

## Error Handling

The `errors/` package (imported as `apperrors`) provides structured, HTTP-aware errors:

```go
import apperrors "github.com/bozz33/sublimego/errors"

err := apperrors.NotFound("user not found")
err := apperrors.Forbidden("access denied")
err := apperrors.Validation("email", "invalid format")

// In a handler
apperrors.Handle(w, r, err)
```

---

## Configuration

Configuration is loaded by `appconfig/loader.go` via Viper.
Priority order: environment variables > `config.yaml` > defaults.

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 15s
  write_timeout: 15s

database:
  driver: "sqlite"
  url: "file:sublimego.db?cache=shared&_fk=1"

engine:
  base_path: "/admin"
  brand_name: "SublimeGo Admin"
  items_per_page: 25

auth:
  session_lifetime: 24h
  cookie_name: "sublimego_session"
  bcrypt_cost: 12
```

---

## Design Patterns

### Builder Pattern
Used throughout `form`, `table`, `infolist`, and `widget` for a fluent, readable API.

### Interface Segregation
`Resource` is composed of small focused interfaces (`ResourceMeta`, `ResourceCRUD`,
`ResourceViews`, `ResourcePermissions`, `ResourceNavigation`)  each implementable independently.

### Dependency Injection
`Panel` accepts its dependencies (DB, Auth, Session, Notifications) via `With*` setters.

### Template Method
`BaseResource` provides default implementations that concrete resources override as needed.

### Registry Pattern
`internal/registry/provider_gen.go` is auto-generated by the scanner (`cmd/scanner`).
Resources are discovered at build time  no manual registration required.

---

## Testing

```bash
go test ./...           # All tests
go test -cover ./...    # With coverage report
go test -race ./...     # Race condition detection
go test ./engine/... -v # Specific package, verbose output
```

---

## Further Reading

- [RESOURCES_GUIDE.md](RESOURCES_GUIDE.md)  Building resources step by step
- [PANEL_CONFIG.md](PANEL_CONFIG.md)  Full configuration reference
- [TEMPLATING.md](TEMPLATING.md)  Working with Templ templates
- [CONTRIBUTING.md](CONTRIBUTING.md)  Contribution guidelines
