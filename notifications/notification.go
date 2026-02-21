package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Level defines the severity of a notification.
type Level string

const (
	LevelInfo    Level = "info"
	LevelSuccess Level = "success"
	LevelWarning Level = "warning"
	LevelDanger  Level = "danger"
)

// Notification represents a single notification message.
type Notification struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Body        string    `json:"body,omitempty"`
	Level       Level     `json:"level"`
	Icon        string    `json:"icon,omitempty"`
	ActionURL   string    `json:"action_url,omitempty"`
	ActionLabel string    `json:"action_label,omitempty"`
	Duration    int       `json:"duration,omitempty"` // auto-dismiss ms (0 = persistent)
	Read        bool      `json:"read"`
	CreatedAt   time.Time `json:"created_at"`
}

// Store manages notifications for all users.
type Store struct {
	mu            sync.RWMutex
	notifications map[string][]*Notification      // userID -> notifications
	subscribers   map[string][]chan *Notification // userID -> SSE channels
	maxPerUser    int
}

var globalStore = &Store{
	notifications: make(map[string][]*Notification),
	subscribers:   make(map[string][]chan *Notification),
	maxPerUser:    100,
}

// NewStore creates a new notification store.
func NewStore(maxPerUser int) *Store {
	if maxPerUser <= 0 {
		maxPerUser = 100
	}
	return &Store{
		notifications: make(map[string][]*Notification),
		subscribers:   make(map[string][]chan *Notification),
		maxPerUser:    maxPerUser,
	}
}

// SetGlobalStore replaces the global store (useful for testing or custom config).
func SetGlobalStore(s *Store) {
	globalStore = s
}

// Send sends a notification to a user via the global store.
func Send(userID string, n *Notification) {
	globalStore.Send(userID, n)
}

// GetUnread returns unread notifications for a user via the global store.
func GetUnread(userID string) []*Notification {
	return globalStore.GetUnread(userID)
}

// GetAll returns all notifications for a user via the global store.
func GetAll(userID string) []*Notification {
	return globalStore.GetAll(userID)
}

// MarkRead marks a notification as read via the global store.
func MarkRead(userID, notifID string) {
	globalStore.MarkRead(userID, notifID)
}

// MarkAllRead marks all notifications as read for a user via the global store.
func MarkAllRead(userID string) {
	globalStore.MarkAllRead(userID)
}

// Subscribe returns a channel that receives new notifications for a user.
func Subscribe(ctx context.Context, userID string) <-chan *Notification {
	return globalStore.Subscribe(ctx, userID)
}

// UnreadCount returns the number of unread notifications for a user.
func UnreadCount(userID string) int {
	return globalStore.UnreadCount(userID)
}

// Send sends a notification to a user and broadcasts to SSE subscribers.
func (s *Store) Send(userID string, n *Notification) {
	if n.ID == "" {
		n.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now()
	}
	if n.Level == "" {
		n.Level = LevelInfo
	}
	n.UserID = userID

	s.mu.Lock()
	list := s.notifications[userID]
	list = append([]*Notification{n}, list...) // prepend (newest first)
	if len(list) > s.maxPerUser {
		list = list[:s.maxPerUser]
	}
	s.notifications[userID] = list

	// Broadcast to SSE subscribers
	subs := s.subscribers[userID]
	s.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- n:
		default:
			// subscriber too slow, skip
		}
	}
}

// GetUnread returns all unread notifications for a user (newest first).
func (s *Store) GetUnread(userID string) []*Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Notification
	for _, n := range s.notifications[userID] {
		if !n.Read {
			result = append(result, n)
		}
	}
	return result
}

// GetAll returns all notifications for a user (newest first).
func (s *Store) GetAll(userID string) []*Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := s.notifications[userID]
	result := make([]*Notification, len(list))
	copy(result, list)
	return result
}

// MarkRead marks a single notification as read.
func (s *Store) MarkRead(userID, notifID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, n := range s.notifications[userID] {
		if n.ID == notifID {
			n.Read = true
			return
		}
	}
}

// MarkAllRead marks all notifications as read for a user.
func (s *Store) MarkAllRead(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, n := range s.notifications[userID] {
		n.Read = true
	}
}

// UnreadCount returns the number of unread notifications.
func (s *Store) UnreadCount(userID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, n := range s.notifications[userID] {
		if !n.Read {
			count++
		}
	}
	return count
}

// Subscribe returns a channel that receives new notifications for a user.
// The channel is automatically cleaned up when ctx is cancelled.
func (s *Store) Subscribe(ctx context.Context, userID string) <-chan *Notification {
	ch := make(chan *Notification, 16)

	s.mu.Lock()
	s.subscribers[userID] = append(s.subscribers[userID], ch)
	s.mu.Unlock()

	go func() {
		<-ctx.Done()
		s.mu.Lock()
		defer s.mu.Unlock()

		subs := s.subscribers[userID]
		for i, sub := range subs {
			if sub == ch {
				s.subscribers[userID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		close(ch)
	}()

	return ch
}

// Builder helpers — fluent API for constructing notifications.

// Info creates an info-level notification.
func Info(title string) *Notification {
	return &Notification{Title: title, Level: LevelInfo, Icon: "info", Duration: 5000}
}

// Success creates a success-level notification.
func Success(title string) *Notification {
	return &Notification{Title: title, Level: LevelSuccess, Icon: "check_circle", Duration: 4000}
}

// Warning creates a warning-level notification.
func Warning(title string) *Notification {
	return &Notification{Title: title, Level: LevelWarning, Icon: "warning", Duration: 6000}
}

// Danger creates a danger-level notification.
func Danger(title string) *Notification {
	return &Notification{Title: title, Level: LevelDanger, Icon: "error", Duration: 0}
}

// WithBody sets the notification body.
func (n *Notification) WithBody(body string) *Notification {
	n.Body = body
	return n
}

// WithAction sets the action URL and label.
func (n *Notification) WithAction(label, url string) *Notification {
	n.ActionLabel = label
	n.ActionURL = url
	return n
}

// WithIcon overrides the default icon.
func (n *Notification) WithIcon(icon string) *Notification {
	n.Icon = icon
	return n
}

// WithDuration sets the auto-dismiss duration in milliseconds (0 = persistent).
func (n *Notification) WithDuration(ms int) *Notification {
	n.Duration = ms
	return n
}

// Persistent marks the notification as non-dismissible (duration = 0).
func (n *Notification) Persistent() *Notification {
	n.Duration = 0
	return n
}

// SendTo sends this notification to a user via the global store.
func (n *Notification) SendTo(userID string) {
	Send(userID, n)
}

// SendToAll broadcasts this notification to all users in the given list.
func (n *Notification) SendToAll(userIDs []string) {
	for _, id := range userIDs {
		copy := *n
		copy.ID = fmt.Sprintf("%d-%s", time.Now().UnixNano(), id)
		Send(id, &copy)
	}
}

// Broadcast sends a notification to all users currently tracked in the global store.
func Broadcast(n *Notification) {
	globalStore.Broadcast(n)
}

// Broadcast sends a notification to all users currently tracked in this store.
func (s *Store) Broadcast(n *Notification) {
	s.mu.RLock()
	userIDs := make([]string, 0, len(s.notifications))
	for id := range s.notifications {
		userIDs = append(userIDs, id)
	}
	s.mu.RUnlock()
	for _, id := range userIDs {
		copy := *n
		copy.ID = fmt.Sprintf("%d-%s", time.Now().UnixNano(), id)
		s.Send(id, &copy)
	}
}

// MarshalSSE formats the notification as a Server-Sent Events message.
func (n *Notification) MarshalSSE() ([]byte, error) {
	data, err := json.Marshal(n)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("event: notification\ndata: %s\n\n", data)), nil
}
