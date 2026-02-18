package engine

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/bozz33/sublimego/auth"
	"github.com/bozz33/sublimego/internal/ent"
	"github.com/bozz33/sublimego/middleware"
	"github.com/bozz33/sublimego/notifications"
	"github.com/bozz33/sublimego/plugin"
	"github.com/bozz33/sublimego/search"
	"github.com/bozz33/sublimego/ui/layouts"
	"github.com/bozz33/sublimego/views/dashboard"
	"github.com/bozz33/sublimego/widget"
)

// Panel represents a complete admin dashboard panel.
// Configure it fluently like Filament:
//
//	engine.NewPanel("admin").
//		SetPath("/admin").
//		SetBrandName("My App").
//		SetLogo("/assets/logo.svg").
//		SetPrimaryColor("blue").
//		EnableRegistration(false).
//		EnableNotifications(true)
type Panel struct {
	ID           string
	Path         string
	BrandName    string
	Logo         string
	Favicon      string
	PrimaryColor string // blue, green, red, purple, orange, pink, indigo
	DarkMode     bool

	Registration      bool
	EmailVerification bool
	PasswordReset     bool
	Profile           bool
	Notifications     bool

	DB          *ent.Client
	Resources   []Resource
	Pages       []Page
	AuthManager *auth.Manager
	Session     *scs.SessionManager

	// Custom middleware applied to all protected routes
	Middlewares []func(http.Handler) http.Handler
}

// NewPanel initializes a Panel with sensible defaults.
func NewPanel(id string) *Panel {
	return &Panel{
		ID:           id,
		BrandName:    "SublimeGo",
		PrimaryColor: "green",
		DarkMode:     false,

		Registration:      true,
		EmailVerification: false,
		PasswordReset:     true,
		Profile:           true,
		Notifications:     true,

		Resources: make([]Resource, 0),
		Pages:     make([]Page, 0),
	}
}

// Builder methods — Filament-style fluent API.

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

func (p *Panel) SetLogo(url string) *Panel {
	p.Logo = url
	return p
}

func (p *Panel) SetFavicon(url string) *Panel {
	p.Favicon = url
	return p
}

// SetPrimaryColor sets the UI accent color.
// Accepted values: "green", "blue", "red", "purple", "orange", "pink", "indigo"
func (p *Panel) SetPrimaryColor(color string) *Panel {
	p.PrimaryColor = color
	return p
}

func (p *Panel) SetDarkMode(enabled bool) *Panel {
	p.DarkMode = enabled
	return p
}

func (p *Panel) EnableRegistration(enabled bool) *Panel {
	p.Registration = enabled
	return p
}

func (p *Panel) EnableEmailVerification(enabled bool) *Panel {
	p.EmailVerification = enabled
	return p
}

func (p *Panel) EnablePasswordReset(enabled bool) *Panel {
	p.PasswordReset = enabled
	return p
}

func (p *Panel) EnableProfile(enabled bool) *Panel {
	p.Profile = enabled
	return p
}

func (p *Panel) EnableNotifications(enabled bool) *Panel {
	p.Notifications = enabled
	return p
}

// WithMiddleware adds custom middleware to all protected routes.
func (p *Panel) WithMiddleware(mw ...func(http.Handler) http.Handler) *Panel {
	p.Middlewares = append(p.Middlewares, mw...)
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

// syncConfig pushes Panel fields into the global layouts.PanelConfig.
// Called once at Router() time so all templates see the correct values.
func (p *Panel) syncConfig() {
	layouts.SetPanelConfig(&layouts.PanelConfig{
		Name:              p.BrandName,
		Path:              p.Path,
		Logo:              p.Logo,
		Favicon:           p.Favicon,
		PrimaryColor:      p.PrimaryColor,
		DarkMode:          p.DarkMode,
		Registration:      p.Registration,
		EmailVerification: p.EmailVerification,
		PasswordReset:     p.PasswordReset,
		Profile:           p.Profile,
		Notifications:     p.Notifications,
	})
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
// It also calls syncConfig() and plugin.BootAll() exactly once.
func (p *Panel) Router() http.Handler {
	// 1. Sync Panel fields -> global PanelConfig used by all templates
	p.syncConfig()

	// 2. Boot all registered plugins
	if err := plugin.Boot(); err != nil {
		panic("sublimego: plugin boot failed: " + err.Error())
	}

	mux := http.NewServeMux()

	// 3. Static assets with Cache-Control and gzip
	fs := http.FileServer(http.Dir("ui/assets"))
	mux.Handle("/assets/", gzipMiddleware(cacheControlMiddleware(http.StripPrefix("/assets/", fs))))

	// 4. Auth routes (conditional)
	if p.AuthManager != nil {
		authHandler := NewAuthHandler(p.AuthManager, p.DB)
		mux.Handle("/login", middleware.RequireGuest(p.AuthManager, "/")(authHandler))
		if p.Registration {
			mux.Handle("/register", middleware.RequireGuest(p.AuthManager, "/")(authHandler))
		}
		mux.Handle("/logout", authHandler)

		if p.Profile {
			profileHandler := NewProfileHandler(p.AuthManager, p.DB)
			mux.Handle("/profile", gzipMiddleware(p.protect(profileHandler)))
		}

		if p.PasswordReset {
			resetHandler := NewPasswordResetHandler(p.AuthManager, p.DB)
			mux.Handle("/forgot-password", resetHandler)
			mux.Handle("/reset-password", resetHandler)
		}
	}

	// 5. Dashboard
	dashboardHandler := p.protect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		widgets := widget.GetAllWidgets(r.Context())
		dashboard.Index(widgets).Render(r.Context(), w)
	}))
	mux.Handle("/", gzipMiddleware(dashboardHandler))

	// 6. Global search API
	mux.Handle("/api/search", p.protect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	// 7. Notifications API (conditional)
	if p.Notifications {
		userIDFunc := func(r *http.Request) string {
			if p.AuthManager != nil {
				if id := p.AuthManager.UserIDFromRequest(r); id > 0 {
					return fmt.Sprintf("%d", id)
				}
			}
			return ""
		}
		notifHandler := notifications.NewHandler(nil, userIDFunc)
		notifHandler.Register(mux, "/api/notifications")
	}

	// 8. Resources
	for _, res := range p.Resources {
		handler := NewCRUDHandler(res)
		slug := res.Slug()
		protectedHandler := p.protect(handler)
		mux.Handle("/"+slug+"/", gzipMiddleware(protectedHandler))
		mux.Handle("/"+slug, gzipMiddleware(protectedHandler))

		rmHandler := NewRelationManagerHandler(res)
		if rmHandler.HasManagers() {
			mux.Handle("/"+slug+"/relations/", p.protect(rmHandler))
		}
	}

	// 9. Custom pages
	for _, pg := range p.Pages {
		pageHandler := NewPageHandler(pg)
		slug := pg.Slug()
		mux.Handle("/"+slug, gzipMiddleware(p.protect(pageHandler)))
	}

	var handler http.Handler = mux
	if p.Session != nil {
		handler = p.Session.LoadAndSave(mux)
	}
	return handler
}

// protect wraps a handler with auth + any custom middlewares.
func (p *Panel) protect(h http.Handler) http.Handler {
	h = middleware.RequireAuth(p.AuthManager)(h)
	for i := len(p.Middlewares) - 1; i >= 0; i-- {
		h = p.Middlewares[i](h)
	}
	return h
}

// gzipMiddleware compresses responses when the client supports it.
func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}

// cacheControlMiddleware sets Cache-Control headers for static assets.
func cacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}
