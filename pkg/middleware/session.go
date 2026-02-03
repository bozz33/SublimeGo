package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

// SessionConfig configures the session manager.
type SessionConfig struct {
	Lifetime       time.Duration
	IdleTimeout    time.Duration
	CookieName     string
	CookieDomain   string
	CookiePath     string
	CookieSecure   bool
	CookieHTTPOnly bool
	CookieSameSite http.SameSite
	Store          scs.Store
}

// DefaultSessionConfig returns a secure default configuration.
func DefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		Lifetime:       24 * time.Hour,
		IdleTimeout:    20 * time.Minute,
		CookieName:     "session_id",
		CookieDomain:   "",
		CookiePath:     "/",
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteLaxMode,
		Store:          nil,
	}
}

// DevSessionConfig returns a development configuration.
func DevSessionConfig() *SessionConfig {
	cfg := DefaultSessionConfig()
	cfg.CookieSecure = false
	cfg.Lifetime = 7 * 24 * time.Hour
	return cfg
}

// SessionManager wraps the SCS session manager.
type SessionManager struct {
	SessionManager *scs.SessionManager
	config         *SessionConfig
}

// NewSessionManager creates a new session manager.
func NewSessionManager(config *SessionConfig) *SessionManager {
	if config == nil {
		config = DefaultSessionConfig()
	}

	manager := scs.New()
	manager.Lifetime = config.Lifetime
	manager.IdleTimeout = config.IdleTimeout
	manager.Cookie.Name = config.CookieName
	manager.Cookie.Domain = config.CookieDomain
	manager.Cookie.Path = config.CookiePath
	manager.Cookie.Secure = config.CookieSecure
	manager.Cookie.HttpOnly = config.CookieHTTPOnly
	manager.Cookie.SameSite = config.CookieSameSite

	if config.Store != nil {
		manager.Store = config.Store
	}

	return &SessionManager{
		SessionManager: manager,
		config:         config,
	}
}

// Middleware returns a middleware that loads and saves sessions.
func (sm *SessionManager) Middleware() Middleware {
	return func(next http.Handler) http.Handler {
		return sm.SessionManager.LoadAndSave(next)
	}
}

// Session is an alias for Middleware.
func Session(sm *SessionManager) Middleware {
	return sm.Middleware()
}

// GetString retrieves a string value from the session.
func (sm *SessionManager) GetString(ctx context.Context, key string) string {
	return sm.SessionManager.GetString(ctx, key)
}

// GetInt retrieves an int value from the session.
func (sm *SessionManager) GetInt(ctx context.Context, key string) int {
	return sm.SessionManager.GetInt(ctx, key)
}

// GetBool retrieves a bool value from the session.
func (sm *SessionManager) GetBool(ctx context.Context, key string) bool {
	return sm.SessionManager.GetBool(ctx, key)
}

// Get retrieves a value from the session.
func (sm *SessionManager) Get(ctx context.Context, key string) any {
	return sm.SessionManager.Get(ctx, key)
}

// Put stores a value in the session.
func (sm *SessionManager) Put(ctx context.Context, key string, val any) {
	sm.SessionManager.Put(ctx, key, val)
}

// PutString stores a string in the session.
func (sm *SessionManager) PutString(ctx context.Context, key, val string) {
	sm.SessionManager.Put(ctx, key, val)
}

// PutInt stores an int in the session.
func (sm *SessionManager) PutInt(ctx context.Context, key string, val int) {
	sm.SessionManager.Put(ctx, key, val)
}

// PutBool stores a bool in the session.
func (sm *SessionManager) PutBool(ctx context.Context, key string, val bool) {
	sm.SessionManager.Put(ctx, key, val)
}

// Remove deletes a key from the session.
func (sm *SessionManager) Remove(ctx context.Context, key string) {
	sm.SessionManager.Remove(ctx, key)
}

// Pop retrieves and removes a value from the session.
func (sm *SessionManager) Pop(ctx context.Context, key string) any {
	return sm.SessionManager.Pop(ctx, key)
}

// PopString retrieves and removes a string from the session.
func (sm *SessionManager) PopString(ctx context.Context, key string) string {
	return sm.SessionManager.PopString(ctx, key)
}

// Clear removes all session data.
func (sm *SessionManager) Clear(ctx context.Context) error {
	return sm.SessionManager.Clear(ctx)
}

// Destroy completely destroys the session.
func (sm *SessionManager) Destroy(ctx context.Context) error {
	return sm.SessionManager.Destroy(ctx)
}

// RenewToken regenerates the session token for CSRF protection.
func (sm *SessionManager) RenewToken(ctx context.Context) error {
	return sm.SessionManager.RenewToken(ctx)
}

// Exists checks if a key exists in the session.
func (sm *SessionManager) Exists(ctx context.Context, key string) bool {
	return sm.SessionManager.Exists(ctx, key)
}

// Keys returns all session keys.
func (sm *SessionManager) Keys(ctx context.Context) []string {
	return sm.SessionManager.Keys(ctx)
}

// Iterate iterates over all key-value pairs in the session.
func (sm *SessionManager) Iterate(ctx context.Context, fn func(key string, value any) error) error {
	keys := sm.SessionManager.Keys(ctx)
	for _, key := range keys {
		value := sm.SessionManager.Get(ctx, key)
		if err := fn(key, value); err != nil {
			return err
		}
	}
	return nil
}

// Status returns the session status.
func (sm *SessionManager) Status(ctx context.Context) scs.Status {
	return sm.SessionManager.Status(ctx)
}

// GetToken returns the session token.
func (sm *SessionManager) GetToken(ctx context.Context) string {
	return sm.SessionManager.Token(ctx)
}

// Context key for storing the session manager.
type sessionManagerKey struct{}

// WithSessionManager adds the session manager to the context.
func WithSessionManager(ctx context.Context, sm *SessionManager) context.Context {
	return context.WithValue(ctx, sessionManagerKey{}, sm)
}

// SessionManagerFromContext retrieves the session manager from the context.
func SessionManagerFromContext(ctx context.Context) *SessionManager {
	if sm, ok := ctx.Value(sessionManagerKey{}).(*SessionManager); ok {
		return sm
	}
	return nil
}

// SessionManagerFromRequest retrieves the session manager from the request.
func SessionManagerFromRequest(r *http.Request) *SessionManager {
	return SessionManagerFromContext(r.Context())
}
