package engine

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

// CSRFConfig contains the CSRF middleware configuration.
type CSRFConfig struct {
	TokenLength   int
	CookieName    string
	HeaderName    string
	FormFieldName string
	CookiePath    string
	CookieMaxAge  int
	Secure        bool
	SameSite      http.SameSite
}

// DefaultCSRFConfig returns the default configuration.
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		TokenLength:   32,
		CookieName:    "_csrf",
		HeaderName:    "X-CSRF-Token",
		FormFieldName: "_token",
		CookiePath:    "/",
		CookieMaxAge:  3600,
		Secure:        false,
		SameSite:      http.SameSiteLaxMode,
	}
}

// CSRFManager manages CSRF tokens.
type CSRFManager struct {
	config CSRFConfig
	tokens sync.Map
}

// NewCSRFManager creates a new CSRF manager.
func NewCSRFManager(config ...CSRFConfig) *CSRFManager {
	cfg := DefaultCSRFConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	return &CSRFManager{
		config: cfg,
	}
}

// GenerateToken generates a new CSRF token.
func (m *CSRFManager) GenerateToken() (string, error) {
	b := make([]byte, m.config.TokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(b)
	m.tokens.Store(token, time.Now().Add(time.Duration(m.config.CookieMaxAge)*time.Second))
	return token, nil
}

// ValidateToken checks if a token is valid.
func (m *CSRFManager) ValidateToken(token string) bool {
	if token == "" {
		return false
	}
	if exp, ok := m.tokens.Load(token); ok {
		if expTime, ok := exp.(time.Time); ok {
			if time.Now().Before(expTime) {
				return true
			}
			m.tokens.Delete(token)
		}
	}
	return false
}

// Middleware creates an HTTP middleware for CSRF protection.
func (m *CSRFManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			cookie, err := r.Cookie(m.config.CookieName)
			if err != nil || cookie.Value == "" {
				token, err := m.GenerateToken()
				if err == nil {
					http.SetCookie(w, &http.Cookie{
						Name:     m.config.CookieName,
						Value:    token,
						Path:     m.config.CookiePath,
						MaxAge:   m.config.CookieMaxAge,
						Secure:   m.config.Secure,
						HttpOnly: false,
						SameSite: m.config.SameSite,
					})
				}
			}
			next.ServeHTTP(w, r)
			return
		}

		token := r.Header.Get(m.config.HeaderName)
		if token == "" {
			token = r.FormValue(m.config.FormFieldName)
		}

		cookie, err := r.Cookie(m.config.CookieName)
		if err != nil || cookie.Value == "" {
			http.Error(w, "CSRF token missing", http.StatusForbidden)
			return
		}

		if token != cookie.Value {
			http.Error(w, "CSRF token mismatch", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetCSRFToken retrieves the CSRF token from the request.
func GetCSRFToken(r *http.Request, cookieName ...string) string {
	name := "_csrf"
	if len(cookieName) > 0 {
		name = cookieName[0]
	}
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// CSRFField returns the HTML for a hidden CSRF field.
func CSRFField(token string, fieldName ...string) string {
	name := "_token"
	if len(fieldName) > 0 {
		name = fieldName[0]
	}
	return `<input type="hidden" name="` + name + `" value="` + token + `">`
}
