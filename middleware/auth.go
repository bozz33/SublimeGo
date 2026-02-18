package middleware

import (
	"net/http"
	"time"

	"github.com/bozz33/sublimego/auth"
	"github.com/bozz33/sublimego/errors"
)

// AuthConfig configures the authentication middleware.
type AuthConfig struct {
	Manager         *auth.Manager
	RedirectURL     string
	ErrorHandler    *errors.Handler
	SaveIntendedURL bool
}

// DefaultAuthConfig returns a default configuration.
func DefaultAuthConfig(manager *auth.Manager) *AuthConfig {
	return &AuthConfig{
		Manager:         manager,
		RedirectURL:     "/login",
		SaveIntendedURL: true,
	}
}

// RequireAuth returns a middleware that requires authentication.
func RequireAuth(manager *auth.Manager) Middleware {
	return RequireAuthWithConfig(DefaultAuthConfig(manager))
}

// RequireAuthWithConfig returns an auth middleware with custom config.
func RequireAuthWithConfig(config *AuthConfig) Middleware {
	if config == nil || config.Manager == nil {
		panic("AuthConfig and Manager are required")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !config.Manager.IsAuthenticatedFromRequest(r) {
				if config.SaveIntendedURL && r.Method == "GET" {
					config.Manager.SetIntendedURLFromRequest(r)
				}

				http.Redirect(w, r, config.RedirectURL, http.StatusFound)
				return
			}

			user, err := config.Manager.UserFromRequest(r)
			if err != nil || user == nil {
				if config.ErrorHandler != nil {
					config.ErrorHandler.Handle(w, r, errors.Unauthorized("Authentication required"))
				} else {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				}
				return
			}

			ctx := auth.WithUser(r.Context(), user)
			ctx = auth.WithManager(ctx, config.Manager)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireGuest returns a middleware that requires an unauthenticated visitor.
func RequireGuest(manager *auth.Manager, redirectURL string) Middleware {
	if redirectURL == "" {
		redirectURL = "/"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if manager.IsAuthenticatedFromRequest(r) {
				http.Redirect(w, r, redirectURL, http.StatusFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission returns a middleware that checks a permission.
func RequirePermission(manager *auth.Manager, permission string) Middleware {
	return RequirePermissions(manager, permission)
}

// RequirePermissions returns a middleware that checks multiple permissions (all required).
func RequirePermissions(manager *auth.Manager, permissions ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !manager.IsAuthenticatedFromRequest(r) {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			user, err := manager.UserFromRequest(r)
			if err != nil || user == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if !user.HasAllPermissions(permissions...) {
				http.Error(w, "Forbidden - Permission denied", http.StatusForbidden)
				return
			}

			ctx := auth.WithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAnyPermission returns a middleware that checks at least one permission.
func RequireAnyPermission(manager *auth.Manager, permissions ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !manager.IsAuthenticatedFromRequest(r) {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			user, err := manager.UserFromRequest(r)
			if err != nil || user == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if !user.HasAnyPermission(permissions...) {
				http.Error(w, "Forbidden - Permission denied", http.StatusForbidden)
				return
			}

			ctx := auth.WithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns a middleware that checks a role.
func RequireRole(manager *auth.Manager, role string) Middleware {
	return RequireRoles(manager, role)
}

// RequireRoles returns a middleware that checks multiple roles (at least one required).
func RequireRoles(manager *auth.Manager, roles ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !manager.IsAuthenticatedFromRequest(r) {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			user, err := manager.UserFromRequest(r)
			if err != nil || user == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if !user.HasAnyRole(roles...) {
				http.Error(w, "Forbidden - Role required", http.StatusForbidden)
				return
			}

			ctx := auth.WithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin returns a middleware that requires the admin role.
func RequireAdmin(manager *auth.Manager) Middleware {
	return RequireRole(manager, auth.RoleAdmin)
}

// RequireSuperAdmin returns a middleware that requires the super admin role.
func RequireSuperAdmin(manager *auth.Manager) Middleware {
	return RequireRole(manager, auth.RoleSuperAdmin)
}

// LoadUser is a middleware that loads the user into the context (without forcing auth).
func LoadUser(manager *auth.Manager) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var user *auth.User

			if manager.IsAuthenticatedFromRequest(r) {
				loadedUser, err := manager.UserFromRequest(r)
				if err == nil && loadedUser != nil {
					user = loadedUser
				}
			}

			if user == nil {
				user = auth.Guest()
			}

			ctx := auth.WithUser(r.Context(), user)
			ctx = auth.WithManager(ctx, manager)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Auth is an alias for LoadUser (loads user without forcing auth).
func Auth(manager *auth.Manager) Middleware {
	return LoadUser(manager)
}

// Verified returns a middleware that requires a verified user.
func Verified(manager *auth.Manager, redirectURL string) Middleware {
	if redirectURL == "" {
		redirectURL = "/verify-email"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !manager.IsAuthenticatedFromRequest(r) {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			user, err := manager.UserFromRequest(r)
			if err != nil || user == nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			if verified := user.GetMetadata("email_verified"); verified != true {
				http.Redirect(w, r, redirectURL, http.StatusFound)
				return
			}

			ctx := auth.WithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ThrottleLogin limits login attempts per IP using the token bucket algorithm.
// maxAttempts is the number of allowed attempts per window duration.
func ThrottleLogin(maxAttempts int, window time.Duration) Middleware {
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	if window <= 0 {
		window = time.Minute
	}

	rpm := int(float64(maxAttempts) / window.Minutes())
	if rpm < 1 {
		rpm = 1
	}

	rl := NewRateLimiter(&RateLimitConfig{
		RequestsPerMinute: rpm,
		Burst:             maxAttempts,
		KeyFunc:           KeyByIP,
		CleanupInterval:   5 * time.Minute,
	})

	return rl.Middleware()
}
