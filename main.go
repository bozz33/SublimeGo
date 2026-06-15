// Command SublimeGo is a starter application for the SublimeAdmin framework.
// It boots a panel with authentication, a dashboard and an example resource,
// backed by an in-memory user repository so it runs with zero external setup.
//
// Run it:
//
//	go run .
//
// then open http://localhost:8080/ and sign in with admin@example.com / password.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/bozz33/sublimeadmin/engine"
	"github.com/bozz33/sublimeadmin/notifications"
	"github.com/bozz33/sublimeadmin/widget"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	db := openDB("sublimego.db")
	defer db.Close()

	users := newMemoryUserRepo()
	users.seed("Admin", "admin@example.com", "password")

	// Session + authentication. WithAuth wires the user repository, session and
	// auth manager in one call; we set an explicit session first to customize
	// its lifetime (WithAuth preserves an already-configured session).
	session := scs.New()
	session.Lifetime = 24 * time.Hour

	panel := engine.NewPanel("sublimego").
		WithPath("/").
		WithBrandName("SublimeGo").
		WithPrimaryColor("green").
		EnableNotifications(true).
		WithSession(session).
		WithAuth(users).
		// Demo multi-tenant switcher in the topbar, backed by a real cookie
		// resolver (see tenant.go). The active workspace persists across requests.
		WithTenantSwitcherList(tenantSwitcherEntries()...)

	// In production behind HTTPS, mark all cookies Secure. Left off by default
	// so cookies still work over plain HTTP during local development.
	if os.Getenv("SUBLIMEGO_SECURE_COOKIES") == "1" {
		panel.EnableSecureCookies()
	}

	// Dashboard widgets.
	panel.WithWidgets(
		widget.NewProvider("overview").WithWidgets(func(context.Context) []widget.Widget {
			return []widget.Widget{
				widget.NewStats(
					widget.Stat{Label: "Users", Value: "1"},
					widget.Stat{Label: "Sessions", Value: "0"},
					widget.Stat{Label: "Uptime", Value: "100%"},
				),
				// Polar-area chart (chart type added for Filament parity).
				widget.NewChart("posts-by-status", "Posts by status", widget.PolarArea).
					SetLabels([]string{"Draft", "Published"}).
					AddSeries("Draft", []int{42}).
					AddSeries("Published", []int{62}).
					WithDescription("Distribution of post statuses"),
			}
		}),
	)

	// Real, SQLite-backed resource (list/view/create/edit/delete).
	panel.AddResources(NewPostResource(db))

	// Built-in notification center page (All / Unread filter). Pages appear in
	// the sidebar automatically.
	panel.AddPages(engine.NewNotificationsPage())

	// Seed a couple of notifications for the admin user (id 1) so the center
	// and the topbar bell show content on first run.
	notifications.Success("Welcome to SublimeGo").SendTo("1")
	notifications.Info("Your starter app is running").SendTo("1")
	// Notification with a custom icon color and an action that also marks the
	// notification read when clicked.
	notifications.Warning("Review your draft posts").
		WithIconColor("primary").
		WithActions(notifications.Action("Open posts", "/posts").MarkAsRead()).
		SendTo("1")

	addr := ":8080"
	log.Printf("SublimeGo starter listening on http://localhost%s/ (login: admin@example.com / password)", addr)
	if err := http.ListenAndServe(addr, withTenancy(panel.Router())); err != nil {
		log.Fatal(err)
	}
}

// --- In-memory user repository (development only) -------------------------

type memoryUser struct {
	id       int
	name     string
	email    string
	password string // bcrypt hash
}

func (u *memoryUser) GetID() int          { return u.id }
func (u *memoryUser) GetName() string     { return u.name }
func (u *memoryUser) GetEmail() string    { return u.email }
func (u *memoryUser) GetPassword() string { return u.password }

type memoryUserRepo struct {
	users  map[string]*memoryUser // email -> user
	nextID int
}

func newMemoryUserRepo() *memoryUserRepo {
	return &memoryUserRepo{users: make(map[string]*memoryUser), nextID: 1}
}

func (r *memoryUserRepo) seed(name, email, password string) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	r.users[email] = &memoryUser{id: r.nextID, name: name, email: email, password: string(hash)}
	r.nextID++
}

func (r *memoryUserRepo) FindByEmail(_ context.Context, email string) (engine.FrameworkUser, error) {
	if u, ok := r.users[email]; ok {
		return u, nil
	}
	return nil, nil
}

func (r *memoryUserRepo) Create(_ context.Context, name, email, hashedPassword string) (engine.FrameworkUser, error) {
	u := &memoryUser{id: r.nextID, name: name, email: email, password: hashedPassword}
	r.users[email] = u
	r.nextID++
	return u, nil
}

func (r *memoryUserRepo) ExistsByEmail(_ context.Context, email string) (bool, error) {
	_, ok := r.users[email]
	return ok, nil
}

func (r *memoryUserRepo) ExistsByEmailExcluding(_ context.Context, email string, excludeID int) (bool, error) {
	u, ok := r.users[email]
	return ok && u.id != excludeID, nil
}

func (r *memoryUserRepo) UpdateNameEmail(_ context.Context, id int, name, email string) error {
	for _, u := range r.users {
		if u.id == id {
			delete(r.users, u.email)
			u.name, u.email = name, email
			r.users[email] = u
			return nil
		}
	}
	return nil
}

func (r *memoryUserRepo) UpdatePassword(_ context.Context, id int, hashedPassword string) error {
	for _, u := range r.users {
		if u.id == id {
			u.password = hashedPassword
			return nil
		}
	}
	return nil
}

func (r *memoryUserRepo) GetByID(_ context.Context, id int) (engine.FrameworkUser, error) {
	for _, u := range r.users {
		if u.id == id {
			return u, nil
		}
	}
	return nil, nil
}
