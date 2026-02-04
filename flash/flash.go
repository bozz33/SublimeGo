package flash

import (
	"context"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/samber/lo"
)

const (
	TypeSuccess = "success"
	TypeError   = "error"
	TypeWarning = "warning"
	TypeInfo    = "info"
)

const sessionKey = "_flash_messages"

// Message represents a flash message.
type Message struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Title string `json:"title,omitempty"`
}

// NewMessage creates a new flash message.
func NewMessage(msgType, text string) *Message {
	return &Message{
		Type: msgType,
		Text: text,
	}
}

// WithTitle adds a title to the message.
func (m *Message) WithTitle(title string) *Message {
	m.Title = title
	return m
}

// Manager handles flash messages.
type Manager struct {
	session *scs.SessionManager
}

// NewManager creates a new flash message manager.
func NewManager(session *scs.SessionManager) *Manager {
	return &Manager{
		session: session,
	}
}

// Success adds a success message.
func (m *Manager) Success(ctx context.Context, text string) {
	m.Add(ctx, NewMessage(TypeSuccess, text))
}

// SuccessWithTitle adds a success message with a title.
func (m *Manager) SuccessWithTitle(ctx context.Context, title, text string) {
	m.Add(ctx, NewMessage(TypeSuccess, text).WithTitle(title))
}

// Error adds an error message.
func (m *Manager) Error(ctx context.Context, text string) {
	m.Add(ctx, NewMessage(TypeError, text))
}

// ErrorWithTitle adds an error message with a title.
func (m *Manager) ErrorWithTitle(ctx context.Context, title, text string) {
	m.Add(ctx, NewMessage(TypeError, text).WithTitle(title))
}

// Warning adds a warning message.
func (m *Manager) Warning(ctx context.Context, text string) {
	m.Add(ctx, NewMessage(TypeWarning, text))
}

// WarningWithTitle adds a warning message with a title.
func (m *Manager) WarningWithTitle(ctx context.Context, title, text string) {
	m.Add(ctx, NewMessage(TypeWarning, text).WithTitle(title))
}

// Info adds an info message.
func (m *Manager) Info(ctx context.Context, text string) {
	m.Add(ctx, NewMessage(TypeInfo, text))
}

// InfoWithTitle adds an info message with a title.
func (m *Manager) InfoWithTitle(ctx context.Context, title, text string) {
	m.Add(ctx, NewMessage(TypeInfo, text).WithTitle(title))
}

// Add adds a custom message.
func (m *Manager) Add(ctx context.Context, message *Message) {
	messages := m.getMessages(ctx)
	messages = append(messages, message)
	m.session.Put(ctx, sessionKey, messages)
}

// Get retrieves all messages without clearing them.
func (m *Manager) Get(ctx context.Context) []*Message {
	return m.getMessages(ctx)
}

// GetAndClear retrieves all messages and clears them.
func (m *Manager) GetAndClear(ctx context.Context) []*Message {
	messages := m.getMessages(ctx)
	m.Clear(ctx)
	return messages
}

// Clear removes all messages.
func (m *Manager) Clear(ctx context.Context) {
	m.session.Remove(ctx, sessionKey)
}

// Has checks if there are any messages.
func (m *Manager) Has(ctx context.Context) bool {
	return len(m.getMessages(ctx)) > 0
}

// HasType checks if there are messages of a specific type.
func (m *Manager) HasType(ctx context.Context, msgType string) bool {
	return lo.SomeBy(m.getMessages(ctx), func(msg *Message) bool {
		return msg.Type == msgType
	})
}

// GetByType retrieves messages of a specific type.
func (m *Manager) GetByType(ctx context.Context, msgType string) []*Message {
	return lo.Filter(m.getMessages(ctx), func(msg *Message, _ int) bool {
		return msg.Type == msgType
	})
}

// Count returns the number of messages.
func (m *Manager) Count(ctx context.Context) int {
	return len(m.getMessages(ctx))
}

// CountByType returns the number of messages of a specific type.
func (m *Manager) CountByType(ctx context.Context, msgType string) int {
	return len(m.GetByType(ctx, msgType))
}

// getMessages retrieves messages from the session.
func (m *Manager) getMessages(ctx context.Context) []*Message {
	data := m.session.Get(ctx, sessionKey)
	if data == nil {
		return []*Message{}
	}

	if messages, ok := data.([]*Message); ok {
		return messages
	}

	return []*Message{}
}

// Keep preserves messages for an additional request.
func (m *Manager) Keep(ctx context.Context) {
	// Messages are already in session, no action needed
}

// Reflash keeps all messages for the next request.
func (m *Manager) Reflash(ctx context.Context) {
	// Already implemented by default with SCS
}

// SuccessFromRequest adds a success message from the request.
func (m *Manager) SuccessFromRequest(r *http.Request, text string) {
	m.Success(r.Context(), text)
}

// ErrorFromRequest adds an error message from the request.
func (m *Manager) ErrorFromRequest(r *http.Request, text string) {
	m.Error(r.Context(), text)
}

// WarningFromRequest adds a warning message from the request.
func (m *Manager) WarningFromRequest(r *http.Request, text string) {
	m.Warning(r.Context(), text)
}

// InfoFromRequest adds an info message from the request.
func (m *Manager) InfoFromRequest(r *http.Request, text string) {
	m.Info(r.Context(), text)
}

// GetFromRequest retrieves messages from the request.
func (m *Manager) GetFromRequest(r *http.Request) []*Message {
	return m.Get(r.Context())
}

// GetAndClearFromRequest retrieves and clears messages from the request.
func (m *Manager) GetAndClearFromRequest(r *http.Request) []*Message {
	return m.GetAndClear(r.Context())
}

type contextKey string

const (
	managerKey  contextKey = "flash_manager"
	messagesKey contextKey = "flash_messages"
)

// WithManager adds the flash manager to the context.
func WithManager(ctx context.Context, manager *Manager) context.Context {
	return context.WithValue(ctx, managerKey, manager)
}

// ManagerFromContext retrieves the flash manager from the context.
func ManagerFromContext(ctx context.Context) *Manager {
	if manager, ok := ctx.Value(managerKey).(*Manager); ok {
		return manager
	}
	return nil
}

// ManagerFromRequest retrieves the flash manager from the request.
func ManagerFromRequest(r *http.Request) *Manager {
	return ManagerFromContext(r.Context())
}

// WithMessages adds messages to the context.
func WithMessages(ctx context.Context, messages []*Message) context.Context {
	return context.WithValue(ctx, messagesKey, messages)
}

// MessagesFromContext retrieves messages from the context.
func MessagesFromContext(ctx context.Context) []*Message {
	if messages, ok := ctx.Value(messagesKey).([]*Message); ok {
		return messages
	}
	return []*Message{}
}

// MessagesFromRequest retrieves messages from the request.
func MessagesFromRequest(r *http.Request) []*Message {
	return MessagesFromContext(r.Context())
}

// Success adds a success message.
func Success(r *http.Request, text string) {
	if manager := ManagerFromRequest(r); manager != nil {
		manager.SuccessFromRequest(r, text)
	}
}

// Error adds an error message.
func Error(r *http.Request, text string) {
	if manager := ManagerFromRequest(r); manager != nil {
		manager.ErrorFromRequest(r, text)
	}
}

// Warning adds a warning message.
func Warning(r *http.Request, text string) {
	if manager := ManagerFromRequest(r); manager != nil {
		manager.WarningFromRequest(r, text)
	}
}

// Info adds an info message.
func Info(r *http.Request, text string) {
	if manager := ManagerFromRequest(r); manager != nil {
		manager.InfoFromRequest(r, text)
	}
}

// Get retrieves messages.
func Get(r *http.Request) []*Message {
	return MessagesFromRequest(r)
}

// Has checks if there are any messages.
func Has(r *http.Request) bool {
	return len(MessagesFromRequest(r)) > 0
}
