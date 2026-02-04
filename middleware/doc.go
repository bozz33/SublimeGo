// Package middleware provides HTTP middleware components.
//
// It includes common middleware like CORS, rate limiting, authentication,
// and a composable middleware stack system. All middleware follow the
// standard http.Handler pattern.
//
// Features:
//   - CORS with configurable origins and methods
//   - Rate limiting with token bucket algorithm
//   - Authentication middleware
//   - Middleware stack composition
//   - Conditional middleware execution
//   - Path-based middleware filtering
//   - Response writer wrapper with status tracking
//
// Basic usage:
//
//	// Create middleware stack
//	stack := middleware.NewStack(
//		middleware.CORS(),
//		middleware.RateLimit(100), // 100 req/min
//		middleware.Auth(authManager),
//	)
//
//	// Apply to handler
//	handler := stack.Then(myHandler)
//
//	// Or use individual middleware
//	router.Use(middleware.CORSWithConfig(&middleware.CORSConfig{
//		AllowedOrigins: []string{"https://example.com"},
//		AllowedMethods: []string{"GET", "POST"},
//	}))
package middleware
