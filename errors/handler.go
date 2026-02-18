package errors

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
)

// Handler manages HTTP errors.
type Handler struct {
	logger Logger

	errorPages map[int]templ.Component

	ShowStack bool

	defaultErrorPage templ.Component
}

// Logger is the minimal interface for logging.
type Logger interface {
	Error(msg string, attrs ...any)
	Warn(msg string, attrs ...any)
	Info(msg string, attrs ...any)
}

// HandlerOption configures the error handler.
type HandlerOption func(*Handler)

// WithLogger configures the logger.
func WithLogger(logger Logger) HandlerOption {
	return func(h *Handler) {
		h.logger = logger
	}
}

// WithShowStack enables stack trace display.
func WithShowStack(show bool) HandlerOption {
	return func(h *Handler) {
		h.ShowStack = show
	}
}

// WithErrorPage configures an error page for a status code.
func WithErrorPage(statusCode int, page templ.Component) HandlerOption {
	return func(h *Handler) {
		h.errorPages[statusCode] = page
	}
}

// WithDefaultErrorPage configures the default error page.
func WithDefaultErrorPage(page templ.Component) HandlerOption {
	return func(h *Handler) {
		h.defaultErrorPage = page
	}
}

// NewHandler creates a new error handler.
func NewHandler(opts ...HandlerOption) *Handler {
	h := &Handler{
		errorPages: make(map[int]templ.Component),
		ShowStack:  false,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Handle handles an error and returns the appropriate page.
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	appErr := ToAppError(err)

	h.logError(r, appErr)

	w.WriteHeader(appErr.StatusCode)

	errorPage := h.getErrorPage(appErr.StatusCode)
	if errorPage == nil {
		http.Error(w, appErr.Message, appErr.StatusCode)
		return
	}

	if err := errorPage.Render(r.Context(), w); err != nil {
		http.Error(w, appErr.Message, appErr.StatusCode)
	}
}

// HandleFunc returns a middleware that captures panics.
func (h *Handler) HandleFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				err := Internal(nil, "An error occurred")
				_ = err.WithField("panic", rec)
				h.Handle(w, r, err)
			}
		}()

		next(w, r)
	}
}

// logError records the error in logs.
func (h *Handler) logError(r *http.Request, appErr *AppError) {
	if h.logger == nil {
		return
	}

	attrs := []any{
		slog.String("code", appErr.Code),
		slog.Int("status", appErr.StatusCode),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
	}

	for k, v := range appErr.Fields {
		attrs = append(attrs, slog.Any(k, v))
	}

	if appErr.Err != nil {
		attrs = append(attrs, slog.Any("error", appErr.Err))
	}

	if h.ShowStack && appErr.Stack != "" {
		attrs = append(attrs, slog.String("stack", appErr.Stack))
	}

	switch {
	case appErr.StatusCode >= 500:
		h.logger.Error("server error", attrs...)
	case appErr.StatusCode >= 400:
		h.logger.Warn("client error", attrs...)
	default:
		h.logger.Info("request completed with error", attrs...)
	}
}

// getErrorPage returns the error page for a status code.
func (h *Handler) getErrorPage(statusCode int) templ.Component {
	if page, exists := h.errorPages[statusCode]; exists {
		return page
	}

	return h.defaultErrorPage
}

// Middleware returns a middleware that captures errors.
func (h *Handler) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					err := Internal(nil, "An error occurred")
					_ = err.WithField("panic", rec)
					h.Handle(w, r, err)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// NotFound returns a 404 handler.
func (h *Handler) NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.Handle(w, r, NotFound(""))
	}
}

// MethodNotAllowed returns a 405 handler.
func (h *Handler) MethodNotAllowed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := New("METHOD_NOT_ALLOWED", "HTTP method not allowed", http.StatusMethodNotAllowed)
		h.Handle(w, r, err)
	}
}

// Global instance
var defaultHandler = NewHandler()

// SetDefaultHandler configures the global handler.
func SetDefaultHandler(h *Handler) {
	defaultHandler = h
}

// Handle uses the global handler.
func Handle(w http.ResponseWriter, r *http.Request, err error) {
	defaultHandler.Handle(w, r, err)
}
