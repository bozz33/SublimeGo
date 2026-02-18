package engine

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/bozz33/sublimego/ui/layouts"
)

// CRUDHandler automatically handles CRUD operations for a resource.
type CRUDHandler struct {
	Resource Resource
}

// NewCRUDHandler creates a CRUD handler for a given resource.
func NewCRUDHandler(r Resource) *CRUDHandler {
	return &CRUDHandler{Resource: r}
}

// List displays the list of items.
// Active filters are extracted from query params (prefix "filter_") and
// injected into the context so BuildTableState / ListFiltered can use them.
func (h *CRUDHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract active filters: ?filter_status=active&filter_role=admin -> {"status":"active","role":"admin"}
	activeFilters := make(map[string]string)
	for key, vals := range r.URL.Query() {
		if strings.HasPrefix(key, "filter_") && len(vals) > 0 && vals[0] != "" {
			activeFilters[strings.TrimPrefix(key, "filter_")] = vals[0]
		}
	}
	if len(activeFilters) > 0 {
		ctx = context.WithValue(ctx, ContextKeyActiveFilters, activeFilters)
	}

	component := h.Resource.Table(ctx)
	render(w, r, h.Resource.PluralLabel(), component)
}

// Create displays the creation form.
func (h *CRUDHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !h.Resource.CanCreate(ctx) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	component := h.Resource.Form(ctx, nil)
	render(w, r, "Create "+h.Resource.Label(), component)
}

// View displays the read-only detail view (Infolist) for a resource.
// Only available if the resource implements ResourceViewable.
func (h *CRUDHandler) View(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	if !h.Resource.CanRead(ctx) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	viewable, ok := h.Resource.(ResourceViewable)
	if !ok {
		// Resource has no View — redirect to edit
		http.Redirect(w, r, fmt.Sprintf("/%s/%s/edit", h.Resource.Slug(), id), http.StatusSeeOther)
		return
	}

	item, err := h.Resource.Get(ctx, id)
	if err != nil || item == nil {
		http.NotFound(w, r)
		return
	}

	component := viewable.View(ctx, item)
	render(w, r, h.Resource.Label(), component)
}

// Edit displays the edit form.
func (h *CRUDHandler) Edit(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	item, err := h.Resource.Get(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	component := h.Resource.Form(ctx, item)
	render(w, r, "Edit "+h.Resource.Label(), component)
}

// Store handles creation.
func (h *CRUDHandler) Store(w http.ResponseWriter, r *http.Request) {
	if !h.Resource.CanCreate(r.Context()) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.Resource.Create(r.Context(), r); err != nil {
		http.Error(w, "Creation error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/"+h.Resource.Slug(), http.StatusSeeOther)
}

// Update handles updates.
func (h *CRUDHandler) Update(w http.ResponseWriter, r *http.Request, id string) {
	if !h.Resource.CanUpdate(r.Context()) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.Resource.Update(r.Context(), id, r); err != nil {
		http.Error(w, "Update error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/"+h.Resource.Slug(), http.StatusSeeOther)
}

// Delete handles deletion.
func (h *CRUDHandler) Delete(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	if !h.Resource.CanDelete(ctx) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.Resource.Delete(ctx, id); err != nil {
		http.Error(w, "Delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/"+h.Resource.Slug(), http.StatusSeeOther)
}

// BulkDelete handles bulk deletion.
func (h *CRUDHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !h.Resource.CanDelete(ctx) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form parsing error", http.StatusBadRequest)
		return
	}

	ids := r.Form["ids[]"]
	if len(ids) == 0 {
		http.Error(w, "No items selected", http.StatusBadRequest)
		return
	}

	if err := h.Resource.BulkDelete(ctx, ids); err != nil {
		http.Error(w, "Bulk delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/"+h.Resource.Slug(), http.StatusSeeOther)
}

// ServeHTTP implements http.Handler with automatic routing.
func (h *CRUDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/"+h.Resource.Slug())
	path = strings.TrimPrefix(path, "/")

	parts := strings.Split(path, "/")

	switch r.Method {
	case http.MethodGet:
		if path == "" || path == "/" {
			h.List(w, r)
		} else if path == "create" {
			h.Create(w, r)
		} else if len(parts) == 2 && parts[1] == "edit" {
			h.Edit(w, r, parts[0])
		} else if len(parts) == 1 && parts[0] != "" {
			h.View(w, r, parts[0])
		} else {
			http.NotFound(w, r)
		}

	case http.MethodPost:
		r.ParseForm()
		methodOverride := r.FormValue("_method")

		if methodOverride == "DELETE" && len(parts) >= 1 {
			h.Delete(w, r, parts[0])
			return
		}

		if path == "bulk-delete" {
			h.BulkDelete(w, r)
			return
		}

		if path == "" || path == "/" {
			h.Store(w, r)
		} else if len(parts) >= 1 {
			h.Update(w, r, parts[0])
		}

	case http.MethodDelete:
		if len(parts) >= 1 {
			h.Delete(w, r, parts[0])
		}
	}
}

// extractID extracts the numeric ID from a path.
func extractID(s string) (int, error) {
	return strconv.Atoi(s)
}

// render is a helper to display a component in the layout.
func render(w http.ResponseWriter, r *http.Request, title string, content templ.Component) {
	fullPage := layouts.Page(title, content)
	fullPage.Render(r.Context(), w)
}
