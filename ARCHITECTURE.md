# Architecture

This document describes the architecture and structure of SublimeGo.

## Project Structure

Following [Go's official module layout](https://go.dev/doc/modules/layout), packages are at the root level for easy importing:

```
sublimego/
├── actions/              # Action system (edit, delete, custom)
├── auth/                 # Authentication & authorization
├── appconfig/            # Configuration management
├── engine/               # Core panel engine
├── errors/               # Error handling
├── export/               # Data export (CSV, Excel, PDF)
├── flash/                # Flash messages
├── form/                 # Form builder
├── generator/            # Code generation utilities
├── jobs/                 # Background job queue
├── logger/               # Logging utilities
├── middleware/           # HTTP middlewares
├── registry/             # Resource registry
├── scanner/              # Code scanner utilities
├── table/                # Table builder
├── ui/                   # UI components
│   ├── atoms/            # Basic UI elements
│   ├── components/       # Complex components
│   ├── icons/            # Icon system
│   └── layouts/          # Layout templates
├── validation/           # Validation rules
├── widget/               # Dashboard widgets
├── internal/             # Private packages (not importable externally)
│   ├── ent/              # Ent ORM schema and generated code
│   ├── providers/        # Application-specific providers
│   └── registry/         # Resource registration
├── cmd/                  # CLI commands
│   ├── scanner/          # Code scanner for resource generation
│   └── sublimego/        # Main CLI application
├── views/                # Application views (Templ templates)
├── config/               # Configuration files
├── go.mod
└── README.md
```

## Core Components

### 1. Engine (`engine/`)

The engine is the heart of SublimeGo. It manages:
- **Panel**: Main admin panel with routing and middleware
- **Resources**: CRUD resource management
- **Authentication**: User authentication and sessions
- **Navigation**: Menu and navigation system

**Key Files:**
- `panel.go` - Main panel implementation
- `resource.go` - Resource interface and base implementation
- `contract.go` - Interfaces and contracts
- `auth_handler.go` - Authentication handlers
- `auth_middleware.go` - Auth middleware

### 2. Form Builder (`form/`)

Fluent API for building forms with validation.

**Features:**
- Type-safe field definitions
- Built-in validation rules
- Conditional fields
- File uploads
- Relationship fields

**Example:**
```go
form.New().SetSchema(
    form.Text("name").Required(),
    form.Email("email").Required(),
    form.Select("role").Options(roleOptions),
)
```

### 3. Table Builder (`table/`)

Interactive data tables with advanced features.

**Features:**
- Sorting and filtering
- Pagination
- Search
- Bulk actions
- Custom columns
- Export functionality

**Example:**
```go
table.New(getData).
    WithColumns(
        table.Text("name").Sortable(),
        table.Badge("status"),
    ).
    SetActions(actions.EditAction(), actions.DeleteAction())
```

### 4. Resource System (`engine/resource.go`)

Resources represent database entities with CRUD operations.

**Interface:**
```go
type Resource interface {
    GetMeta() ResourceMeta
    GetNavigation() ResourceNavigation
    GetViews() ResourceViews
    GetPermissions() ResourcePermissions
    GetCRUD() ResourceCRUD
}
```

### 5. Authentication (`auth/`)

Built-in authentication system with:
- Password hashing (bcrypt)
- Session management (SCS)
- User context
- Login/logout handlers

### 6. UI Components (`ui/`)

Reusable UI components built with Templ:
- **Atoms**: Buttons, badges, inputs, modals, toasts
- **Components**: Tables, forms, cards, navigation
- **Layouts**: Base layouts, dashboard layout
- **Icons**: Lucide icon system

## Data Flow

### Request Lifecycle

```
HTTP Request
    ↓
Middleware Stack
    ├─ Logger
    ├─ Recovery
    ├─ Rate Limiter
    └─ Authentication
    ↓
Router (Panel.Router())
    ↓
Resource Handler
    ├─ Index (List)
    ├─ Create (Form)
    ├─ Store (Save)
    ├─ Edit (Form)
    ├─ Update (Save)
    └─ Delete
    ↓
Database (Ent ORM)
    ↓
Response (Templ Template)
```

### Resource Registration

```
1. Define Ent Schema (internal/ent/schema/)
2. Generate Ent Code (go generate)
3. Create Resource (views/resources/)
4. Register Resource (internal/registry/resources.go)
5. Access via Panel
```

## Design Patterns

### 1. Builder Pattern
Used extensively in Form and Table builders for fluent API.

```go
form.New().
    SetSchema(...).
    SetValidation(...).
    Build()
```

### 2. Interface Segregation
Resources implement multiple small interfaces instead of one large interface.

### 3. Dependency Injection
Panel accepts dependencies (DB, Auth, Session) via setters.

```go
panel := engine.NewPanel("admin").
    SetDatabase(db).
    SetAuthManager(auth).
    SetSession(session)
```

### 4. Template Method
Base resource provides default implementations, resources override as needed.

## Configuration

Configuration is managed via `config.yaml` and loaded using Viper.

**Key Sections:**
- `server` - HTTP server settings
- `database` - Database connection
- `engine` - Panel configuration
- `auth` - Authentication settings
- `jobs` - Background job settings

See [PANEL_CONFIG.md](PANEL_CONFIG.md) for details.

## Database Layer

SublimeGo uses **Ent** as the ORM:

1. **Schema Definition** (`internal/ent/schema/`)
   - Define entities with fields, edges, indexes
   
2. **Code Generation**
   ```bash
   go generate ./internal/ent
   ```

3. **Migrations**
   - Automatic schema migration on startup
   - Manual migrations via `ent migrate`

4. **Querying**
   ```go
   users, err := client.User.Query().
       Where(user.EmailContains("@example.com")).
       All(ctx)
   ```

## View Layer

Views are built with **Templ** (type-safe Go templates):

1. **Define Template** (`.templ` file)
   ```templ
   package views
   
   templ UserList(users []User) {
       <div>
           for _, user := range users {
               <p>{ user.Name }</p>
           }
       </div>
   }
   ```

2. **Generate Go Code**
   ```bash
   templ generate
   ```

3. **Render in Handler**
   ```go
   templ.Handler(views.UserList(users)).ServeHTTP(w, r)
   ```

See [TEMPLATING.md](TEMPLATING.md) for details.

## Extension Points

### Custom Actions
```go
actions.NewAction("approve", "Approve").
    SetHandler(func(ctx context.Context, ids []int) error {
        // Custom logic
    })
```

### Custom Widgets
```go
widget.NewCustom("sales", func() templ.Component {
    return views.SalesWidget()
})
```

### Custom Middleware
```go
panel.Router().Use(myCustomMiddleware)
```

### Custom Fields
```go
form.NewField("custom", "CustomField").
    SetRender(func() templ.Component {
        return views.CustomField()
    })
```

## Performance Considerations

1. **Database Queries**
   - Use eager loading for relationships
   - Add indexes for frequently queried fields
   - Implement pagination for large datasets

2. **Caching**
   - Session data cached in memory
   - Static assets served with cache headers

3. **Asset Optimization**
   - Tailwind CSS purged in production
   - JavaScript minified
   - Icons loaded on-demand

## Security

1. **Authentication**
   - Bcrypt password hashing
   - Secure session cookies
   - CSRF protection (TODO)

2. **Authorization**
   - Resource-level permissions
   - Action-level permissions
   - Role-based access control (TODO)

3. **Input Validation**
   - Server-side validation
   - XSS prevention via Templ escaping
   - SQL injection prevention via Ent

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific package
go test ./form
```

See individual package `*_test.go` files for examples.

## Next Steps

- Read [RESOURCES_GUIDE.md](RESOURCES_GUIDE.md) to learn how to create resources
- Read [TEMPLATING.md](TEMPLATING.md) to learn about Templ templates
- Read [PANEL_CONFIG.md](PANEL_CONFIG.md) to configure your panel
