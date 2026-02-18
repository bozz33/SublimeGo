package engine

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/alexedwards/scs/v2"
	"github.com/bozz33/sublimego/auth"
	"github.com/bozz33/sublimego/export"
	"github.com/bozz33/sublimego/internal/ent"
	"github.com/bozz33/sublimego/mailer"
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

	// Mailer is used for password reset emails. Defaults to LogMailer if nil.
	Mailer  mailer.Mailer
	BaseURL string // e.g. "https://example.com" — used to build reset links

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

// SetMailer sets the mailer used for password reset emails.
// Use mailer.NewSMTPMailer(cfg) for production, mailer.LogMailer{} for dev.
func (p *Panel) SetMailer(m mailer.Mailer) *Panel {
	p.Mailer = m
	return p
}

// SetBaseURL sets the public base URL of the panel (e.g. "https://example.com").
// Required for building password reset links in emails.
func (p *Panel) SetBaseURL(url string) *Panel {
	p.BaseURL = url
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
	allItems := p.collectNavItems()
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].sort < allItems[j].sort
	})
	layouts.SetNavGroups(groupNavItems(allItems))
}

// collectNavItems builds the flat list of nav items from resources and pages.
func (p *Panel) collectNavItems() []navItem {
	items := make([]navItem, 0, len(p.Resources)+len(p.Pages))
	for _, r := range p.Resources {
		items = append(items, navItem{
			slug: r.Slug(), label: r.PluralLabel(),
			icon: r.Icon(), group: r.Group(), sort: r.Sort(),
		})
	}
	for _, pg := range p.Pages {
		items = append(items, navItem{
			slug: pg.Slug(), label: pg.Label(),
			icon: pg.Icon(), group: pg.Group(), sort: pg.Sort(),
		})
	}
	return items
}

// groupNavItems groups sorted nav items into NavGroups (root first, then named groups).
func groupNavItems(allItems []navItem) []layouts.NavGroup {
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
		navGroups = append(navGroups, layouts.NavGroup{Label: "", Items: toNavItems(rootItems)})
	}

	groupNames := make([]string, 0, len(grouped))
	for k := range grouped {
		if k != "_root" {
			groupNames = append(groupNames, k)
		}
	}
	sort.Strings(groupNames)
	for _, name := range groupNames {
		navGroups = append(navGroups, layouts.NavGroup{Label: name, Items: toNavItems(grouped[name])})
	}
	return navGroups
}

// toNavItems converts internal navItem slice to layouts.NavItem slice.
func toNavItems(items []navItem) []layouts.NavItem {
	result := make([]layouts.NavItem, len(items))
	for i, item := range items {
		result[i] = layouts.NavItem{Slug: item.slug, Label: item.label, Icon: item.icon}
	}
	return result
}

// Router generates the standard HTTP Handler with automatic CRUD.
// It also calls syncConfig() and plugin.BootAll() exactly once.
func (p *Panel) Router() http.Handler {
	p.syncConfig()
	if err := plugin.Boot(); err != nil {
		panic("sublimego: plugin boot failed: " + err.Error())
	}
	mux := http.NewServeMux()
	p.registerStaticRoutes(mux)
	p.registerAuthRoutes(mux)
	p.registerCoreRoutes(mux)
	p.registerResourceRoutes(mux)
	p.registerPageRoutes(mux)
	var handler http.Handler = p.injectConfig(mux)
	if p.Session != nil {
		handler = p.Session.LoadAndSave(handler)
	}
	return SecurityHeadersMiddleware(handler)
}

func (p *Panel) registerStaticRoutes(mux *http.ServeMux) {
	fs := http.FileServer(http.Dir("ui/assets"))
	mux.Handle("/assets/", gzipMiddleware(cacheControlMiddleware(http.StripPrefix("/assets/", fs))))
}

func (p *Panel) registerAuthRoutes(mux *http.ServeMux) {
	if p.AuthManager == nil {
		return
	}
	authHandler := NewAuthHandler(p.AuthManager, p.DB)
	loginLimiter := middleware.NewRateLimiter(&middleware.RateLimitConfig{
		RequestsPerMinute: 5, Burst: 3, KeyFunc: middleware.KeyByIP,
	})
	mux.Handle("/login", middleware.RequireGuest(p.AuthManager, "/")(loginLimiter.Middleware()(authHandler)))
	mux.Handle("/logout", authHandler)
	if p.Registration {
		mux.Handle("/register", middleware.RequireGuest(p.AuthManager, "/")(authHandler))
	}
	if p.Profile {
		mux.Handle("/profile", gzipMiddleware(p.protect(NewProfileHandler(p.AuthManager, p.DB))))
	}
	if p.PasswordReset {
		rh := NewPasswordResetHandler(p.AuthManager, p.DB, p.Mailer, p.BaseURL)
		mux.Handle("/forgot-password", rh)
		mux.Handle("/reset-password", rh)
	}
}

func (p *Panel) registerCoreRoutes(mux *http.ServeMux) {
	// Dashboard
	mux.Handle("/", gzipMiddleware(p.protect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = dashboard.Index(widget.GetAllWidgets(r.Context())).Render(r.Context(), w)
	}))))
	// Global search
	mux.Handle("/api/search", p.protect(http.HandlerFunc(p.handleSearch)))
	// Notifications
	if p.Notifications {
		notifHandler := notifications.NewHandler(nil, func(r *http.Request) string {
			if p.AuthManager != nil {
				if id := p.AuthManager.UserIDFromRequest(r); id > 0 {
					return fmt.Sprintf("%d", id)
				}
			}
			return ""
		})
		notifHandler.Register(mux, "/api/notifications")
	}
}

func (p *Panel) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	w.Header().Set("Content-Type", "application/json")
	if query == "" {
		_ = json.NewEncoder(w).Encode([]search.Result{})
		return
	}
	results, err := search.QuickSearch(r.Context(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(results)
}

func (p *Panel) registerResourceRoutes(mux *http.ServeMux) {
	for _, res := range p.Resources {
		p.mountResource(mux, res)
	}
}

func (p *Panel) mountResource(mux *http.ServeMux, res Resource) {
	slug := res.Slug()
	h := gzipMiddleware(p.protect(NewCRUDHandler(res)))
	mux.Handle("/"+slug+"/", h)
	mux.Handle("/"+slug, h)
	mux.Handle("/"+slug+"/export", p.protect(NewExportHandler(res, export.FormatCSV)))
	if _, ok := res.(ResourceImportable); ok {
		mux.Handle("/"+slug+"/import", p.protect(NewImportHandler(res)))
	}
	if rm := NewRelationManagerHandler(res); rm.HasManagers() {
		mux.Handle("/"+slug+"/relations/", p.protect(rm))
	}
}

func (p *Panel) registerPageRoutes(mux *http.ServeMux) {
	for _, pg := range p.Pages {
		mux.Handle("/"+pg.Slug(), gzipMiddleware(p.protect(NewPageHandler(pg))))
	}
}

// EnableDebug mounts pprof at /debug/pprof/ (call before Router()).
// Protect with PprofAuthMiddleware in non-local environments.
func (p *Panel) EnableDebug(mux *http.ServeMux) {
	EnablePprof(mux)
}

// protect wraps a handler with auth + any custom middlewares.
func (p *Panel) protect(h http.Handler) http.Handler {
	h = middleware.RequireAuth(p.AuthManager)(h)
	for i := len(p.Middlewares) - 1; i >= 0; i-- {
		h = p.Middlewares[i](h)
	}
	return h
}

// injectConfig injects the Panel's PanelConfig and NavGroups into every request context.
// This enables multi-panel setups where each panel has its own config and navigation.
func (p *Panel) injectConfig(next http.Handler) http.Handler {
	cfg := layouts.GetPanelConfig()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := layouts.WithPanelConfig(r.Context(), cfg)
		ctx = layouts.WithNavGroups(ctx, layouts.GetNavGroups(ctx))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// gzipPool reuses gzip.Writer instances to avoid per-request allocations.
var gzipPool = sync.Pool{
	New: func() any {
		// BestSpeed never returns an error for valid levels.
		gz, err := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
		if err != nil {
			panic("gzip: " + err.Error())
		}
		return gz
	},
}

// gzipMiddleware compresses responses when the client supports it.
// Uses a sync.Pool to reuse gzip.Writer instances (zero allocation hot path).
func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")
		gz := gzipPool.Get().(*gzip.Writer) //nolint:errcheck
		gz.Reset(w)
		defer func() {
			_ = gz.Close()
			gzipPool.Put(gz)
		}()
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
