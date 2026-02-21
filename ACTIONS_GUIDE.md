# Actions Guide

> Complete reference for SublimeGo actions — header actions, bulk actions, modal actions, and lifecycle hooks.

---

## Overview

Actions are reusable operations that can be attached to resources. They support:

- **Header actions** — always-visible buttons in the table header
- **Bulk actions** — applied to selected rows
- **Row actions** — edit / delete / view per row
- **Modal actions** — confirmation dialogs via HTMX
- **Lifecycle hooks** — Before/After/OnSuccess/OnFailure callbacks

---

## Quick Start

```go
// Generate a new action
sublimego make:action PublishPost
```

This creates `internal/actions/publish_post_action.go` with a ready-to-use scaffold.

---

## Creating an Action

```go
import "github.com/bozz33/sublimego/actions"

var PublishAction = actions.New("publish").
    Label("Publish").
    Icon("publish").
    Color("primary").
    Handler(func(ctx context.Context, r *http.Request, item any) error {
        // your logic here
        return nil
    })
```

### Available Options

| Method | Description |
|--------|-------------|
| `Label(string)` | Display label |
| `Icon(string)` | Material Icons name |
| `Color(string)` | `primary`, `danger`, `success`, `warning` |
| `Handler(fn)` | The action function |
| `RequiresConfirmation(title, desc)` | Show a confirmation modal before executing |
| `WithForm(fields)` | Show a form modal before executing |
| `Authorize(fn)` | Gate the action behind a permission check |
| `RateLimit(n, duration)` | Limit execution frequency |
| `Method(string)` | HTTP method (`POST`, `DELETE`, `PATCH`) |
| `Redirect(url)` | Redirect after success |
| `Notification(fn)` | Custom success notification |

---

## Confirmation Modal

```go
var DeleteAllAction = actions.New("delete-all").
    Label("Delete All").
    Icon("delete_forever").
    Color("danger").
    RequiresConfirmation(
        "Delete all records?",
        "This action cannot be undone. All records will be permanently deleted.",
    ).
    Handler(func(ctx context.Context, r *http.Request, item any) error {
        return db.DeleteAll(ctx)
    })
```

---

## Modal with Form

```go
var BanUserAction = actions.New("ban").
    Label("Ban User").
    Icon("block").
    Color("danger").
    WithForm([]form.Field{
        form.Textarea("reason").Label("Reason").Required(),
        form.Select("duration").Label("Duration").Options([]form.SelectOption{
            {Value: "1d", Label: "1 day"},
            {Value: "7d", Label: "7 days"},
            {Value: "permanent", Label: "Permanent"},
        }),
    }).
    Handler(func(ctx context.Context, r *http.Request, item any) error {
        reason := r.FormValue("reason")
        duration := r.FormValue("duration")
        // ban user with reason + duration
        return nil
    })
```

---

## Lifecycle Hooks

```go
var ExportAction = actions.New("export").
    Label("Export CSV").
    Before(func(ctx context.Context, item any) error {
        log.Println("Starting export...")
        return nil
    }).
    After(func(ctx context.Context, item any) error {
        log.Println("Export complete.")
        return nil
    }).
    OnSuccess(func(ctx context.Context, item any) {
        notifications.Send(userID, notifications.Success("Export ready"))
    }).
    OnFailure(func(ctx context.Context, item any, err error) {
        notifications.Send(userID, notifications.Danger("Export failed").WithBody(err.Error()))
    }).
    Handler(func(ctx context.Context, r *http.Request, item any) error {
        return exportToCSV(ctx, item)
    })
```

---

## Attaching Actions to a Resource

### Header Actions (table toolbar)

```go
func (r *PostResource) HeaderActions() []engine.HeaderAction {
    return []engine.HeaderAction{
        {Label: "Publish All", Icon: "publish", URL: "/admin/posts/publish-all", Method: "POST", Color: "primary"},
        {Label: "Export", Icon: "download", URL: "/admin/posts/export", Color: "secondary"},
    }
}
```

### Bulk Actions (selected rows)

```go
func (r *PostResource) BulkActions() []engine.BulkActionDef {
    return []engine.BulkActionDef{
        {Label: "Delete Selected", Icon: "delete", URL: "/admin/posts/bulk-delete", Color: "danger"},
        {Label: "Publish Selected", Icon: "publish", URL: "/admin/posts/bulk-publish", Color: "primary"},
    }
}
```

---

## Authorization

```go
var AdminOnlyAction = actions.New("reset").
    Label("Reset Password").
    Authorize(func(ctx context.Context, item any) bool {
        user := auth.UserFromContext(ctx)
        return user.HasRole("admin")
    }).
    Handler(func(ctx context.Context, r *http.Request, item any) error {
        return resetPassword(ctx, item)
    })
```

---

## Rate Limiting

```go
var SendEmailAction = actions.New("send-email").
    Label("Send Email").
    RateLimit(5, time.Minute). // max 5 per minute
    Handler(func(ctx context.Context, r *http.Request, item any) error {
        return mailer.Send(ctx, item)
    })
```
