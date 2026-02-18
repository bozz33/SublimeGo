package notifications

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// Broadcaster manages WebSocket-like broadcast connections for real-time
// notification delivery. It uses Server-Sent Events (SSE) as the transport
// because SSE is natively supported by browsers, requires no external
// dependencies, and works through proxies/load-balancers without special
// configuration. For true WebSocket support, wrap this with nhooyr.io/websocket.
//
// Architecture:
//   - Each connected client gets a dedicated channel.
//   - Broadcast() fans out a notification to every subscriber of a given user.
//   - BroadcastAll() fans out to every connected client (system-wide alerts).
type Broadcaster struct {
	mu      sync.RWMutex
	clients map[string]map[*client]struct{} // userID -> set of clients
	logger  *slog.Logger
}

// client represents a single connected SSE client.
type client struct {
	ch     chan *Notification
	userID string
}

// NewBroadcaster creates a new Broadcaster.
func NewBroadcaster(logger *slog.Logger) *Broadcaster {
	if logger == nil {
		logger = slog.Default()
	}
	return &Broadcaster{
		clients: make(map[string]map[*client]struct{}),
		logger:  logger,
	}
}

// Subscribe registers a new client for the given user and returns a
// receive-only channel. The channel is closed when ctx is cancelled.
func (b *Broadcaster) Subscribe(ctx context.Context, userID string) <-chan *Notification {
	c := &client{
		ch:     make(chan *Notification, 32),
		userID: userID,
	}

	b.mu.Lock()
	if b.clients[userID] == nil {
		b.clients[userID] = make(map[*client]struct{})
	}
	b.clients[userID][c] = struct{}{}
	b.mu.Unlock()

	go func() {
		<-ctx.Done()
		b.mu.Lock()
		delete(b.clients[userID], c)
		if len(b.clients[userID]) == 0 {
			delete(b.clients, userID)
		}
		b.mu.Unlock()
		close(c.ch)
	}()

	return c.ch
}

// Broadcast sends a notification to all connected clients of a specific user.
func (b *Broadcaster) Broadcast(userID string, n *Notification) {
	b.mu.RLock()
	clients := b.clients[userID]
	b.mu.RUnlock()

	for c := range clients {
		select {
		case c.ch <- n:
		default:
			b.logger.Warn("broadcast: client channel full, dropping notification",
				slog.String("user_id", userID),
				slog.String("notification_id", n.ID),
			)
		}
	}
}

// BroadcastAll sends a notification to every connected client (system-wide).
func (b *Broadcaster) BroadcastAll(n *Notification) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, clients := range b.clients {
		for c := range clients {
			select {
			case c.ch <- n:
			default:
			}
		}
	}
}

// ConnectedUsers returns the number of users with at least one active connection.
func (b *Broadcaster) ConnectedUsers() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

// ConnectedClients returns the total number of active client connections.
func (b *Broadcaster) ConnectedClients() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	total := 0
	for _, clients := range b.clients {
		total += len(clients)
	}
	return total
}

// ServeSSE is an http.HandlerFunc that streams notifications via SSE.
// It expects userIDFunc to extract the authenticated user ID from the request.
func (b *Broadcaster) ServeSSE(userIDFunc func(r *http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := userIDFunc(r)
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

		ch := b.Subscribe(r.Context(), userID)

		// Send initial heartbeat so the client knows the connection is alive.
		writeSSEEvent(w, "connected", map[string]any{
			"time": time.Now().UTC().Format(time.RFC3339),
		})
		flusher.Flush()

		// Heartbeat ticker to keep the connection alive through proxies.
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case n, ok := <-ch:
				if !ok {
					return
				}
				data, err := json.Marshal(n)
				if err != nil {
					continue
				}
				writeSSERaw(w, "notification", data)
				flusher.Flush()
			case <-ticker.C:
				writeSSEEvent(w, "heartbeat", map[string]any{
					"time": time.Now().UTC().Format(time.RFC3339),
				})
				flusher.Flush()
			}
		}
	}
}

func writeSSEEvent(w http.ResponseWriter, event string, data any) {
	payload, _ := json.Marshal(data)
	writeSSERaw(w, event, payload)
}

func writeSSERaw(w http.ResponseWriter, event string, data []byte) {
	_, _ = w.Write([]byte("event: " + event + "\ndata: " + string(data) + "\n\n"))
}
