# SublimeGo — starter

A minimal starter application for the [SublimeAdmin](https://github.com/bozz33/sublime-admin)
Go admin-panel framework. It boots a panel with authentication, a dashboard and an example
resource, backed by an in-memory user repository so it runs with zero setup.

## Run

```bash
go run .
```

Open http://localhost:8080/ and sign in:

- **Email:** `admin@example.com`
- **Password:** `password`

## What you get

- Authentication (login / logout / register / password reset / profile)
- A dashboard with a stats widget
- An example `Post` resource (empty table — wire it to your database)
- Global search, notifications, dark mode

## Layout

```
main.go     panel setup + in-memory user repository (development)
go.mod      imports the framework; `replace` points to the local checkout
```

## Next steps

1. Replace the in-memory `memoryUserRepo` with a database-backed `engine.UserRepository`
   (e.g. Ent or database/sql).
2. Replace `engine.NewAutoResource("Post")` with real resources that fetch your data.
3. Add widgets, pages and custom actions. See the framework docs under
   `sublimego-core-refactor/docs/`.

## Development note

`go.mod` uses a `replace` directive pointing at `../sublimego-core-refactor` so changes to the
framework are picked up immediately. Remove it and pin a version once the framework is tagged.
