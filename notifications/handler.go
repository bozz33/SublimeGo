package notifications

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Handler provides HTTP endpoints for the notification system.
//
// Routes to register:
//
//	GET  /notifications          -> list all notifications (JSON)
//	GET  /notifications/unread   -> list unread notifications (JSON)
//	GET  /notifications/stream   -> SSE stream of live notifications
//	POST /notifications/{id}/read -> mark one as read
//	POST /notifications/read-all  -> mark all as read
type Handler struct {
	store      *Store
	userIDFunc func(r *http.Request) string
}

// NewHandler creates a notification HTTP handler.
// userIDFunc extracts the authenticated user ID from the request.
func NewHandler(store *Store, userIDFunc func(r *http.Request) string) *Handler {
	if store == nil {
		store = globalStore
	}
	return &Handler{store: store, userIDFunc: userIDFunc}
}

// Register mounts all notification routes on the given mux.
func (h *Handler) Register(mux *http.ServeMux, prefix string) {
	if prefix == "" {
		prefix = "/notifications"
	}
	mux.HandleFunc(prefix, h.handleList)
	mux.HandleFunc(prefix+"/unread", h.handleUnread)
	mux.HandleFunc(prefix+"/stream", h.handleStream)
	mux.HandleFunc(prefix+"/read-all", h.handleReadAll)
	// /notifications/{id}/read — handled via prefix match
	mux.HandleFunc(prefix+"/", h.handleByID)
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := h.userIDFunc(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	items := h.store.GetAll(userID)
	writeJSON(w, map[string]any{
		"notifications": items,
		"unread_count":  h.store.UnreadCount(userID),
	})
}

func (h *Handler) handleUnread(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := h.userIDFunc(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	items := h.store.GetUnread(userID)
	writeJSON(w, map[string]any{
		"notifications": items,
		"unread_count":  len(items),
	})
}

// handleStream streams live notifications via Server-Sent Events.
func (h *Handler) handleStream(w http.ResponseWriter, r *http.Request) {
	userID := h.userIDFunc(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Send current unread count as first event
	unread := h.store.UnreadCount(userID)
	_, _ = fmt.Fprintf(w, "event: connected\ndata: {\"unread_count\": %d}\n\n", unread)
	flusher.Flush()

	ch := h.store.Subscribe(r.Context(), userID)

	for {
		select {
		case <-r.Context().Done():
			return
		case n, ok := <-ch:
			if !ok {
				return
			}
			msg, err := n.MarshalSSE()
			if err != nil {
				continue
			}
			_, _ = w.Write(msg)
			flusher.Flush()
		}
	}
}

func (h *Handler) handleReadAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := h.userIDFunc(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	h.store.MarkAllRead(userID)
	w.WriteHeader(http.StatusNoContent)
}

// handleByID handles /notifications/{id}/read
func (h *Handler) handleByID(w http.ResponseWriter, r *http.Request) {
	userID := h.userIDFunc(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path: /notifications/{id}/read
	path := r.URL.Path
	// strip leading /notifications/
	rest := path[len("/notifications/"):]
	// expect: {id}/read
	var notifID string
	if len(rest) > 5 && rest[len(rest)-5:] == "/read" {
		notifID = rest[:len(rest)-5]
	}

	if notifID == "" || r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	h.store.MarkRead(userID, notifID)
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
