package engine

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/a-h/templ"
	authpkg "github.com/bozz33/sublimego/auth"
	"github.com/bozz33/sublimego/internal/ent"
	"github.com/bozz33/sublimego/internal/ent/user"
	authtemplates "github.com/bozz33/sublimego/views/auth"
)

// resetToken holds a password reset token with expiry.
type resetToken struct {
	email     string
	expiresAt time.Time
}

// resetStore is an in-memory token store (replace with DB for production).
var resetStore = struct {
	mu     sync.Mutex
	tokens map[string]resetToken
}{tokens: make(map[string]resetToken)}

// PasswordResetHandler handles /forgot-password and /reset-password.
type PasswordResetHandler struct {
	authManager *authpkg.Manager
	db          *ent.Client
}

// NewPasswordResetHandler creates a new password reset handler.
func NewPasswordResetHandler(authManager *authpkg.Manager, db *ent.Client) *PasswordResetHandler {
	return &PasswordResetHandler{authManager: authManager, db: db}
}

func (h *PasswordResetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/forgot-password":
		switch r.Method {
		case http.MethodGet:
			templ.Handler(authtemplates.ForgotPasswordPage("", "")).ServeHTTP(w, r)
		case http.MethodPost:
			h.handleForgotPassword(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/reset-password":
		switch r.Method {
		case http.MethodGet:
			token := r.URL.Query().Get("token")
			email := r.URL.Query().Get("email")
			templ.Handler(authtemplates.ResetPasswordPage(token, email, "")).ServeHTTP(w, r)
		case http.MethodPost:
			h.handleResetPassword(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.NotFound(w, r)
	}
}

func (h *PasswordResetHandler) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	if email == "" {
		templ.Handler(authtemplates.ForgotPasswordPage("Email is required.", "")).ServeHTTP(w, r)
		return
	}

	// Check user exists — always show success to prevent email enumeration
	exists, _ := h.db.User.Query().Where(user.EmailEQ(email)).Exist(r.Context())
	if exists {
		token := generateToken()
		resetStore.mu.Lock()
		resetStore.tokens[token] = resetToken{
			email:     email,
			expiresAt: time.Now().Add(1 * time.Hour),
		}
		resetStore.mu.Unlock()

		// TODO: send email with link: /reset-password?token=TOKEN&email=EMAIL
		// For now the token is logged — replace with your mailer
		_ = token
	}

	templ.Handler(authtemplates.ForgotPasswordPage("",
		"If that email exists, a reset link has been sent.")).ServeHTTP(w, r)
}

func (h *PasswordResetHandler) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	token := r.FormValue("token")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirm := r.FormValue("password_confirmation")

	showErr := func(msg string) {
		templ.Handler(authtemplates.ResetPasswordPage(token, email, msg)).ServeHTTP(w, r)
	}

	if password != confirm {
		showErr("Passwords do not match.")
		return
	}
	if len(password) < 8 {
		showErr("Password must be at least 8 characters.")
		return
	}

	resetStore.mu.Lock()
	entry, ok := resetStore.tokens[token]
	if ok {
		delete(resetStore.tokens, token)
	}
	resetStore.mu.Unlock()

	if !ok || entry.email != email || time.Now().After(entry.expiresAt) {
		showErr("This reset link is invalid or has expired.")
		return
	}

	ah := &AuthHandler{}
	newHash := ah.hashPassword(password)

	_, err := h.db.User.Update().
		Where(user.EmailEQ(email)).
		SetPassword(newHash).
		Save(r.Context())
	if err != nil {
		showErr("Failed to reset password. Please try again.")
		return
	}

	http.Redirect(w, r, "/login?reset=1", http.StatusFound)
}

// generateToken returns a cryptographically random 32-byte hex token.
func generateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
