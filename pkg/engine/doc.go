// Package engine is the core of the SublimeGo framework.
//
// It provides the main application engine that orchestrates all components
// including routing, middleware, resource management, and template rendering.
// The engine follows a resource-oriented architecture inspired by Laravel Filament.
//
// Features:
//   - HTTP server with Chi router
//   - Resource registration and discovery
//   - Middleware stack management
//   - Template rendering with Templ
//   - Authentication handlers
//   - Panel configuration
//
// Basic usage:
//
//	e := engine.New(engine.Config{
//		Port: 8080,
//		DB:   entClient,
//	})
//
//	// Register resources
//	e.RegisterResource(&UserResource{})
//	e.RegisterResource(&ProductResource{})
//
//	// Start server
//	e.Start()
package engine
