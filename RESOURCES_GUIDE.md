# Resources Guide

This guide explains how to create and manage resources in SublimeGo.

## What is a Resource?

A resource represents a database entity (like User, Product, Post) with full CRUD operations and admin panel integration.

## Quick Start

### 1. Create Ent Schema

```bash
# Generate a new Ent schema
go run -mod=mod entgo.io/ent/cmd/ent new Product
```

Edit `internal/ent/schema/product.go`:

```go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
)

type Product struct {
    ent.Schema
}

func (Product) Fields() []ent.Field {
    return []ent.Field{
        field.String("name"),
        field.Text("description").Optional(),
        field.Float("price"),
        field.String("status").Default("draft"),
        field.Time("created_at").Default(time.Now),
    }
}
```

Generate Ent code:

```bash
go generate ./internal/ent
```

### 2. Create Resource File

Create `views/resources/product.go`:

```go
package resources

import (
    "context"
    "fmt"
    
    "github.com/bozz33/SublimeGo/internal/ent"
    "github.com/bozz33/SublimeGo/internal/ent/product"
    "github.com/bozz33/SublimeGo/pkg/actions"
    "github.com/bozz33/SublimeGo/pkg/engine"
    "github.com/bozz33/SublimeGo/pkg/form"
    "github.com/bozz33/SublimeGo/pkg/table"
)

type ProductResource struct {
    engine.BaseResource
    client *ent.Client
}

func NewProductResource(client *ent.Client) *ProductResource {
    return &ProductResource{client: client}
}

// Meta information
func (r *ProductResource) GetMeta() engine.ResourceMeta {
    return engine.ResourceMeta{
        Name:         "product",
        Label:        "Product",
        PluralLabel:  "Products",
        Icon:         "package",
        Description:  "Manage your products",
    }
}

// Navigation
func (r *ProductResource) GetNavigation() engine.ResourceNavigation {
    return engine.ResourceNavigation{
        Group:    "Catalog",
        Position: 1,
        Visible:  true,
    }
}

// Form builder
func (r *ProductResource) GetForm() *form.Form {
    return form.New().SetSchema(
        form.Text("name").
            Label("Product Name").
            Required().
            Placeholder("Enter product name"),
            
        form.Textarea("description").
            Label("Description").
            Rows(5),
            
        form.Number("price").
            Label("Price").
            Required().
            Min(0).
            Step(0.01),
            
        form.Select("status").
            Label("Status").
            Options([]form.Option{
                {Value: "draft", Label: "Draft"},
                {Value: "published", Label: "Published"},
                {Value: "archived", Label: "Archived"},
            }).
            Default("draft"),
    )
}

// Table builder
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
        ).
        SetBulkActions(
            actions.BulkDeleteAction(),
        )
}

// CRUD Operations
func (r *ProductResource) GetData(ctx context.Context, state table.TableState) ([]table.Row, int, error) {
    query := r.client.Product.Query()
    
    // Search
    if state.Search != "" {
        query = query.Where(product.NameContains(state.Search))
    }
    
    // Filters
    for _, filter := range state.Filters {
        if filter.Field == "status" {
            query = query.Where(product.StatusEQ(filter.Value))
        }
    }
    
    // Count total
    total, err := query.Count(ctx)
    if err != nil {
        return nil, 0, err
    }
    
    // Sort
    if state.SortField != "" {
        if state.SortDirection == "asc" {
            query = query.Order(ent.Asc(state.SortField))
        } else {
            query = query.Order(ent.Desc(state.SortField))
        }
    }
    
    // Pagination
    query = query.
        Limit(state.PerPage).
        Offset((state.Page - 1) * state.PerPage)
    
    products, err := query.All(ctx)
    if err != nil {
        return nil, 0, err
    }
    
    // Convert to rows
    rows := make([]table.Row, len(products))
    for i, p := range products {
        rows[i] = table.Row{
            "id":          p.ID,
            "name":        p.Name,
            "description": p.Description,
            "price":       p.Price,
            "status":      p.Status,
            "created_at":  p.CreatedAt,
        }
    }
    
    return rows, total, nil
}

func (r *ProductResource) Create(ctx context.Context, data map[string]interface{}) error {
    _, err := r.client.Product.Create().
        SetName(data["name"].(string)).
        SetNillableDescription(toStringPtr(data["description"])).
        SetPrice(data["price"].(float64)).
        SetStatus(data["status"].(string)).
        Save(ctx)
    return err
}

func (r *ProductResource) Update(ctx context.Context, id int, data map[string]interface{}) error {
    return r.client.Product.UpdateOneID(id).
        SetName(data["name"].(string)).
        SetNillableDescription(toStringPtr(data["description"])).
        SetPrice(data["price"].(float64)).
        SetStatus(data["status"].(string)).
        Exec(ctx)
}

func (r *ProductResource) Delete(ctx context.Context, id int) error {
    return r.client.Product.DeleteOneID(id).Exec(ctx)
}

func (r *ProductResource) Find(ctx context.Context, id int) (map[string]interface{}, error) {
    p, err := r.client.Product.Get(ctx, id)
    if err != nil {
        return nil, err
    }
    
    return map[string]interface{}{
        "id":          p.ID,
        "name":        p.Name,
        "description": p.Description,
        "price":       p.Price,
        "status":      p.Status,
        "created_at":  p.CreatedAt,
    }, nil
}

// Helper
func toStringPtr(v interface{}) *string {
    if v == nil {
        return nil
    }
    s := v.(string)
    return &s
}
```

### 3. Register Resource

Edit `internal/registry/resources.go`:

```go
func GetResources(client *ent.Client) []engine.Resource {
    return []engine.Resource{
        resources.NewProductResource(client),
        // ... other resources
    }
}
```

### 4. Run Migrations

```bash
go run cmd/sublimego/main.go serve
```

The schema will be created automatically.

## Advanced Features

### Custom Actions

Add custom actions to your resource:

```go
func (r *ProductResource) GetActions() []actions.Action {
    return []actions.Action{
        actions.NewAction("publish", "Publish").
            SetIcon("check").
            SetColor("green").
            SetHandler(func(ctx context.Context, ids []int) error {
                return r.client.Product.Update().
                    Where(product.IDIn(ids...)).
                    SetStatus("published").
                    Exec(ctx)
            }),
            
        actions.NewAction("archive", "Archive").
            SetIcon("archive").
            SetConfirmation("Are you sure you want to archive these products?").
            SetHandler(func(ctx context.Context, ids []int) error {
                return r.client.Product.Update().
                    Where(product.IDIn(ids...)).
                    SetStatus("archived").
                    Exec(ctx)
            }),
    }
}
```

### Filters

Add filters to your table:

```go
func (r *ProductResource) GetFilters() []table.Filter {
    return []table.Filter{
        table.SelectFilter("status", "Status").
            Options([]table.FilterOption{
                {Value: "draft", Label: "Draft"},
                {Value: "published", Label: "Published"},
                {Value: "archived", Label: "Archived"},
            }),
            
        table.DateRangeFilter("created_at", "Created Date"),
        
        table.NumberRangeFilter("price", "Price Range"),
    }
}
```

### Relationships

Handle relationships in your resource:

```go
// In Ent schema
func (Product) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("category", Category.Type).
            Ref("products").
            Unique(),
    }
}

// In form
form.Select("category_id").
    Label("Category").
    Options(r.getCategoryOptions()).
    Required()

// In table
table.Text("category").
    Value(func(row table.Row) string {
        if cat := row["category"]; cat != nil {
            return cat.(*ent.Category).Name
        }
        return "-"
    })

// In GetData - eager load
products, err := query.WithCategory().All(ctx)
```

### Validation

Add custom validation:

```go
func (r *ProductResource) Validate(ctx context.Context, data map[string]interface{}) error {
    if price, ok := data["price"].(float64); ok {
        if price < 0 {
            return errors.New("price must be positive")
        }
    }
    
    // Check unique name
    name := data["name"].(string)
    exists, err := r.client.Product.Query().
        Where(product.NameEQ(name)).
        Exist(ctx)
    if err != nil {
        return err
    }
    if exists {
        return errors.New("product name already exists")
    }
    
    return nil
}
```

### Permissions

Control access to resources:

```go
func (r *ProductResource) GetPermissions() engine.ResourcePermissions {
    return engine.ResourcePermissions{
        View:   func(ctx context.Context) bool { return true },
        Create: func(ctx context.Context) bool { 
            user := auth.UserFromContext(ctx)
            return user != nil && user.Role == "admin"
        },
        Edit:   func(ctx context.Context) bool { 
            user := auth.UserFromContext(ctx)
            return user != nil && user.Role == "admin"
        },
        Delete: func(ctx context.Context) bool { 
            user := auth.UserFromContext(ctx)
            return user != nil && user.Role == "admin"
        },
    }
}
```

### Export

Enable data export:

```go
func (r *ProductResource) GetExportConfig() *export.Config {
    return export.NewConfig().
        SetFormats([]string{"csv", "excel", "pdf"}).
        SetColumns([]export.Column{
            {Field: "id", Label: "ID"},
            {Field: "name", Label: "Name"},
            {Field: "price", Label: "Price"},
            {Field: "status", Label: "Status"},
        })
}
```

## Field Types

### Form Fields

```go
// Text input
form.Text("name").Label("Name").Required()

// Email
form.Email("email").Required()

// Password
form.Password("password").MinLength(8)

// Number
form.Number("quantity").Min(0).Max(100)

// Textarea
form.Textarea("description").Rows(5)

// Select
form.Select("category").Options(options)

// Checkbox
form.Checkbox("active").Default(true)

// Date
form.Date("birth_date")

// DateTime
form.DateTime("published_at")

// File upload
form.File("image").Accept("image/*")

// Rich text editor
form.RichText("content")

// JSON editor
form.JSON("metadata")
```

### Table Columns

```go
// ID column
table.ID("id")

// Text
table.Text("name").Sortable().Searchable()

// Number
table.Number("quantity").Sortable()

// Currency
table.Currency("price").Sortable()

// Badge
table.Badge("status").Colors(colorMap)

// Boolean
table.Boolean("active")

// Date
table.Date("created_at").Sortable()

// DateTime
table.DateTime("updated_at").Sortable()

// Image
table.Image("avatar")

// Custom
table.Custom("actions", func(row table.Row) templ.Component {
    return views.CustomCell(row)
})
```

## Best Practices

1. **Keep resources focused** - One resource per entity
2. **Use eager loading** - Load relationships when needed
3. **Add indexes** - For sortable and searchable fields
4. **Validate input** - Both client and server side
5. **Handle errors** - Return meaningful error messages
6. **Use transactions** - For complex operations
7. **Cache when possible** - For dropdown options, etc.
8. **Test your resources** - Write unit tests for CRUD operations

## CLI Commands

```bash
# Generate resource scaffold
go run cmd/sublimego/main.go make:resource Product

# List all resources
go run cmd/sublimego/main.go resource list

# Generate Ent code
go generate ./internal/ent

# Generate Templ code
templ generate
```

## Troubleshooting

### Resource not showing in menu
- Check `GetNavigation().Visible` is true
- Verify resource is registered in `registry/resources.go`
- Check permissions allow viewing

### Form not saving
- Verify field names match database columns
- Check validation rules
- Look for errors in server logs

### Table not loading data
- Check `GetData` returns correct format
- Verify database query is correct
- Check for errors in browser console

## Examples

See `views/resources/` for complete examples of:
- Simple CRUD resource
- Resource with relationships
- Resource with custom actions
- Resource with file uploads
- Resource with complex validation
