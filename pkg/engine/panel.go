package engine

import (
	"net/http"
	"sort"

	"github.com/alexedwards/scs/v2"
	"github.com/bozz33/SublimeGo/internal/ent"
	"github.com/bozz33/SublimeGo/internal/providers"
	"github.com/bozz33/SublimeGo/pkg/auth"
	"github.com/bozz33/SublimeGo/pkg/ui/layouts"
	"github.com/bozz33/SublimeGo/views/dashboard"
	"github.com/samber/lo"
)

// Panel represents a complete dashboard.
type Panel struct {
	ID          string
	Path        string
	BrandName   string
	DB          *ent.Client
	Resources   []Resource
	AuthManager *auth.Manager
	Session     *scs.SessionManager
}

// NewPanel initializes an empty Panel.
func NewPanel(id string) *Panel {
	return &Panel{
		ID:        id,
		BrandName: "SublimeGo",
		Resources: make([]Resource, 0),
	}
}

// Builder methods for fluent configuration.

func (p *Panel) SetPath(path string) *Panel {
	p.Path = path
	return p
}

func (p *Panel) SetDatabase(db *ent.Client) *Panel {
	p.DB = db
	return p
}

func (p *Panel) SetBrandName(name string) *Panel {
	p.BrandName = name
	return p
}

func (p *Panel) SetAuthManager(authManager *auth.Manager) *Panel {
	p.AuthManager = authManager
	return p
}

func (p *Panel) SetSession(session *scs.SessionManager) *Panel {
	p.Session = session
	return p
}

// AddResources adds a block of resources.
func (p *Panel) AddResources(rs ...Resource) *Panel {
	p.Resources = append(p.Resources, rs...)
	p.registerNavItems()
	return p
}

// registerNavItems injects navigation items into the sidebar.
func (p *Panel) registerNavItems() {
	sortedResources := make([]Resource, len(p.Resources))
	copy(sortedResources, p.Resources)
	sort.Slice(sortedResources, func(i, j int) bool {
		return sortedResources[i].Sort() < sortedResources[j].Sort()
	})
	grouped := lo.GroupBy(sortedResources, func(r Resource) string {
		group := r.Group()
		if group == "" {
			return "_root"
		}
		return group
	})

	var navGroups []layouts.NavGroup

	if rootItems, ok := grouped["_root"]; ok {
		items := lo.Map(rootItems, func(r Resource, _ int) layouts.NavItem {
			return layouts.NavItem{
				Slug:  r.Slug(),
				Label: r.PluralLabel(),
				Icon:  r.Icon(),
			}
		})
		navGroups = append(navGroups, layouts.NavGroup{
			Label: "",
			Items: items,
		})
	}

	for groupName, resources := range grouped {
		if groupName == "_root" {
			continue
		}
		items := lo.Map(resources, func(r Resource, _ int) layouts.NavItem {
			return layouts.NavItem{
				Slug:  r.Slug(),
				Label: r.PluralLabel(),
				Icon:  r.Icon(),
			}
		})
		navGroups = append(navGroups, layouts.NavGroup{
			Label: groupName,
			Items: items,
		})
	}

	layouts.SetNavGroups(navGroups)
}

// Router generates the standard HTTP Handler with automatic CRUD.
func (p *Panel) Router() http.Handler {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("pkg/ui/assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	if p.AuthManager != nil {
		authHandler := NewAuthHandler(p.AuthManager, p.DB)
		mux.Handle("/login", RequireGuest(p.AuthManager)(authHandler))
		mux.Handle("/register", RequireGuest(p.AuthManager)(authHandler))
		mux.Handle("/logout", authHandler)
	}

	dashboardHandler := RequireAuth(p.AuthManager, p.DB)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		widgets := providers.GetDashboardStats(r.Context(), p.DB)
		dashboard.Index(widgets).Render(r.Context(), w)
	}))
	mux.Handle("/", dashboardHandler)

	for _, res := range p.Resources {
		handler := NewCRUDHandler(res)
		slug := res.Slug()
		protectedHandler := RequireAuth(p.AuthManager, p.DB)(handler)
		mux.Handle("/"+slug+"/", protectedHandler)
		mux.Handle("/"+slug, protectedHandler)
	}

	if p.Session != nil {
		return p.Session.LoadAndSave(mux)
	}

	return mux
}
