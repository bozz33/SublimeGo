package engine

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bozz33/sublimego/export"
	importer "github.com/bozz33/sublimego/import"
)

// ExportHandler serves CSV/Excel exports for a resource.
// Register it at e.g. GET /{slug}/export?format=csv|xlsx
type ExportHandler struct {
	resource Resource
	format   export.Format
}

// NewExportHandler creates an export handler for the given resource and format.
func NewExportHandler(r Resource, format export.Format) *ExportHandler {
	return &ExportHandler{resource: r, format: format}
}

// ServeHTTP streams the export file to the client.
func (h *ExportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	items, err := h.resource.List(r.Context())
	if err != nil {
		http.Error(w, "Failed to list items: "+err.Error(), http.StatusInternalServerError)
		return
	}

	format := h.format
	if q := r.URL.Query().Get("format"); q == "xlsx" {
		format = export.FormatExcel
	} else if q == "csv" {
		format = export.FormatCSV
	}

	filename := export.GenerateFilename(h.resource.Slug(), format)
	w.Header().Set("Content-Type", export.GetContentType(format))
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	exp := export.New(format).FromStructs(items)
	if err := exp.Write(w); err != nil {
		http.Error(w, "Export failed: "+err.Error(), http.StatusInternalServerError)
	}
}

// ResourceExportable is an optional interface for resources that support export.
// Implement it to customise headers and row data instead of using reflection.
type ResourceExportable interface {
	ExportHeaders() []string
	ExportRow(item any) []string
}

// ImportHandler handles CSV/Excel/JSON file uploads and delegates to the resource.
// Register it at e.g. GET+POST /{slug}/import
type ImportHandler struct {
	resource Resource
}

// NewImportHandler creates an import handler for the given resource.
func NewImportHandler(r Resource) *ImportHandler {
	return &ImportHandler{resource: r}
}

// ServeHTTP handles GET (show form) and POST (process upload).
func (h *ImportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.showForm(w, r)
	case http.MethodPost:
		h.handleUpload(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ImportHandler) showForm(w http.ResponseWriter, r *http.Request) {
	slug := h.resource.Slug()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html><html><body>
<h2>Import %s</h2>
<form method="POST" enctype="multipart/form-data">
  <input type="file" name="file" accept=".csv,.xlsx,.json" required />
  <button type="submit">Upload</button>
</form>
</body></html>`, slug)
}

func (h *ImportHandler) handleUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	// Resource must implement ResourceImportable to handle rows
	importable, ok := h.resource.(ResourceImportable)
	if !ok {
		http.Error(w, "This resource does not support import", http.StatusNotImplemented)
		return
	}

	imp := importer.New(importer.DefaultConfig())
	result, err := imp.ImportFromFile(r.Context(), file, header, importable.ImportRow)
	if err != nil {
		http.Error(w, "Import failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprintf(w, `<p>Import complete: %d success, %d errors, %d skipped.</p>
<a href="/%s">Back to list</a>`,
		result.SuccessCount, result.ErrorCount, result.SkippedCount, h.resource.Slug())
}

// ResourceImportable is an optional interface for resources that support import.
type ResourceImportable interface {
	ImportRow(ctx context.Context, row map[string]any) error
}
