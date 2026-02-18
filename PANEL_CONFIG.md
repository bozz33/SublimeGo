# Panel Configuration

Complete reference for configuring a SublimeGo admin panel.

---

## Configuration File

Configuration is managed via `config/config.yaml` and loaded by `appconfig/loader.go` using Viper.

Priority order (highest to lowest):

1. Environment variables (prefixed with `SUBLIMEGO_`)
2. `config/config.yaml`
3. Built-in defaults

---

## Full Configuration Reference

```yaml
#  Server 
server:
  host: "0.0.0.0"          # Bind address
  port: 8080               # HTTP port
  read_timeout: 15s        # Max time to read request
  write_timeout: 15s       # Max time to write response
  idle_timeout: 60s        # Keep-alive timeout

#  Database 
database:
  driver: "sqlite"         # sqlite | postgres | mysql
  url: "file:sublimego.db?cache=shared&_fk=1"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

#  Engine (Panel) 
engine:
  base_path: "/admin"      # URL prefix for the admin panel
  brand_name: "My Admin"   # Displayed in the sidebar header
  brand_logo: ""           # Optional path to a logo image
  items_per_page: 25       # Default pagination size
  date_format: "2006-01-02"
  datetime_format: "2006-01-02 15:04"

#  Authentication 
auth:
  session_lifetime: 24h
  cookie_name: "sublimego_session"
  cookie_secure: true      # Set false for local HTTP development
  cookie_same_site: "lax"  # strict | lax | none
  bcrypt_cost: 12          # 1014 recommended
  login_throttle: 5        # Max failed attempts before lockout
  lockout_duration: 15m

#  Logger 
logger:
  level: "info"            # debug | info | warn | error
  format: "json"           # json | text
  output: "stdout"         # stdout | stderr | /path/to/file.log
  max_size_mb: 100         # Log rotation: max file size
  max_backups: 3           # Log rotation: number of backups
  max_age_days: 28         # Log rotation: max age

#  Jobs 
jobs:
  workers: 5               # Concurrent job workers
  poll_interval: 5s        # How often to poll for new jobs
  retention: 72h           # How long to keep completed jobs

#  Mailer 
mailer:
  host: "smtp.example.com"
  port: 587
  username: ""
  password: ""             # Use SUBLIMEGO_MAILER_PASSWORD env var
  from: "admin@example.com"
  from_name: "SublimeGo Admin"
  tls: true
```

---

## Environment Variables

Every config key maps to an environment variable using the `SUBLIMEGO_` prefix and `_` as separator:

| Config key | Environment variable |
|------------|---------------------|
| `server.port` | `SUBLIMEGO_SERVER_PORT` |
| `database.url` | `SUBLIMEGO_DATABASE_URL` |
| `auth.bcrypt_cost` | `SUBLIMEGO_AUTH_BCRYPT_COST` |
| `mailer.password` | `SUBLIMEGO_MAILER_PASSWORD` |

---

## Panel Setup in Code

```go
package main

import (
    "net/http"

    "github.com/alexedwards/scs/v2"
    "github.com/bozz33/sublimego/auth"
    "github.com/bozz33/sublimego/engine"
    "github.com/bozz33/sublimego/notifications"
)

func main() {
    // Session manager
    session := scs.New()
    session.Lifetime = 24 * time.Hour

    // Auth manager
    authManager := auth.NewManager(session)

    // Notification store (use DatabaseStore for production)
    notifStore := notifications.NewStore(100)

    // Panel
    panel := engine.NewPanel("admin").
        WithPath("/admin").
        WithBrandName("My Application").
        WithDatabase(db).
        WithAuthManager(authManager).
        WithSession(session).
        WithNotificationStore(notifStore)

    // Register resources
    panel.AddResources(
        NewProductResource(db),
        NewUserResource(db),
    )

    // Register custom pages
    panel.AddPages(
        NewDashboardPage(),
        NewSettingsPage(),
    )

    http.ListenAndServe(":8080", panel.Router())
}
```

---

## Multi-Tenancy Setup

```go
// Subdomain-based: acme.example.com, globex.example.com
resolver := engine.NewSubdomainResolver("example.com").
    Register(&engine.Tenant{ID: "acme",   Subdomain: "acme",   Name: "Acme Corp"}).
    Register(&engine.Tenant{ID: "globex", Subdomain: "globex", Name: "Globex Inc"})

router := engine.NewMultiPanelRouter(resolver).
    RegisterPanel("acme",   acmePanel).
    RegisterPanel("globex", globexPanel)

http.ListenAndServe(":8080", engine.TenantMiddleware(resolver, true)(router))
```

---

## MFA Setup

```go
mfa := auth.NewMFA(auth.DefaultMFAConfig())

// Generate a secret for a user
secret, err := mfa.GenerateSecret()

// Get the provisioning URI for QR code generation
uri := mfa.ProvisioningURI(secret, user.Email)

// Validate a TOTP code submitted by the user
if !mfa.Validate(secret, submittedCode) {
    // invalid code
}

// Generate recovery codes
codes, err := mfa.GenerateRecoveryCodes()
auth.SetRecoveryCodes(user, codes)

// Use a recovery code (consumes it)
if auth.UseRecoveryCode(user, submittedCode) {
    // valid recovery code, now consumed
}
```

---

## Render Hooks Setup

```go
import "github.com/bozz33/sublimego/hooks"

// Inject a custom banner after the main content
hooks.Register(hooks.AfterContent, func(ctx context.Context) templ.Component {
    return views.MaintenanceBanner()
})

// Inject custom CSS/JS in <head>
hooks.Register(hooks.InHead, func(ctx context.Context) templ.Component {
    return views.CustomHeadAssets()
})
```

---

## Plugin Setup

```go
import "github.com/bozz33/sublimego/plugin"

type AuditPlugin struct{}

func (p *AuditPlugin) Name() string { return "audit" }
func (p *AuditPlugin) Boot() error {
    // Register hooks, routes, etc.
    hooks.Register(hooks.AfterContent, auditBanner)
    return nil
}

func init() {
    plugin.Register(&AuditPlugin{})
}

// In main(), before starting the server
if err := plugin.Boot(); err != nil {
    log.Fatal(err)
}
```
