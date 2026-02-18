package engine

import (
	"context"
	"net/http"
	"strings"
	"sync"
)

// ---------------------------------------------------------------------------
// Tenant — core type
// ---------------------------------------------------------------------------

// Tenant represents an isolated tenant in a multi-tenant setup.
type Tenant struct {
	// ID is the unique identifier (e.g. "acme", "globex").
	ID string
	// Domain is the full hostname that maps to this tenant (e.g. "acme.example.com").
	// Leave empty to use subdomain-based resolution only.
	Domain string
	// Subdomain is the subdomain prefix (e.g. "acme" for "acme.example.com").
	Subdomain string
	// Name is the display name of the tenant.
	Name string
	// Meta holds arbitrary tenant-specific key/value data (DB DSN, plan, etc.).
	Meta map[string]any
}

// ---------------------------------------------------------------------------
// TenantResolver — extracts tenant from request
// ---------------------------------------------------------------------------

// TenantResolver resolves the current tenant from an HTTP request.
type TenantResolver interface {
	Resolve(r *http.Request) (*Tenant, bool)
}

// SubdomainResolver resolves tenants by subdomain (e.g. {tenant}.example.com).
type SubdomainResolver struct {
	mu      sync.RWMutex
	tenants map[string]*Tenant // subdomain -> Tenant
	base    string             // base domain, e.g. "example.com"
}

// NewSubdomainResolver creates a resolver that maps subdomains to tenants.
func NewSubdomainResolver(baseDomain string) *SubdomainResolver {
	return &SubdomainResolver{
		tenants: make(map[string]*Tenant),
		base:    baseDomain,
	}
}

// Register adds a tenant to the resolver.
func (r *SubdomainResolver) Register(t *Tenant) *SubdomainResolver {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants[t.Subdomain] = t
	if t.Domain != "" {
		r.tenants[t.Domain] = t
	}
	return r
}

// Resolve extracts the subdomain from the host and looks up the tenant.
func (r *SubdomainResolver) Resolve(req *http.Request) (*Tenant, bool) {
	host := req.Host
	// Strip port
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try exact domain match first
	if t, ok := r.tenants[host]; ok {
		return t, true
	}

	// Try subdomain extraction: "acme.example.com" -> "acme"
	if strings.HasSuffix(host, "."+r.base) {
		sub := strings.TrimSuffix(host, "."+r.base)
		if t, ok := r.tenants[sub]; ok {
			return t, true
		}
	}

	return nil, false
}

// PathResolver resolves tenants from the URL path prefix (e.g. /acme/admin/...).
type PathResolver struct {
	mu      sync.RWMutex
	tenants map[string]*Tenant // path segment -> Tenant
}

// NewPathResolver creates a resolver that maps URL path prefixes to tenants.
func NewPathResolver() *PathResolver {
	return &PathResolver{tenants: make(map[string]*Tenant)}
}

// Register adds a tenant.
func (r *PathResolver) Register(t *Tenant) *PathResolver {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants[t.ID] = t
	return r
}

// Resolve extracts the first path segment and looks up the tenant.
func (r *PathResolver) Resolve(req *http.Request) (*Tenant, bool) {
	path := strings.TrimPrefix(req.URL.Path, "/")
	seg := strings.SplitN(path, "/", 2)[0]

	r.mu.RLock()
	defer r.mu.RUnlock()

	if t, ok := r.tenants[seg]; ok {
		return t, true
	}
	return nil, false
}

// ---------------------------------------------------------------------------
// Context helpers
// ---------------------------------------------------------------------------

const contextKeyTenant contextKey = "tenant"

// WithTenant injects a tenant into the context.
func WithTenant(ctx context.Context, t *Tenant) context.Context {
	return context.WithValue(ctx, contextKeyTenant, t)
}

// TenantFromContext retrieves the current tenant from context.
// Returns nil if no tenant is set (single-tenant mode).
func TenantFromContext(ctx context.Context) *Tenant {
	if t, ok := ctx.Value(contextKeyTenant).(*Tenant); ok {
		return t
	}
	return nil
}

// ---------------------------------------------------------------------------
// TenantMiddleware — injects tenant into every request context
// ---------------------------------------------------------------------------

// TenantMiddleware resolves the tenant for each request and injects it into ctx.
// If the tenant cannot be resolved and requireTenant is true, it returns 404.
func TenantMiddleware(resolver TenantResolver, requireTenant bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := resolver.Resolve(r)
			if !ok {
				if requireTenant {
					http.NotFound(w, r)
					return
				}
				next.ServeHTTP(w, r)
				return
			}
			ctx := WithTenant(r.Context(), tenant)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ---------------------------------------------------------------------------
// MultiPanelRouter — routes requests to the correct Panel per tenant
// ---------------------------------------------------------------------------

// MultiPanelRouter dispatches HTTP requests to the correct Panel
// based on the resolved tenant. Each tenant gets its own Panel instance
// (with its own resources, auth, DB, etc.).
type MultiPanelRouter struct {
	resolver TenantResolver
	panels   map[string]*Panel // tenantID -> Panel
	mu       sync.RWMutex
	handlers map[string]http.Handler // tenantID -> compiled http.Handler
	fallback http.Handler            // shown when no tenant matches
}

// NewMultiPanelRouter creates a router that dispatches to per-tenant panels.
func NewMultiPanelRouter(resolver TenantResolver) *MultiPanelRouter {
	return &MultiPanelRouter{
		resolver: resolver,
		panels:   make(map[string]*Panel),
		handlers: make(map[string]http.Handler),
	}
}

// RegisterPanel associates a Panel with a tenant ID.
// Call Build() on the panel before registering, or let ServeHTTP build lazily.
func (m *MultiPanelRouter) RegisterPanel(tenantID string, p *Panel) *MultiPanelRouter {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.panels[tenantID] = p
	return m
}

// SetFallback sets the handler used when no tenant matches.
func (m *MultiPanelRouter) SetFallback(h http.Handler) *MultiPanelRouter {
	m.fallback = h
	return m
}

// ServeHTTP resolves the tenant and dispatches to the correct panel handler.
func (m *MultiPanelRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tenant, ok := m.resolver.Resolve(r)
	if !ok {
		if m.fallback != nil {
			m.fallback.ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
		return
	}

	// Inject tenant into context
	ctx := WithTenant(r.Context(), tenant)
	r = r.WithContext(ctx)

	// Get or build the handler for this tenant
	h := m.getHandler(tenant.ID)
	if h == nil {
		http.Error(w, "Panel not configured for tenant: "+tenant.ID, http.StatusServiceUnavailable)
		return
	}

	h.ServeHTTP(w, r)
}

// getHandler returns the compiled http.Handler for a tenant, building it lazily.
func (m *MultiPanelRouter) getHandler(tenantID string) http.Handler {
	m.mu.RLock()
	h, ok := m.handlers[tenantID]
	m.mu.RUnlock()
	if ok {
		return h
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if h, ok = m.handlers[tenantID]; ok {
		return h
	}

	p, ok := m.panels[tenantID]
	if !ok {
		return nil
	}

	h = p.Router()
	m.handlers[tenantID] = h
	return h
}

// ---------------------------------------------------------------------------
// TenantAware — optional interface for resources that need tenant isolation
// ---------------------------------------------------------------------------

// TenantAware is an optional interface for resources that scope their data
// to the current tenant. Implement it to filter DB queries by tenant ID.
type TenantAware interface {
	// SetTenant configures the resource for the given tenant.
	// Called automatically by TenantResourceMiddleware before each request.
	SetTenant(tenant *Tenant)
}

// TenantResourceMiddleware wraps a Panel's handler to call SetTenant on all
// TenantAware resources before each request.
func TenantResourceMiddleware(p *Panel) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenant := TenantFromContext(r.Context())
			if tenant != nil {
				for _, res := range p.Resources {
					if ta, ok := res.(TenantAware); ok {
						ta.SetTenant(tenant)
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
