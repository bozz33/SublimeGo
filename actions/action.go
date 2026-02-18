package actions

import (
	"fmt"
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

	RequiresConfirmation bool
	ModalTitle           string
	ModalDescription     string
	ConfirmLabel         string
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
	a.Type = Button
	return a
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
