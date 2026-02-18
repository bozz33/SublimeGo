package engine

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/alexedwards/scs/v2"
	"github.com/bozz33/sublimego/auth"
	"github.com/bozz33/sublimego/internal/ent"
	"github.com/bozz33/sublimego/middleware"
	"github.com/bozz33/sublimego/search"
	"github.com/bozz33/sublimego/ui/layouts"
	"github.com/bozz33/sublimego/views/dashboard"
	"github.com/bozz33/sublimego/widget"
)

// Panel represents a complete dashboard.
type Panel struct {
	ID          string
	Path        string
	BrandName   string
	DB          *ent.Client
	Resources   []Resource
	Pages       []Page
	AuthManager *auth.Manager
	Session     *scs.SessionManager
}

// NewPanel initializes an empty Panel.
func NewPanel(id string) *Panel {
	return &Panel{
		ID:        id,
		BrandName: "SublimeGo",
		Resources: make([]Resource, 0),
		Pages:     make([]Page, 0),
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

// AddPages adds custom pages to the panel.
// Pages are standalone views (reports, settings, analytics, etc.)
func (p *Panel) AddPages(pages ...Page) *Panel {
	p.Pages = append(p.Pages, pages...)
	p.registerNavItems()
	return p
}

// navItem is a unified type for navigation items (resources and pages)
type navItem struct {
	slug  string
	label string
	icon  string
	group string
	sort  int
}

// registerNavItems injects navigation items into the sidebar.
func (p *Panel) registerNavItems() {
	// Collect all nav items from resources and pages
	var allItems []navItem

	for _, r := range p.Resources {
		allItems = append(allItems, navItem{
			slug:  r.Slug(),
			label: r.PluralLabel(),
			icon:  r.Icon(),
			group: r.Group(),
			sort:  r.Sort(),
		})
	}

	for _, pg := range p.Pages {
		allItems = append(allItems, navItem{
			slug:  pg.Slug(),
			label: pg.Label(),
			icon:  pg.Icon(),
			group: pg.Group(),
			sort:  pg.Sort(),
		})
	}

	// Sort by sort order
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].sort < allItems[j].sort
	})

	// Group items by group name using stdlib
	grouped := make(map[string][]navItem)
	for _, item := range allItems {
		key := item.group
		if key == "" {
			key = "_root"
		}
		grouped[key] = append(grouped[key], item)
	}

	var navGroups []layouts.NavGroup

	if rootItems, ok := grouped["_root"]; ok {
		items := make([]layouts.NavItem, len(rootItems))
		for i, item := range rootItems {
			items[i] = layouts.NavItem{Slug: item.slug, Label: item.label, Icon: item.icon}
		}
		navGroups = append(navGroups, layouts.NavGroup{Label: "", Items: items})
	}

	// Sort group names for deterministic order
	groupNames := make([]string, 0, len(grouped))
	for k := range grouped {
		if k != "_root" {
			groupNames = append(groupNames, k)
		}
	}
	sort.Strings(groupNames)

	for _, groupName := range groupNames {
		groupItems := grouped[groupName]
		navItems := make([]layouts.NavItem, len(groupItems))
		for i, item := range groupItems {
			navItems[i] = layouts.NavItem{Slug: item.slug, Label: item.label, Icon: item.icon}
		}
		navGroups = append(navGroups, layouts.NavGroup{Label: groupName, Items: navItems})
	}

	layouts.SetNavGroups(navGroups)
}

// Router generates the standard HTTP Handler with automatic CRUD.
func (p *Panel) Router() http.Handler {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("ui/assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	if p.AuthManager != nil {
		authHandler := NewAuthHandler(p.AuthManager, p.DB)
		mux.Handle("/login", middleware.RequireGuest(p.AuthManager, "/")(authHandler))
		mux.Handle("/register", middleware.RequireGuest(p.AuthManager, "/")(authHandler))
		mux.Handle("/logout", authHandler)
	}

	dashboardHandler := middleware.RequireAuth(p.AuthManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use declarative widget providers instead of hardcoded providers
		widgets := widget.GetAllWidgets(r.Context())
		dashboard.Index(widgets).Render(r.Context(), w)
	}))
	mux.Handle("/", dashboardHandler)

	// Global search API endpoint
	mux.Handle("/api/search", middleware.RequireAuth(p.AuthManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]search.Result{})
			return
		}
		results, err := search.QuickSearch(r.Context(), query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})))

	for _, res := range p.Resources {
		handler := NewCRUDHandler(res)
		slug := res.Slug()
		protectedHandler := middleware.RequireAuth(p.AuthManager)(handler)
		mux.Handle("/"+slug+"/", protectedHandler)
		mux.Handle("/"+slug, protectedHandler)

		// Register relation manager sub-routes if the resource implements RelationManagerAware
		rmHandler := NewRelationManagerHandler(res)
		if rmHandler.HasManagers() {
			rmProtected := middleware.RequireAuth(p.AuthManager)(rmHandler)
			mux.Handle("/"+slug+"/relations/", rmProtected)
		}
	}

	// Register custom pages
	for _, pg := range p.Pages {
		pageHandler := NewPageHandler(pg)
		slug := pg.Slug()
		protectedHandler := middleware.RequireAuth(p.AuthManager)(pageHandler)
		mux.Handle("/"+slug, protectedHandler)
	}

	if p.Session != nil {
		return p.Session.LoadAndSave(mux)
	}

	return mux
}
