package actions

import (
	"context"
	"net/http"
)

// ModalSize defines the size of a modal dialog.
type ModalSize string

const (
	ModalSizeSM  ModalSize = "sm"
	ModalSizeMD  ModalSize = "md"
	ModalSizeLG  ModalSize = "lg"
	ModalSizeXL  ModalSize = "xl"
	ModalSizeFull ModalSize = "full"
)

// ModalAction is an action that opens a modal dialog.
// It extends Action with modal-specific configuration.
type ModalAction struct {
	*Action

	// Modal display
	Size        ModalSize
	CloseOnEsc  bool
	CloseOnBack bool

	// Form inside modal (optional)
	FormFields  []ModalField
	FormAction  string // POST URL for the form
	FormMethod  string // default "POST"

	// HTMX: load modal content from a URL
	HTMXLoadURL string // if set, modal content is fetched via HTMX
}

// ModalField describes a simple field rendered inside a modal form.
type ModalField struct {
	Name        string
	Label       string
	Type        string // "text", "email", "select", "textarea", "hidden"
	Placeholder string
	Required    bool
	Options     []ModalFieldOption
	Value       string
}

// ModalFieldOption is an option for a select field inside a modal.
type ModalFieldOption struct {
	Value string
	Label string
}

// NewModal creates a new ModalAction wrapping a base Action.
func NewModal(name string) *ModalAction {
	return &ModalAction{
		Action:      New(name),
		Size:        ModalSizeMD,
		CloseOnEsc:  true,
		CloseOnBack: true,
		FormMethod:  "POST",
	}
}

// WithSize sets the modal size.
func (m *ModalAction) WithSize(size ModalSize) *ModalAction {
	m.Size = size
	return m
}

// WithHTMXLoad sets a URL to load the modal content via HTMX.
func (m *ModalAction) WithHTMXLoad(url string) *ModalAction {
	m.HTMXLoadURL = url
	return m
}

// WithForm adds a simple inline form to the modal.
func (m *ModalAction) WithForm(action string, fields ...ModalField) *ModalAction {
	m.FormAction = action
	m.FormFields = fields
	return m
}

// SetLabel sets the label (fluent, returns *ModalAction).
func (m *ModalAction) SetLabel(label string) *ModalAction {
	m.Action.Label = label
	return m
}

// SetIcon sets the icon (fluent, returns *ModalAction).
func (m *ModalAction) SetIcon(icon string) *ModalAction {
	m.Action.Icon = icon
	return m
}

// SetColor sets the color (fluent, returns *ModalAction).
func (m *ModalAction) SetColor(color string) *ModalAction {
	m.Action.Color = color
	return m
}

// Authorize sets the authorization function (fluent, returns *ModalAction).
func (m *ModalAction) Authorize(fn func(ctx context.Context, item any) bool) *ModalAction {
	m.Action.AuthorizeFunc = fn
	return m
}

// OnSuccess sets the success callback (fluent, returns *ModalAction).
func (m *ModalAction) OnSuccess(fn func(ctx context.Context, item any)) *ModalAction {
	m.Action.OnSuccessFunc = fn
	return m
}

// WithSuccessMessage sets the flash message shown on success (fluent, returns *ModalAction).
func (m *ModalAction) WithSuccessMessage(msg string) *ModalAction {
	m.Action.SuccessMessage = msg
	return m
}

// ServeHTTP handles the modal form submission.
func (m *ModalAction) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	// Delegate to action execute with a no-op handler by default.
	// Callers should override by registering a proper HTTP handler.
	http.Error(w, "modal action handler not configured", http.StatusNotImplemented)
}

// ConfirmAction creates a pre-configured delete confirmation modal action.
func ConfirmAction(name, title, description string) *ModalAction {
	m := NewModal(name)
	m.Action.RequiresConfirmation = true
	m.Action.ModalTitle = title
	m.Action.ModalDescription = description
	m.Action.ConfirmLabel = "Confirm"
	m.Action.CancelLabel = "Cancel"
	m.Action.Type = Button
	return m
}

// DeleteModalAction creates a standard delete confirmation modal.
func DeleteModalAction(baseURL string) *ModalAction {
	m := ConfirmAction("delete", "Delete this item?", "This action cannot be undone.")
	m.Action.SetLabel("Delete").
		SetIcon("delete").
		SetColor("danger").
		SetUrl(func(item any) string {
			return baseURL + "/" + getItemID(item)
		})
	m.Action.Method = "DELETE"
	m.Action.SuccessMessage = "Item deleted successfully."
	m.Action.FailureMessage = "Failed to delete item."
	return m
}
