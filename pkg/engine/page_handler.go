package engine

import (
	"net/http"

	"github.com/bozz33/SublimeGo/pkg/ui/layouts"
)

// PageHandler handles HTTP requests for custom pages.
type PageHandler struct {
	page Page
}

// NewPageHandler creates a new handler for a custom page.
func NewPageHandler(page Page) *PageHandler {
	return &PageHandler{page: page}
}

// ServeHTTP handles the page request.
func (h *PageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check access permission
	if !h.page.CanAccess(ctx) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Render the page content
	content := h.page.Render(ctx, r)

	// Wrap in the base layout
	layouts.Page(h.page.Label(), content).Render(ctx, w)
}
