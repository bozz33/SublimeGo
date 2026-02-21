# Multi-Tenant Guide

> Complete guide for building multi-tenant applications with SublimeGo — automatic database provisioning, subdomain/domain routing, and tenant isolation.

---

## Overview

SublimeGo provides built-in multi-tenancy support with:

- **Automatic database creation** for each tenant
- **Subdomain routing** (`{tenant}.example.com`)
- **Custom domain routing** (`custom-domain.com`)
- **Path-based routing** (`/tenant/admin/...`)
- **Tenant isolation** for resources
- **Multiple panels** per tenant

---

## Quick Start

### 1. Create a Tenant Provisioner

```go
import "github.com/bozz33/sublimego/engine"

provisioner := engine.NewTenantProvisioner("./tenants").
    WithMigrations(func(ctx context.Context, client *ent.Client) error {
        // Run schema migrations
        return client.Schema.Create(ctx)
    })
```

### 2. Register Tenants

```go
resolver := engine.NewSubdomainResolver("example.com")

acme := &engine.Tenant{
    ID:        "acme",
    Name:      "Acme Corp",
    Subdomain: "acme",
}

globex := &engine.Tenant{
    ID:        "globex",
    Name:      "Globex Inc",
    Subdomain: "globex",
}

// Provision databases automatically
provisioner.ProvisionTenant(ctx, acme)
provisioner.ProvisionTenant(ctx, globex)

// Register for routing
resolver.Register(acme).Register(globex)
```

### 3. Add Tenant Middleware

```go
panel := engine.NewPanel("admin").
    SetPath("/admin").
    WithMiddleware(engine.TenantResourceMiddleware(panel))

// Inject tenant resolver middleware
mux.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tenant, ok := resolver.Resolve(r)
        if ok {
            r = r.WithContext(engine.WithTenant(r.Context(), tenant))
        }
        next.ServeHTTP(w, r)
    })
})
```

---

## Automatic Database Provisioning

### TenantProvisioner

The `TenantProvisioner` handles automatic database creation and migration:

```go
provisioner := engine.NewTenantProvisioner("./tenants")

// Provision a single tenant
dsn, err := provisioner.ProvisionTenant(ctx, tenant)
if err != nil {
    log.Fatal(err)
}

// DSN is automatically stored in tenant.Meta["db_dsn"]
fmt.Println(dsn) // file:./tenants/acme.db?cache=shared&_fk=1
```

### Running Migrations

```go
provisioner := engine.NewTenantProvisioner("./tenants").
    WithMigrations(func(ctx context.Context, client *ent.Client) error {
        // Create schema
        if err := client.Schema.Create(ctx); err != nil {
            return err
        }
        
        // Seed initial data
        _, err := client.User.Create().
            SetEmail("admin@tenant.com").
            SetPassword("password").
            Save(ctx)
        
        return err
    })
```

### Provisioning Multiple Tenants

```go
tenants := []*engine.Tenant{acme, globex, initech}

if err := provisioner.ProvisionAll(ctx, tenants); err != nil {
    log.Fatal(err)
}
```

### Getting a Tenant's Database Client

```go
client, err := provisioner.GetTenantClient(tenant)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Use the client for tenant-specific queries
users, err := client.User.Query().All(ctx)
```

### Deleting a Tenant

```go
// WARNING: This permanently deletes the tenant's database
if err := provisioner.DeleteTenant(tenant); err != nil {
    log.Fatal(err)
}
```

---

## Tenant Routing

### Subdomain-Based Routing

```go
resolver := engine.NewSubdomainResolver("example.com")

resolver.Register(&engine.Tenant{
    ID:        "acme",
    Subdomain: "acme",
    Name:      "Acme Corp",
})

// Resolves:
// acme.example.com → acme tenant
// acme.example.com:8080 → acme tenant (port stripped)
```

### Custom Domain Routing

```go
resolver.Register(&engine.Tenant{
    ID:     "acme",
    Domain: "acmecorp.com",
    Name:   "Acme Corp",
})

// Resolves:
// acmecorp.com → acme tenant
// www.acmecorp.com → not resolved (exact match only)
```

### Path-Based Routing

```go
resolver := engine.NewPathResolver()

resolver.Register(&engine.Tenant{
    ID:   "acme",
    Name: "Acme Corp",
})

// Resolves:
// /acme/admin/users → acme tenant
// /acme/dashboard → acme tenant
```

---

## Tenant-Aware Resources

Implement `TenantAware` to automatically scope resource queries by tenant:

```go
type UserResource struct {
    engine.BaseResource
    db     *ent.Client
    tenant *engine.Tenant
}

// Implement TenantAware interface
func (r *UserResource) SetTenant(tenant *engine.Tenant) {
    r.tenant = tenant
}

// Use tenant in queries
func (r *UserResource) List(ctx context.Context) ([]any, error) {
    client, err := provisioner.GetTenantClient(r.tenant)
    if err != nil {
        return nil, err
    }
    defer client.Close()
    
    users, err := client.User.Query().All(ctx)
    // ... convert to []any
    return result, nil
}
```

The `TenantResourceMiddleware` automatically calls `SetTenant()` on all resources before each request.

---

## Accessing the Current Tenant

### In Middleware

```go
func MyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tenant := engine.TenantFromContext(r.Context())
        if tenant != nil {
            log.Printf("Request from tenant: %s", tenant.Name)
        }
        next.ServeHTTP(w, r)
    })
}
```

### In Handlers

```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
    tenant := engine.TenantFromContext(r.Context())
    if tenant == nil {
        http.Error(w, "No tenant", http.StatusBadRequest)
        return
    }
    
    fmt.Fprintf(w, "Hello from %s!", tenant.Name)
}
```

---

## Tenant Metadata

Store arbitrary tenant-specific data in `tenant.Meta`:

```go
tenant := &engine.Tenant{
    ID:   "acme",
    Name: "Acme Corp",
    Meta: map[string]any{
        "plan":       "enterprise",
        "max_users":  100,
        "features":   []string{"sso", "api", "webhooks"},
        "db_dsn":     "...", // automatically set by provisioner
    },
}

// Access metadata
plan := tenant.Meta["plan"].(string)
maxUsers := tenant.Meta["max_users"].(int)
```

---

## Complete Example

```go
package main

import (
    "context"
    "log"
    "net/http"
    
    "github.com/bozz33/sublimego/engine"
)

func main() {
    ctx := context.Background()
    
    // 1. Create provisioner
    provisioner := engine.NewTenantProvisioner("./tenants").
        WithMigrations(func(ctx context.Context, client *ent.Client) error {
            return client.Schema.Create(ctx)
        })
    
    // 2. Define tenants
    tenants := []*engine.Tenant{
        {ID: "acme", Name: "Acme Corp", Subdomain: "acme"},
        {ID: "globex", Name: "Globex Inc", Subdomain: "globex"},
    }
    
    // 3. Provision databases
    if err := provisioner.ProvisionAll(ctx, tenants); err != nil {
        log.Fatal(err)
    }
    
    // 4. Setup resolver
    resolver := engine.NewSubdomainResolver("example.com")
    for _, t := range tenants {
        resolver.Register(t)
    }
    
    // 5. Create panel
    panel := engine.NewPanel("admin").
        SetPath("/admin").
        WithMiddleware(engine.TenantResourceMiddleware(panel))
    
    // 6. Setup routing
    mux := http.NewServeMux()
    
    // Tenant resolver middleware
    mux.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            tenant, ok := resolver.Resolve(r)
            if !ok {
                http.Error(w, "Tenant not found", http.StatusNotFound)
                return
            }
            r = r.WithContext(engine.WithTenant(r.Context(), tenant))
            next.ServeHTTP(w, r)
        })
    })
    
    mux.Handle("/admin/", panel.Router())
    
    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

---

## Best Practices

1. **Always use TenantProvisioner** — Don't create databases manually
2. **Store tenant config in Meta** — Plan, features, limits, etc.
3. **Implement TenantAware** — For automatic query scoping
4. **Use middleware** — Don't resolve tenants in every handler
5. **Validate tenant access** — Check permissions before operations
6. **Backup tenant databases** — Each tenant has its own DB file
7. **Monitor tenant usage** — Track DB size, request count, etc.
