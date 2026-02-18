package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSubdomainResolver(t *testing.T) {
	resolver := NewSubdomainResolver("example.com")
	resolver.Register(&Tenant{ID: "acme", Subdomain: "acme", Name: "Acme Corp"})
	resolver.Register(&Tenant{ID: "globex", Subdomain: "globex", Name: "Globex"})

	tests := []struct {
		host      string
		wantID    string
		wantFound bool
	}{
		{"acme.example.com", "acme", true},
		{"globex.example.com", "globex", true},
		{"unknown.example.com", "", false},
		{"example.com", "", false},
		{"acme.example.com:8080", "acme", true}, // with port
	}

	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Host = tt.host
		tenant, ok := resolver.Resolve(r)
		if ok != tt.wantFound {
			t.Errorf("host=%q: found=%v, want %v", tt.host, ok, tt.wantFound)
			continue
		}
		if ok && tenant.ID != tt.wantID {
			t.Errorf("host=%q: tenantID=%q, want %q", tt.host, tenant.ID, tt.wantID)
		}
	}
}

func TestSubdomainResolverExactDomain(t *testing.T) {
	resolver := NewSubdomainResolver("example.com")
	resolver.Register(&Tenant{ID: "custom", Domain: "custom-domain.io", Subdomain: "custom", Name: "Custom"})

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Host = "custom-domain.io"
	tenant, ok := resolver.Resolve(r)
	if !ok {
		t.Fatal("expected tenant to be found by exact domain")
	}
	if tenant.ID != "custom" {
		t.Errorf("got tenantID=%q, want %q", tenant.ID, "custom")
	}
}

func TestPathResolver(t *testing.T) {
	resolver := NewPathResolver()
	resolver.Register(&Tenant{ID: "acme", Name: "Acme"})
	resolver.Register(&Tenant{ID: "globex", Name: "Globex"})

	tests := []struct {
		path      string
		wantID    string
		wantFound bool
	}{
		{"/acme/admin/users", "acme", true},
		{"/globex/dashboard", "globex", true},
		{"/unknown/page", "", false},
		{"/", "", false},
	}

	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodGet, tt.path, nil)
		tenant, ok := resolver.Resolve(r)
		if ok != tt.wantFound {
			t.Errorf("path=%q: found=%v, want %v", tt.path, ok, tt.wantFound)
			continue
		}
		if ok && tenant.ID != tt.wantID {
			t.Errorf("path=%q: tenantID=%q, want %q", tt.path, tenant.ID, tt.wantID)
		}
	}
}

func TestTenantContext(t *testing.T) {
	tenant := &Tenant{ID: "acme", Name: "Acme Corp"}
	ctx := WithTenant(context.Background(), tenant)

	got := TenantFromContext(ctx)
	if got == nil {
		t.Fatal("expected tenant in context, got nil")
	}
	if got.ID != "acme" {
		t.Errorf("got ID=%q, want %q", got.ID, "acme")
	}
}

func TestTenantFromContextNil(t *testing.T) {
	got := TenantFromContext(context.Background())
	if got != nil {
		t.Errorf("expected nil tenant from empty context, got %+v", got)
	}
}

func TestTenantMiddlewareInjectsContext(t *testing.T) {
	resolver := NewSubdomainResolver("example.com")
	resolver.Register(&Tenant{ID: "acme", Subdomain: "acme", Name: "Acme"})

	var capturedTenant *Tenant
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTenant = TenantFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := TenantMiddleware(resolver, false)(inner)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Host = "acme.example.com"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if capturedTenant == nil || capturedTenant.ID != "acme" {
		t.Errorf("expected tenant acme in context, got %+v", capturedTenant)
	}
}

func TestTenantMiddlewareRequireTenant404(t *testing.T) {
	resolver := NewSubdomainResolver("example.com")

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := TenantMiddleware(resolver, true)(inner)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Host = "unknown.example.com"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestTenantMiddlewareOptionalPassthrough(t *testing.T) {
	resolver := NewSubdomainResolver("example.com")

	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := TenantMiddleware(resolver, false)(inner)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Host = "unknown.example.com"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if !called {
		t.Error("expected inner handler to be called when requireTenant=false")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
