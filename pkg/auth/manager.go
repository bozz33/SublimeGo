package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

const (
	sessionKeyUserID   = "_auth_user_id"
	sessionKeyUser     = "_auth_user"
	sessionKeyIntended = "_auth_intended_url"
)

// Manager handles user authentication and session management.
type Manager struct {
	session *scs.SessionManager
}

// NewManager creates an authentication manager with the given session store.
func NewManager(session *scs.SessionManager) *Manager {
	return &Manager{
		session: session,
	}
}

// Login authenticates a user and creates a new session.
func (m *Manager) Login(ctx context.Context, user *User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	// Renew session token for CSRF protection
	if err := m.session.RenewToken(ctx); err != nil {
		return fmt.Errorf("failed to renew session token: %w", err)
	}

	// Store only the user ID to keep session data minimal
	m.session.Put(ctx, sessionKeyUserID, user.ID)

	return nil
}

// LoginWithRequest is a convenience wrapper around Login using the request context.
func (m *Manager) LoginWithRequest(r *http.Request, user *User) error {
	return m.Login(r.Context(), user)
}

// Logout ends the user session and clears all auth data.
func (m *Manager) Logout(ctx context.Context) error {
	m.session.Remove(ctx, sessionKeyUserID)
	m.session.Remove(ctx, sessionKeyUser)
	m.session.Remove(ctx, sessionKeyIntended)
	if err := m.session.Destroy(ctx); err != nil {
		return fmt.Errorf("failed to destroy session: %w", err)
	}

	return nil
}

// LogoutWithRequest is a convenience wrapper around Logout using the request context.
func (m *Manager) LogoutWithRequest(r *http.Request) error {
	return m.Logout(r.Context())
}

// User retrieves the authenticated user from the session.
func (m *Manager) User(ctx context.Context) (*User, error) {
	userID := m.session.GetInt(ctx, sessionKeyUserID)
	if userID == 0 {
		return Guest(), nil
	}

	// Application must load full user data from database
	// For now, return a minimal user with just the ID
	return &User{ID: userID}, nil
}

// UserFromRequest retrieves the authenticated user from the request.
func (m *Manager) UserFromRequest(r *http.Request) (*User, error) {
	return m.User(r.Context())
}

// UserID returns the authenticated user's ID.
func (m *Manager) UserID(ctx context.Context) int {
	return m.session.GetInt(ctx, sessionKeyUserID)
}

// UserIDFromRequest returns the authenticated user's ID from the request.
func (m *Manager) UserIDFromRequest(r *http.Request) int {
	return m.UserID(r.Context())
}

// IsAuthenticated checks if a user is currently authenticated.
func (m *Manager) IsAuthenticated(ctx context.Context) bool {
	return m.UserID(ctx) > 0
}

// IsAuthenticatedFromRequest checks if a user is authenticated from the request.
func (m *Manager) IsAuthenticatedFromRequest(r *http.Request) bool {
	return m.IsAuthenticated(r.Context())
}

// IsGuest checks if the user is a guest (not authenticated).
func (m *Manager) IsGuest(ctx context.Context) bool {
	return !m.IsAuthenticated(ctx)
}

// Can checks if the authenticated user has the given permission.
func (m *Manager) Can(ctx context.Context, permission string) bool {
	user, err := m.User(ctx)
	if err != nil || user == nil {
		return false
	}

	return user.Can(permission)
}

// CanFromRequest checks if the user has permission from the request.
func (m *Manager) CanFromRequest(r *http.Request, permission string) bool {
	return m.Can(r.Context(), permission)
}

// Cannot checks if the user does NOT have the given permission.
func (m *Manager) Cannot(ctx context.Context, permission string) bool {
	return !m.Can(ctx, permission)
}

// HasRole checks if the user has the specified role.
func (m *Manager) HasRole(ctx context.Context, role string) bool {
	user, err := m.User(ctx)
	if err != nil || user == nil {
		return false
	}

	return user.HasRole(role)
}

// HasRoleFromRequest checks if the user has the role from the request.
func (m *Manager) HasRoleFromRequest(r *http.Request, role string) bool {
	return m.HasRole(r.Context(), role)
}

// IsAdmin checks if the user has admin privileges.
func (m *Manager) IsAdmin(ctx context.Context) bool {
	return m.HasRole(ctx, RoleAdmin) || m.HasRole(ctx, RoleSuperAdmin)
}

// SetIntendedURL stores the redirect URL for after login.
func (m *Manager) SetIntendedURL(ctx context.Context, url string) {
	m.session.Put(ctx, sessionKeyIntended, url)
}

// SetIntendedURLFromRequest stores the current request URL for post-login redirect.
func (m *Manager) SetIntendedURLFromRequest(r *http.Request) {
	m.SetIntendedURL(r.Context(), r.URL.String())
}

// IntendedURL retrieves and clears the intended redirect URL.
func (m *Manager) IntendedURL(ctx context.Context, defaultURL string) string {
	url := m.session.PopString(ctx, sessionKeyIntended)
	if url == "" {
		return defaultURL
	}
	return url
}

// IntendedURLFromRequest retrieves the intended URL from the request.
func (m *Manager) IntendedURLFromRequest(r *http.Request, defaultURL string) string {
	return m.IntendedURL(r.Context(), defaultURL)
}

// UpdateUser refreshes the user data in the session.
func (m *Manager) UpdateUser(ctx context.Context, user *User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	m.session.Put(ctx, sessionKeyUserID, user.ID)

	return nil
}

// UpdateUserFromRequest updates the user from the request.
func (m *Manager) UpdateUserFromRequest(r *http.Request, user *User) error {
	return m.UpdateUser(r.Context(), user)
}

// Session returns the underlying session manager.
func (m *Manager) Session() *scs.SessionManager {
	return m.session
}

// SessionManager returns the session manager (alias for Session).
func (m *Manager) SessionManager() *scs.SessionManager {
	return m.session
}

type contextKey string

const (
	userKey    contextKey = "auth_user"
	managerKey contextKey = "auth_manager"
)

// WithUser adds the user to the context.
func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext retrieves the user from the context.
func UserFromContext(ctx context.Context) *User {
	if user, ok := ctx.Value(userKey).(*User); ok {
		return user
	}
	return Guest()
}

// WithManager adds the auth manager to the context.
func WithManager(ctx context.Context, manager *Manager) context.Context {
	return context.WithValue(ctx, managerKey, manager)
}

// ManagerFromContext retrieves the auth manager from the context.
func ManagerFromContext(ctx context.Context) *Manager {
	if manager, ok := ctx.Value(managerKey).(*Manager); ok {
		return manager
	}
	return nil
}

// ManagerFromRequest retrieves the auth manager from the request.
func ManagerFromRequest(r *http.Request) *Manager {
	return ManagerFromContext(r.Context())
}

// CurrentUser retrieves the authenticated user from the request.
func CurrentUser(r *http.Request) *User {
	return UserFromContext(r.Context())
}

// IsAuthenticated checks if the request is authenticated.
func IsAuthenticated(r *http.Request) bool {
	return CurrentUser(r).IsAuthenticated()
}

// IsGuest checks if the request is from a guest user.
func IsGuest(r *http.Request) bool {
	return CurrentUser(r).IsGuest()
}

// Can checks if the request user has the given permission.
func Can(r *http.Request, permission string) bool {
	return CurrentUser(r).Can(permission)
}

// HasRole checks if the request user has the given role.
func HasRole(r *http.Request, role string) bool {
	return CurrentUser(r).HasRole(role)
}

// IsAdmin checks if the request user is an admin.
func IsAdmin(r *http.Request) bool {
	return CurrentUser(r).IsAdmin()
}
