package notifications

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bozz33/sublimego/internal/ent"
	entnotif "github.com/bozz33/sublimego/internal/ent/notification"
)

// DatabaseStore is a persistent notification store backed by Ent.
// It keeps the in-memory SSE subscriber map for live streaming while
// persisting every notification to the database.
type DatabaseStore struct {
	db          *ent.Client
	mu          sync.RWMutex
	subscribers map[string][]chan *Notification
	maxPerUser  int
}

// NewDatabaseStore creates a DatabaseStore using the provided Ent client.
func NewDatabaseStore(db *ent.Client, maxPerUser int) *DatabaseStore {
	if maxPerUser <= 0 {
		maxPerUser = 100
	}
	return &DatabaseStore{
		db:          db,
		subscribers: make(map[string][]chan *Notification),
		maxPerUser:  maxPerUser,
	}
}

// Send persists a notification and broadcasts it to SSE subscribers.
func (s *DatabaseStore) Send(userID string, n *Notification) {
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

	ctx := context.Background()
	_, _ = s.db.Notification.Create().
		SetUserID(userID).
		SetTitle(n.Title).
		SetBody(n.Body).
		SetLevel(string(n.Level)).
		SetIcon(n.Icon).
		SetActionURL(n.ActionURL).
		SetActionLabel(n.ActionLabel).
		SetRead(false).
		SetCreatedAt(n.CreatedAt).
		Save(ctx)

	s.mu.RLock()
	subs := s.subscribers[userID]
	s.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- n:
		default:
		}
	}
}

// GetAll returns all notifications for a user (newest first).
func (s *DatabaseStore) GetAll(userID string) []*Notification {
	ctx := context.Background()
	rows, err := s.db.Notification.Query().
		Where(entnotif.UserID(userID)).
		Order(ent.Desc(entnotif.FieldCreatedAt)).
		Limit(s.maxPerUser).
		All(ctx)
	if err != nil {
		return nil
	}
	return entRowsToNotifications(rows)
}

// GetUnread returns unread notifications for a user (newest first).
func (s *DatabaseStore) GetUnread(userID string) []*Notification {
	ctx := context.Background()
	rows, err := s.db.Notification.Query().
		Where(entnotif.UserID(userID), entnotif.Read(false)).
		Order(ent.Desc(entnotif.FieldCreatedAt)).
		Limit(s.maxPerUser).
		All(ctx)
	if err != nil {
		return nil
	}
	return entRowsToNotifications(rows)
}

// MarkRead marks a single notification as read.
func (s *DatabaseStore) MarkRead(userID, notifID string) {
	ctx := context.Background()
	_ = s.db.Notification.Update().
		Where(entnotif.UserID(userID), entnotif.ID(mustParseID(notifID))).
		SetRead(true).
		Exec(ctx)
}

// MarkAllRead marks all notifications as read for a user.
func (s *DatabaseStore) MarkAllRead(userID string) {
	ctx := context.Background()
	_ = s.db.Notification.Update().
		Where(entnotif.UserID(userID), entnotif.Read(false)).
		SetRead(true).
		Exec(ctx)
}

// UnreadCount returns the number of unread notifications for a user.
func (s *DatabaseStore) UnreadCount(userID string) int {
	ctx := context.Background()
	count, err := s.db.Notification.Query().
		Where(entnotif.UserID(userID), entnotif.Read(false)).
		Count(ctx)
	if err != nil {
		return 0
	}
	return count
}

// Subscribe returns a channel that receives new notifications for a user.
func (s *DatabaseStore) Subscribe(ctx context.Context, userID string) <-chan *Notification {
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

func entRowsToNotifications(rows []*ent.Notification) []*Notification {
	out := make([]*Notification, len(rows))
	for i, r := range rows {
		out[i] = &Notification{
			ID:          fmt.Sprintf("%d", r.ID),
			UserID:      r.UserID,
			Title:       r.Title,
			Body:        r.Body,
			Level:       Level(r.Level),
			Icon:        r.Icon,
			ActionURL:   r.ActionURL,
			ActionLabel: r.ActionLabel,
			Read:        r.Read,
			CreatedAt:   r.CreatedAt,
		}
	}
	return out
}

func mustParseID(s string) int {
	var id int
	fmt.Sscanf(s, "%d", &id)
	return id
}
