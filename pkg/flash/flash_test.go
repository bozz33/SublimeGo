package flash

import (
	"context"
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMessage(t *testing.T) {
	msg := NewMessage(TypeSuccess, "Test message")

	assert.Equal(t, TypeSuccess, msg.Type)
	assert.Equal(t, "Test message", msg.Text)
	assert.Empty(t, msg.Title)
}

func TestMessageWithTitle(t *testing.T) {
	msg := NewMessage(TypeSuccess, "Test").WithTitle("Title")

	assert.Equal(t, "Title", msg.Title)
	assert.Equal(t, "Test", msg.Text)
}

func TestNewManager(t *testing.T) {
	session := scs.New()
	manager := NewManager(session)

	require.NotNil(t, manager)
	assert.Equal(t, session, manager.session)
}

func TestManagerSuccess(t *testing.T) {
	session := scs.New()
	manager := NewManager(session)

	// Utiliser LoadAndSave pour initialiser le contexte
	ctx, _ := session.Load(context.Background(), "")

	manager.Success(ctx, "Success message")

	messages := manager.Get(ctx)
	require.Len(t, messages, 1)
	assert.Equal(t, TypeSuccess, messages[0].Type)
	assert.Equal(t, "Success message", messages[0].Text)
}

func TestManagerError(t *testing.T) {
	session := scs.New()
	manager := NewManager(session)
	ctx, _ := session.Load(context.Background(), "")

	manager.Error(ctx, "Error message")

	messages := manager.Get(ctx)
	require.Len(t, messages, 1)
	assert.Equal(t, TypeError, messages[0].Type)
}

func TestManagerWarning(t *testing.T) {
	session := scs.New()
	manager := NewManager(session)
	ctx, _ := session.Load(context.Background(), "")

	manager.Warning(ctx, "Warning message")

	messages := manager.Get(ctx)
	require.Len(t, messages, 1)
	assert.Equal(t, TypeWarning, messages[0].Type)
}

func TestManagerInfo(t *testing.T) {
	session := scs.New()
	manager := NewManager(session)
	ctx, _ := session.Load(context.Background(), "")

	manager.Info(ctx, "Info message")

	messages := manager.Get(ctx)
	require.Len(t, messages, 1)
	assert.Equal(t, TypeInfo, messages[0].Type)
}

func TestManagerWithTitle(t *testing.T) {
	session := scs.New()
	manager := NewManager(session)
	ctx, _ := session.Load(context.Background(), "")

	manager.SuccessWithTitle(ctx, "Title", "Text")

	messages := manager.Get(ctx)
	require.Len(t, messages, 1)
	assert.Equal(t, "Title", messages[0].Title)
	assert.Equal(t, "Text", messages[0].Text)
}

func TestManagerMultipleMessages(t *testing.T) {
	session := scs.New()
	manager := NewManager(session)
	ctx, _ := session.Load(context.Background(), "")

	manager.Success(ctx, "Success")
	manager.Error(ctx, "Error")
	manager.Warning(ctx, "Warning")

	messages := manager.Get(ctx)
	assert.Len(t, messages, 3)
}

func TestManagerGetAndClear(t *testing.T) {
	session := scs.New()
	manager := NewManager(session)
	ctx, _ := session.Load(context.Background(), "")

	manager.Success(ctx, "Test")

	// First Get
	messages := manager.GetAndClear(ctx)
	assert.Len(t, messages, 1)

	// Second Get (should be empty)
	messages = manager.Get(ctx)
	assert.Len(t, messages, 0)
}

func TestManagerClear(t *testing.T) {
	session := scs.New()
	manager := NewManager(session)
	ctx, _ := session.Load(context.Background(), "")

	manager.Success(ctx, "Test 1")
	manager.Error(ctx, "Test 2")

	assert.Len(t, manager.Get(ctx), 2)

	manager.Clear(ctx)

	assert.Len(t, manager.Get(ctx), 0)
}

func TestManagerHas(t *testing.T) {
	session := scs.New()
	manager := NewManager(session)
	ctx, _ := session.Load(context.Background(), "")

	assert.False(t, manager.Has(ctx))

	manager.Success(ctx, "Test")

	assert.True(t, manager.Has(ctx))
}

func TestManagerHasType(t *testing.T) {
	session := scs.New()
	manager := NewManager(session)
	ctx, _ := session.Load(context.Background(), "")

	manager.Success(ctx, "Test")

	assert.True(t, manager.HasType(ctx, TypeSuccess))
	assert.False(t, manager.HasType(ctx, TypeError))
}

func TestManagerGetByType(t *testing.T) {
	session := scs.New()
	manager := NewManager(session)
	ctx, _ := session.Load(context.Background(), "")

	manager.Success(ctx, "Success 1")
	manager.Success(ctx, "Success 2")
	manager.Error(ctx, "Error 1")

	successMessages := manager.GetByType(ctx, TypeSuccess)
	assert.Len(t, successMessages, 2)

	errorMessages := manager.GetByType(ctx, TypeError)
	assert.Len(t, errorMessages, 1)
}

func TestManagerCount(t *testing.T) {
	session := scs.New()
	manager := NewManager(session)
	ctx, _ := session.Load(context.Background(), "")

	assert.Equal(t, 0, manager.Count(ctx))

	manager.Success(ctx, "Test 1")
	manager.Error(ctx, "Test 2")

	assert.Equal(t, 2, manager.Count(ctx))
}

func TestManagerCountByType(t *testing.T) {
	session := scs.New()
	manager := NewManager(session)
	ctx, _ := session.Load(context.Background(), "")

	manager.Success(ctx, "Success 1")
	manager.Success(ctx, "Success 2")
	manager.Error(ctx, "Error 1")

	assert.Equal(t, 2, manager.CountByType(ctx, TypeSuccess))
	assert.Equal(t, 1, manager.CountByType(ctx, TypeError))
	assert.Equal(t, 0, manager.CountByType(ctx, TypeWarning))
}

func TestManagerAdd(t *testing.T) {
	session := scs.New()
	manager := NewManager(session)
	ctx, _ := session.Load(context.Background(), "")

	customMsg := &Message{
		Type:  "custom",
		Text:  "Custom message",
		Title: "Custom",
	}

	manager.Add(ctx, customMsg)

	messages := manager.Get(ctx)
	require.Len(t, messages, 1)
	assert.Equal(t, "custom", messages[0].Type)
	assert.Equal(t, "Custom message", messages[0].Text)
	assert.Equal(t, "Custom", messages[0].Title)
}

func TestWithManager(t *testing.T) {
	session := scs.New()
	manager := NewManager(session)

	ctx := WithManager(context.Background(), manager)

	retrieved := ManagerFromContext(ctx)
	assert.Equal(t, manager, retrieved)
}

func TestWithMessages(t *testing.T) {
	messages := []*Message{
		NewMessage(TypeSuccess, "Test"),
	}

	ctx := WithMessages(context.Background(), messages)

	retrieved := MessagesFromContext(ctx)
	assert.Equal(t, messages, retrieved)
}

func TestMessagesFromContextEmpty(t *testing.T) {
	ctx := context.Background()

	messages := MessagesFromContext(ctx)
	assert.NotNil(t, messages)
	assert.Len(t, messages, 0)
}

func BenchmarkManagerAdd(b *testing.B) {
	session := scs.New()
	manager := NewManager(session)
	ctx, _ := session.Load(context.Background(), "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.Success(ctx, "Test message")
	}
}

func BenchmarkManagerGet(b *testing.B) {
	session := scs.New()
	manager := NewManager(session)
	ctx, _ := session.Load(context.Background(), "")

	// Ajouter quelques messages
	for i := 0; i < 10; i++ {
		manager.Success(ctx, "Test")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.Get(ctx)
	}
}

func BenchmarkManagerGetByType(b *testing.B) {
	session := scs.New()
	manager := NewManager(session)
	ctx, _ := session.Load(context.Background(), "")

	manager.Success(ctx, "Success 1")
	manager.Success(ctx, "Success 2")
	manager.Error(ctx, "Error 1")
	manager.Warning(ctx, "Warning 1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetByType(ctx, TypeSuccess)
	}
}
