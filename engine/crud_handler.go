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

const contextKeyListQuery contextKey = "list_query"

// GetListQuery retrieves the ListQuery from context (set by CRUDHandler.List).
func GetListQuery(ctx context.Context) *ListQuery {
	if q, ok := ctx.Value(contextKeyListQuery).(*ListQuery); ok {
		return q
	}
	return nil
}

// CRUDHandler automatically handles CRUD operations for a resource.
type CRUDHandler struct {
	Resource Resource
}

// NewCRUDHandler creates a CRUD handler for a given resource.
func NewCRUDHandler(r Resource) *CRUDHandler {
	return &CRUDHandler{Resource: r}
}

// List displays the list of items.
// Extracts filter_*, search, sort, dir, page, per_page from query params
// and injects them into context as both ActiveFilters and ListQuery.
func (h *CRUDHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	// Build ListQuery from all relevant params
	lq := &ListQuery{
		Filters: make(map[string]string),
		Search:  q.Get("search"),
		SortKey: q.Get("sort"),
		SortDir: q.Get("dir"),
		Page:    1,
		PerPage: 20,
	}
	if lq.SortDir != "desc" {
		lq.SortDir = "asc"
	}
	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		lq.Page = p
	}
	if pp, err := strconv.Atoi(q.Get("per_page")); err == nil && pp > 0 && pp <= 200 {
		lq.PerPage = pp
	}
	for key, vals := range q {
		if strings.HasPrefix(key, "filter_") && len(vals) > 0 && vals[0] != "" {
			lq.Filters[strings.TrimPrefix(key, "filter_")] = vals[0]
		}
	}

	// Inject into context
	ctx = context.WithValue(ctx, contextKeyListQuery, lq)
	if len(lq.Filters) > 0 {
		ctx = context.WithValue(ctx, ContextKeyActiveFilters, lq.Filters)
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
		h.routeGET(w, r, path, parts)
	case http.MethodPost:
		h.routePOST(w, r, path, parts)
	case http.MethodDelete:
		if len(parts) >= 1 {
			h.Delete(w, r, parts[0])
		}
	}
}

// routeGET dispatches GET requests.
func (h *CRUDHandler) routeGET(w http.ResponseWriter, r *http.Request, path string, parts []string) {
	switch {
	case path == "" || path == "/":
		h.List(w, r)
	case path == "create":
		h.Create(w, r)
	case len(parts) == 2 && parts[1] == "edit":
		h.Edit(w, r, parts[0])
	case len(parts) == 1 && parts[0] != "":
		h.View(w, r, parts[0])
	default:
		http.NotFound(w, r)
	}
}

// routePOST dispatches POST requests (including _method override).
func (h *CRUDHandler) routePOST(w http.ResponseWriter, r *http.Request, path string, parts []string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if r.FormValue("_method") == "DELETE" && len(parts) >= 1 {
		h.Delete(w, r, parts[0])
		return
	}
	switch {
	case path == "bulk-delete":
		h.BulkDelete(w, r)
	case path == "" || path == "/":
		h.Store(w, r)
	case len(parts) >= 1:
		h.Update(w, r, parts[0])
	}
}

// render is a helper to display a component in the layout.
func render(w http.ResponseWriter, r *http.Request, title string, content templ.Component) {
	fullPage := layouts.Page(title, content)
	_ = fullPage.Render(r.Context(), w)
}
