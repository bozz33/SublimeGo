package actions

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ActionType defines the action style.
type ActionType string

const (
	Link   ActionType = "link"
	Button ActionType = "button"
)

// Action defines a possible interaction on a resource.
type Action struct {
	Name        string
	Label       string
	Icon        string
	Type        ActionType
	Method      string
	Color       string
	UrlResolver func(item any) string

	// Confirmation modal
	RequiresConfirmation bool
	ModalTitle           string
	ModalDescription     string
	ConfirmLabel         string
	CancelLabel          string

	// Lifecycle hooks
	BeforeFunc    func(ctx context.Context, item any) error
	AfterFunc     func(ctx context.Context, item any) error
	OnSuccessFunc func(ctx context.Context, item any)
	OnFailureFunc func(ctx context.Context, item any, err error)

	// Authorization
	AuthorizeFunc func(ctx context.Context, item any) bool

	// Rate limiting
	RateLimitMax    int           // max calls per window (0 = disabled)
	RateLimitWindow time.Duration // window duration

	// Notification
	SuccessMessage string
	FailureMessage string

	// Redirect
	RedirectURL      string                // static redirect after action; empty = back to list
	RedirectResolver func(item any) string // dynamic redirect
}

// New creates a new base action.
func New(name string) *Action {
	return &Action{
		Name:   name,
		Label:  name,
		Type:   Link,
		Method: "GET",
		Color:  "gray",
	}
}

// SetLabel sets the label.
func (a *Action) SetLabel(label string) *Action {
	a.Label = label
	return a
}

// SetIcon sets the icon.
func (a *Action) SetIcon(icon string) *Action {
	a.Icon = icon
	return a
}

// SetColor sets the color.
func (a *Action) SetColor(color string) *Action {
	a.Color = color
	return a
}

// SetUrl sets the URL resolver.
func (a *Action) SetUrl(resolver func(item any) string) *Action {
	a.UrlResolver = resolver
	return a
}

// RequiresDialog enables the confirmation modal.
func (a *Action) RequiresDialog(title, desc string) *Action {
	a.RequiresConfirmation = true
	a.ModalTitle = title
	a.ModalDescription = desc
	a.ConfirmLabel = "Confirm"
	a.CancelLabel = "Cancel"
	a.Type = Button
	return a
}

// WithConfirmLabels customises the confirm/cancel button labels.
func (a *Action) WithConfirmLabels(confirm, cancel string) *Action {
	a.ConfirmLabel = confirm
	a.CancelLabel = cancel
	return a
}

// Before registers a hook called before the action executes.
// Return a non-nil error to abort execution.
func (a *Action) Before(fn func(ctx context.Context, item any) error) *Action {
	a.BeforeFunc = fn
	return a
}

// After registers a hook called after the action executes (regardless of outcome).
func (a *Action) After(fn func(ctx context.Context, item any) error) *Action {
	a.AfterFunc = fn
	return a
}

// OnSuccess registers a callback invoked when the action succeeds.
func (a *Action) OnSuccess(fn func(ctx context.Context, item any)) *Action {
	a.OnSuccessFunc = fn
	return a
}

// OnFailure registers a callback invoked when the action fails.
func (a *Action) OnFailure(fn func(ctx context.Context, item any, err error)) *Action {
	a.OnFailureFunc = fn
	return a
}

// Authorize sets a function that returns false to hide/disable the action.
func (a *Action) Authorize(fn func(ctx context.Context, item any) bool) *Action {
	a.AuthorizeFunc = fn
	return a
}

// IsAuthorized returns true if the action is allowed for the given context and item.
func (a *Action) IsAuthorized(ctx context.Context, item any) bool {
	if a.AuthorizeFunc == nil {
		return true
	}
	return a.AuthorizeFunc(ctx, item)
}

// RateLimit sets a rate limit (max calls per window).
func (a *Action) RateLimit(max int, window time.Duration) *Action {
	a.RateLimitMax = max
	a.RateLimitWindow = window
	return a
}

// WithSuccessMessage sets the flash message shown on success.
func (a *Action) WithSuccessMessage(msg string) *Action {
	a.SuccessMessage = msg
	return a
}

// WithFailureMessage sets the flash message shown on failure.
func (a *Action) WithFailureMessage(msg string) *Action {
	a.FailureMessage = msg
	return a
}

// RedirectTo sets a static redirect URL after the action completes.
func (a *Action) RedirectTo(url string) *Action {
	a.RedirectURL = url
	return a
}

// RedirectWith sets a dynamic redirect URL resolver.
func (a *Action) RedirectWith(fn func(item any) string) *Action {
	a.RedirectResolver = fn
	return a
}

// ResolveRedirect returns the redirect URL for a given item.
func (a *Action) ResolveRedirect(item any) string {
	if a.RedirectResolver != nil {
		return a.RedirectResolver(item)
	}
	return a.RedirectURL
}

// Execute runs the action lifecycle: Before → handler → After → OnSuccess/OnFailure.
// The handler func encapsulates the actual HTTP action (e.g. delete, update).
func (a *Action) Execute(ctx context.Context, item any, handler func() error) error {
	if a.BeforeFunc != nil {
		if err := a.BeforeFunc(ctx, item); err != nil {
			if a.OnFailureFunc != nil {
				a.OnFailureFunc(ctx, item, err)
			}
			return fmt.Errorf("action %s before hook: %w", a.Name, err)
		}
	}

	err := handler()

	if a.AfterFunc != nil {
		_ = a.AfterFunc(ctx, item)
	}

	if err != nil {
		if a.OnFailureFunc != nil {
			a.OnFailureFunc(ctx, item, err)
		}
		return fmt.Errorf("action %s: %w", a.Name, err)
	}

	if a.OnSuccessFunc != nil {
		a.OnSuccessFunc(ctx, item)
	}
	return nil
}

// URL resolves the action URL for a given item.
func (a *Action) URL(item any) string {
	if a.UrlResolver != nil {
		return a.UrlResolver(item)
	}
	return ""
}

// ServeHTTP allows an Action to act as an http.Handler for standalone endpoints.
func (a *Action) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "action not configured as handler", http.StatusNotImplemented)
}

// EditAction creates a standard Edit button.
func EditAction(baseURL string) *Action {
	return New("edit").
		SetLabel("Edit").
		SetIcon("pencil").
		SetColor("primary").
		SetUrl(func(item any) string {
			return fmt.Sprintf("%s/%v/edit", baseURL, getItemID(item))
		})
}

// DeleteAction creates a Delete button with modal.
func DeleteAction(baseURL string) *Action {
	a := New("delete").
		SetLabel("Delete").
		SetIcon("trash").
		SetColor("danger").
		RequiresDialog("Delete this item?", "This action cannot be undone.").
		SetUrl(func(item any) string {
			return fmt.Sprintf("%s/%v", baseURL, getItemID(item))
		})

	a.Method = "DELETE"
	return a
}

// ViewAction creates a standard View button.
func ViewAction(baseURL string) *Action {
	return New("view").
		SetLabel("Voir").
		SetIcon("eye").
		SetColor("secondary").
		SetUrl(func(item any) string {
			return fmt.Sprintf("%s/%v", baseURL, getItemID(item))
		})
}

// Identifiable interface for entities with ID.
type Identifiable interface {
	GetID() int
}

// getItemID extracts the ID from an item robustly.
func getItemID(item any) string {
	if id, ok := item.(Identifiable); ok {
		return fmt.Sprintf("%d", id.GetID())
	}
	return fmt.Sprintf("%v", item)
}

// GetItemID is the exported version for external use.
func GetItemID(item any) string {
	return getItemID(item)
}

// CreateAction creates a standard Create button (header-level, no item context).
func CreateAction(baseURL string) *Action {
	return New("create").
		SetLabel("New").
		SetIcon("plus").
		SetColor("primary").
		SetUrl(func(_ any) string {
			return fmt.Sprintf("%s/create", baseURL)
		})
}

// ExportAction creates an Export button that triggers a CSV/Excel download.
// format: "csv" or "xlsx"
func ExportAction(baseURL string, format string) *Action {
	if format == "" {
		format = "csv"
	}
	return New("export").
		SetLabel("Export").
		SetIcon("arrow-down-tray").
		SetColor("gray").
		SetUrl(func(_ any) string {
			return fmt.Sprintf("%s/export?format=%s", baseURL, format)
		})
}

// ImportAction creates an Import button that opens the import form.
func ImportAction(baseURL string) *Action {
	return New("import").
		SetLabel("Import").
		SetIcon("arrow-up-tray").
		SetColor("gray").
		SetUrl(func(_ any) string {
			return fmt.Sprintf("%s/import", baseURL)
		})
}

// RestoreAction creates a Restore button for soft-deleted items.
func RestoreAction(baseURL string) *Action {
	a := New("restore").
		SetLabel("Restore").
		SetIcon("arrow-path").
		SetColor("success").
		RequiresDialog("Restore this item?", "The item will be restored and become active again.").
		SetUrl(func(item any) string {
			return fmt.Sprintf("%s/%v/restore", baseURL, getItemID(item))
		})
	a.Method = "POST"
	return a
}

// ForceDeleteAction creates a permanent delete button with strong confirmation.
func ForceDeleteAction(baseURL string) *Action {
	a := New("force-delete").
		SetLabel("Delete permanently").
		SetIcon("trash").
		SetColor("danger").
		RequiresDialog("Permanently delete?", "This action CANNOT be undone. The record will be deleted forever.").
		SetUrl(func(item any) string {
			return fmt.Sprintf("%s/%v/force-delete", baseURL, getItemID(item))
		})
	a.Method = "DELETE"
	return a
}
