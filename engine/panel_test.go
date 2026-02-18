package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bozz33/sublimego/ui/layouts"
)

func TestNewPanel_Defaults(t *testing.T) {
	p := NewPanel("admin")

	if p.ID != "admin" {
		t.Errorf("expected ID=admin, got %s", p.ID)
	}
	if p.BrandName != "SublimeGo" {
		t.Errorf("expected BrandName=SublimeGo, got %s", p.BrandName)
	}
	if p.PrimaryColor != "green" {
		t.Errorf("expected PrimaryColor=green, got %s", p.PrimaryColor)
	}
	if !p.Registration {
		t.Error("expected Registration=true by default")
	}
	if !p.Profile {
		t.Error("expected Profile=true by default")
	}
	if !p.Notifications {
		t.Error("expected Notifications=true by default")
	}
	if !p.PasswordReset {
		t.Error("expected PasswordReset=true by default")
	}
}

func TestPanel_FluentAPI(t *testing.T) {
	p := NewPanel("test").
		WithBrandName("MyApp").
		WithLogo("/logo.svg").
		WithFavicon("/favicon.ico").
		WithPrimaryColor("blue").
		WithDarkMode(true).
		EnableRegistration(false).
		EnableNotifications(false).
		EnableProfile(false).
		EnablePasswordReset(false).
		WithPath("/admin")

	if p.BrandName != "MyApp" {
		t.Errorf("expected MyApp, got %s", p.BrandName)
	}
	if p.Logo != "/logo.svg" {
		t.Errorf("expected /logo.svg, got %s", p.Logo)
	}
	if p.PrimaryColor != "blue" {
		t.Errorf("expected blue, got %s", p.PrimaryColor)
	}
	if !p.DarkMode {
		t.Error("expected DarkMode=true")
	}
	if p.Registration {
		t.Error("expected Registration=false")
	}
	if p.Notifications {
		t.Error("expected Notifications=false")
	}
	if p.Profile {
		t.Error("expected Profile=false")
	}
	if p.Path != "/admin" {
		t.Errorf("expected /admin, got %s", p.Path)
	}
}

func TestPanel_SyncConfig(t *testing.T) {
	p := NewPanel("sync-test").
		WithBrandName("SyncApp").
		WithPrimaryColor("purple").
		WithPath("/sync")

	p.syncConfig()

	cfg := layouts.GetPanelConfig()
	if cfg.Name != "SyncApp" {
		t.Errorf("expected SyncApp, got %s", cfg.Name)
	}
	if cfg.PrimaryColor != "purple" {
		t.Errorf("expected purple, got %s", cfg.PrimaryColor)
	}
	if cfg.Path != "/sync" {
		t.Errorf("expected /sync, got %s", cfg.Path)
	}
}

func TestPanel_InjectConfig(t *testing.T) {
	p := NewPanel("inject-test").
		WithBrandName("InjectedApp").
		WithPrimaryColor("indigo")

	p.syncConfig()

	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		cfg := layouts.GetPanelConfigFromContext(r.Context())
		if cfg == nil {
			t.Error("expected PanelConfig in context, got nil")
			return
		}
		if cfg.Name != "InjectedApp" {
			t.Errorf("expected InjectedApp in context, got %s", cfg.Name)
		}
	})

	handler := p.injectConfig(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if !called {
		t.Error("inner handler was not called")
	}
}

func TestPanel_InjectNavGroups(t *testing.T) {
	p := NewPanel("nav-test")
	p.syncConfig()

	layouts.SetNavGroups([]layouts.NavGroup{
		{Label: "Main", Items: []layouts.NavItem{{Slug: "users", Label: "Users"}}},
	})

	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		groups := layouts.GetNavGroups(r.Context())
		if len(groups) != 1 {
			t.Errorf("expected 1 nav group in context, got %d", len(groups))
		}
	})

	handler := p.injectConfig(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if !called {
		t.Error("inner handler was not called")
	}
}

func TestPanel_WithMiddleware_Chain(t *testing.T) {
	p := NewPanel("protect-test")

	order := []string{}
	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw1")
			next.ServeHTTP(w, r)
		})
	}
	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw2")
			next.ServeHTTP(w, r)
		})
	}
	p.WithMiddleware(mw1, mw2)

	if len(p.Middlewares) != 2 {
		t.Errorf("expected 2 middlewares, got %d", len(p.Middlewares))
	}

	// Test the middleware chain directly (without protect which requires AuthManager)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "inner")
	})

	var h http.Handler = inner
	for i := len(p.Middlewares) - 1; i >= 0; i-- {
		h = p.Middlewares[i](h)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if len(order) != 3 {
		t.Errorf("expected [mw1, mw2, inner], got %v", order)
	}
	if order[0] != "mw1" || order[1] != "mw2" || order[2] != "inner" {
		t.Errorf("unexpected order: %v", order)
	}
}

func TestPanel_WithNavGroups_Context(t *testing.T) {
	groups := []layouts.NavGroup{
		{Label: "G1", Items: []layouts.NavItem{{Slug: "a", Label: "A"}}},
		{Label: "G2", Items: []layouts.NavItem{{Slug: "b", Label: "B"}}},
	}

	ctx := layouts.WithNavGroups(context.Background(), groups)
	got := layouts.GetNavGroups(ctx)

	if len(got) != 2 {
		t.Errorf("expected 2 groups, got %d", len(got))
	}
	if got[0].Label != "G1" {
		t.Errorf("expected G1, got %s", got[0].Label)
	}
}
