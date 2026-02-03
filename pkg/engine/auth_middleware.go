package engine

import (
	"net/http"

	"github.com/bozz33/SublimeGo/internal/ent"
	"github.com/bozz33/SublimeGo/internal/ent/user"
	"github.com/bozz33/SublimeGo/pkg/auth"
)

// RequireAuth creates a middleware that requires authentication.
func RequireAuth(authManager *auth.Manager, db *ent.Client, loginPath ...string) func(http.Handler) http.Handler {
	redirectPath := "/login"
	if len(loginPath) > 0 && loginPath[0] != "" {
		redirectPath = loginPath[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !authManager.IsAuthenticatedFromRequest(r) {
				http.Redirect(w, r, redirectPath, http.StatusFound)
				return
			}

			userID := authManager.UserIDFromRequest(r)
			if userID > 0 && db != nil {
				dbUser, err := db.User.Query().Where(user.IDEQ(userID)).Only(r.Context())
				if err == nil && dbUser != nil {
					authUser := &auth.User{
						ID:    dbUser.ID,
						Name:  dbUser.Name,
						Email: dbUser.Email,
					}
					ctx := auth.WithUser(r.Context(), authUser)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireGuest creates a middleware that requires the user to not be authenticated.
func RequireGuest(authManager *auth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authManager.IsAuthenticatedFromRequest(r) {
				http.Redirect(w, r, "/admin", http.StatusFound)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
