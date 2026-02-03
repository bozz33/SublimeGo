// Package registry provides resource registration and discovery.
//
// It maintains a central registry of all resources in the application,
// allowing for dynamic resource lookup, filtering, and grouping.
// The registry is used by the engine to build navigation and route handlers.
//
// Features:
//   - Resource registration
//   - Lookup by slug or type
//   - Filtering by group or capability
//   - Navigation item generation
//   - Resource metadata access
//
// Basic usage:
//
//	reg := registry.New()
//
//	// Register resources
//	reg.Register(&UserResource{})
//	reg.Register(&ProductResource{})
//
//	// Lookup
//	resource := reg.Get("users")
//
//	// Get all resources
//	all := reg.All()
//
//	// Filter by group
//	adminResources := reg.ByGroup("Administration")
package registry
