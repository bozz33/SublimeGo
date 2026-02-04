# Panel Configuration

This guide explains how to configure your SublimeGo admin panel.

## Configuration File

Configuration is managed via `config.yaml` at the project root.

## Basic Configuration

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 15s
  write_timeout: 15s

database:
  driver: "sqlite3"
  dsn: "file:sublimego.db?cache=shared&_fk=1"

engine:
  base_path: "/admin"
  brand_name: "My Admin Panel"
  items_per_page: 25

auth:
  session_lifetime: 24h
  cookie_name: "session"
```

## Panel Setup

```go
panel := engine.NewPanel("admin").
    SetPath("/admin").
    SetBrandName("My App").
    SetDatabase(db).
    SetAuthManager(auth).
    SetSession(session)
```

## Configuration Options

See `config.yaml` for all available options.
