package main

import (
	"net/http"

	"github.com/bozz33/sublimeadmin/engine"
)

// demoTenants is the static list of workspaces shown in the topbar switcher.
var demoTenants = []*engine.Tenant{
	{ID: "acme", Name: "Acme Inc."},
	{ID: "globex", Name: "Globex"},
}

// cookieTenantResolver resolves the active tenant from the "tenant" cookie,
// defaulting to the first tenant. It implements engine.TenantResolver.
type cookieTenantResolver struct{}

func (cookieTenantResolver) Resolve(r *http.Request) (*engine.Tenant, bool) {
	id := demoTenants[0].ID
	if c, err := r.Cookie("tenant"); err == nil && c.Value != "" {
		id = c.Value
	}
	for _, t := range demoTenants {
		if t.ID == id {
			return t, true
		}
	}
	return demoTenants[0], true
}

// tenantSwitcherEntries builds the topbar switcher entries pointing at the
// switch endpoint. The active entry is computed by the engine from the tenant
// resolved into the request context.
func tenantSwitcherEntries() []engine.TenantSwitchEntry {
	entries := make([]engine.TenantSwitchEntry, 0, len(demoTenants))
	for _, t := range demoTenants {
		entries = append(entries, engine.TenantSwitchEntry{
			ID:    t.ID,
			Label: t.Name,
			URL:   "/switch-tenant?id=" + t.ID,
		})
	}
	return entries
}

// withTenancy wraps the panel router with the switch endpoint and the tenant
// resolution middleware, so the active workspace persists across requests via a
// cookie. This is the real backend behind the topbar workspace switcher.
func withTenancy(panelHandler http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/switch-tenant", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		// Only persist a known tenant id; ignore anything else.
		for _, t := range demoTenants {
			if t.ID == id {
				http.SetCookie(w, &http.Cookie{Name: "tenant", Value: id, Path: "/", HttpOnly: true})
				break
			}
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
	mux.Handle("/", engine.TenantMiddleware(cookieTenantResolver{}, false)(panelHandler))
	return mux
}
