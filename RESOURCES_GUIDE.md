# Resources Guide

A resource represents a database entity (User, Product, Order) with full CRUD operations
and automatic admin panel integration.

---

## Quick Start

### 1. Define an Ent Schema

```bash
go run -mod=mod entgo.io/ent/cmd/ent new Product
```

Edit `internal/ent/schema/product.go`:

```go
package schema

import (
    "time"

    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
)

type Product struct{ ent.Schema }

func (Product) Fields() []ent.Field {
    return []ent.Field{
        field.String("name"),
        field.Text("description").Optional(),
        field.Float("price"),
        field.String("status").Default("draft"),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (Product) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("status"),
    }
}
```

### 2. Generate Ent Code

```bash
go generate ./internal/ent
```

### 3. Generate the Resource Scaffold

```bash
sublimego make:resource Product
```

This creates `views/resources/product_resource.go` with all required methods stubbed out.

### 4. Register the Resource

```go
panel.AddResources(NewProductResource(db))
```

---

## Resource Interface

A resource must implement the `engine.Resource` interface, which is composed of:

```go
type Resource interface {
    ResourceMeta        // Slug, Label, PluralLabel, Icon, Group, Sort
    ResourceViews       // Table(ctx), Form(ctx, item)
    ResourcePermissions // CanCreate, CanRead, CanUpdate, CanDelete
    ResourceCRUD        // List, Get, Create, Update, Delete, BulkDelete
    ResourceNavigation  // Badge(ctx), BadgeColor(ctx)
}
```

Use `engine.BaseResource` to get default no-op implementations, then override only what you need.

---

## Complete Resource Example

```go
package resources

import (
    "context"
    "fmt"
    "net/http"
    "strconv"

    "github.com/a-h/templ"
    "github.com/bozz33/sublimego/engine"
    "github.com/bozz33/sublimego/form"
    "github.com/bozz33/sublimego/table"
    "github.com/bozz33/sublimego/views"
    "github.com/bozz33/sublimego/internal/ent"
    entproduct "github.com/bozz33/sublimego/internal/ent/product"
)

type ProductResource struct {
    engine.BaseResource
    db *ent.Client
}

func NewProductResource(db *ent.Client) *ProductResource {
    return &ProductResource{db: db}
}

//  Meta 

func (r *ProductResource) Slug() string        { return "products" }
func (r *ProductResource) Label() string       { return "Product" }
func (r *ProductResource) PluralLabel() string { return "Products" }
func (r *ProductResource) Icon() string        { return "package" }
func (r *ProductResource) Group() string       { return "Catalog" }
func (r *ProductResource) Sort() int           { return 10 }

//  Navigation badge 

func (r *ProductResource) Badge(ctx context.Context) string {
    count, _ := r.db.Product.Query().
        Where(entproduct.Status("draft")).
        Count(ctx)
    if count == 0 {
        return ""
    }
    return strconv.Itoa(count)
}

func (r *ProductResource) BadgeColor(_ context.Context) string { return "warning" }

//  Permissions 

func (r *ProductResource) CanCreate(_ context.Context) bool { return true }
func (r *ProductResource) CanRead(_ context.Context) bool   { return true }
func (r *ProductResource) CanUpdate(_ context.Context) bool { return true }
func (r *ProductResource) CanDelete(_ context.Context) bool { return true }

//  Views 

func (r *ProductResource) Table(_ context.Context) templ.Component {
    t := table.New(nil).
        WithColumns(
            table.Text("name").WithLabel("Name").WithSortable(true).WithSearchable(true),
            table.Text("price").WithLabel("Price").WithSortable(true),
            table.Badge("status").WithLabel("Status"),
            table.Date("created_at").WithLabel("Created").WithSortable(true),
        ).
        WithFilters(
            table.SelectFilter("status", "Status").WithOptions(
                table.FilterOption{Value: "draft",     Label: "Draft"},
                table.FilterOption{Value: "published", Label: "Published"},
                table.FilterOption{Value: "archived",  Label: "Archived"},
            ),
        ).
        WithSummaries(
            table.NewSummary("price", table.SummarySum).WithLabel("Total").WithFormat("$%.2f"),
        ).
        WithGroups(
            table.GroupBy("status").WithLabel("By status").Collapsible(),
        )
    return views.GenericTable(t)
}

func (r *ProductResource) Form(_ context.Context, _ any) templ.Component {
    f := form.New().SetSchema(
        form.NewSection("General").SetSchema(
            form.NewText("name").Label("Name").Required(),
            form.NewRichEditor("description").Label("Description"),
            form.NewNumber("price").Label("Price").Required(),
        ),
        form.NewSection("Settings").SetSchema(
            form.NewSelect("status").Label("Status").WithOptions(
                form.Option{Value: "draft",     Label: "Draft"},
                form.Option{Value: "published", Label: "Published"},
                form.Option{Value: "archived",  Label: "Archived"},
            ),
            form.NewTagsInput("tags").Label("Tags"),
            form.NewColorPicker("color").Label("Color"),
        ),
    )
    return views.GenericForm(f)
}

//  CRUD 

func (r *ProductResource) List(ctx context.Context) ([]any, error) {
    products, err := r.db.Product.Query().
        Order(ent.Desc(entproduct.FieldCreatedAt)).
        All(ctx)
    if err != nil {
        return nil, fmt.Errorf("list products: %w", err)
    }
    out := make([]any, len(products))
    for i, p := range products {
        out[i] = p
    }
    return out, nil
}

func (r *ProductResource) Get(ctx context.Context, id string) (any, error) {
    n, err := strconv.Atoi(id)
    if err != nil {
        return nil, fmt.Errorf("invalid id: %w", err)
    }
    return r.db.Product.Get(ctx, n)
}

func (r *ProductResource) Create(ctx context.Context, req *http.Request) error {
    if err := req.ParseForm(); err != nil {
        return err
    }
    price, _ := strconv.ParseFloat(req.FormValue("price"), 64)
    _, err := r.db.Product.Create().
        SetName(req.FormValue("name")).
        SetDescription(req.FormValue("description")).
        SetPrice(price).
        SetStatus(req.FormValue("status")).
        Save(ctx)
    return err
}

func (r *ProductResource) Update(ctx context.Context, id string, req *http.Request) error {
    n, err := strconv.Atoi(id)
    if err != nil {
        return fmt.Errorf("invalid id: %w", err)
    }
    if err := req.ParseForm(); err != nil {
        return err
    }
    price, _ := strconv.ParseFloat(req.FormValue("price"), 64)
    return r.db.Product.UpdateOneID(n).
        SetName(req.FormValue("name")).
        SetDescription(req.FormValue("description")).
        SetPrice(price).
        SetStatus(req.FormValue("status")).
        Exec(ctx)
}

func (r *ProductResource) Delete(ctx context.Context, id string) error {
    n, err := strconv.Atoi(id)
    if err != nil {
        return fmt.Errorf("invalid id: %w", err)
    }
    return r.db.Product.DeleteOneID(n).Exec(ctx)
}

func (r *ProductResource) BulkDelete(ctx context.Context, ids []string) error {
    ns := make([]int, 0, len(ids))
    for _, id := range ids {
        n, err := strconv.Atoi(id)
        if err != nil {
            continue
        }
        ns = append(ns, n)
    }
    _, err := r.db.Product.Delete().
        Where(entproduct.IDIn(ns...)).
        Exec(ctx)
    return err
}
```

---

## Optional Interfaces

Implement these to unlock additional features:

### `ResourceViewable`  Detail view (Infolist)

```go
func (r *ProductResource) View(ctx context.Context, item any) templ.Component {
    p := item.(*ent.Product)
    il := infolist.New(p).SetSchema(
        infolist.Text("name").Label("Name"),
        infolist.Text("price").Label("Price"),
        infolist.Badge("status").Label("Status"),
    )
    return views.GenericInfolist(il)
}
```

### `ResourceQueryable`  Full query control (search + sort + filter + pagination)

```go
func (r *ProductResource) ListQuery(ctx context.Context, q engine.ListQuery) ([]any, int, error) {
    query := r.db.Product.Query()

    if q.Search != "" {
        query = query.Where(entproduct.NameContainsFold(q.Search))
    }
    if status := q.Filters["status"]; status != "" {
        query = query.Where(entproduct.Status(status))
    }

    total, _ := query.Count(ctx)

    if q.SortKey != "" {
        // apply sort...
    }

    products, err := query.
        Offset((q.Page - 1) * q.PerPage).
        Limit(q.PerPage).
        All(ctx)

    out := make([]any, len(products))
    for i, p := range products {
        out[i] = p
    }
    return out, total, err
}
```

### `TenantAware`  Multi-tenancy

```go
func (r *ProductResource) SetTenant(t *engine.Tenant) {
    r.tenantID = t.ID
}
```

---

## Nested Resources (Relation Manager)

```go
// In your resource, define a relation
func (r *OrderResource) Relations() []engine.RelationManager {
    return []engine.RelationManager{
        engine.HasMany("items", "Order Items", NewOrderItemResource(r.db)),
        engine.BelongsTo("customer", "Customer", NewCustomerResource(r.db)),
    }
}
```

---

## Custom Pages

```go
sublimego make:page Dashboard
```

```go
type DashboardPage struct{}

func (p *DashboardPage) Slug() string  { return "dashboard" }
func (p *DashboardPage) Label() string { return "Dashboard" }
func (p *DashboardPage) Icon() string  { return "home" }

func (p *DashboardPage) Render(ctx context.Context) templ.Component {
    return views.DashboardView()
}
```

---

## Export and Import

```go
// Export  add to your resource
func (r *ProductResource) ExportURL() string { return "/admin/products/export" }

// Import  add to your resource
func (r *ProductResource) ImportURL() string { return "/admin/products/import" }
```

The framework handles the HTTP endpoints automatically.
