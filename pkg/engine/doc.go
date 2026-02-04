// Package engine is the core of the SublimeGo framework.
//
// It provides the main application engine that orchestrates all components
// including routing, middleware, resource management, and template rendering.
// The engine follows a resource-oriented architecture inspired by Laravel Filament.
//
// Features:
//   - HTTP server with standard library
//   - Resource registration and discovery
//   - Custom pages support (like Filament)
//   - Middleware stack management
//   - Template rendering with Templ
//   - Authentication handlers
//   - Panel configuration
//
// Basic usage:
//
//	panel := engine.NewPanel("admin").
//		SetDatabase(db).
//		SetAuthManager(authManager).
//		AddResources(&UserResource{}, &ProductResource{}).
//		AddPages(settingsPage, reportsPage)
//
//	http.ListenAndServe(":8080", panel.Router())
//
// Custom Pages:
//
// Pages are standalone views that don't follow the CRUD pattern.
// Use them for settings, reports, analytics, or any custom content.
//
//	settingsPage := engine.NewSimplePage("settings", "Settings", func(ctx context.Context, r *http.Request) templ.Component {
//		return views.SettingsPage()
//	}).WithIcon("settings").WithGroup("System")
//
//	panel.AddPages(settingsPage)
package engine
