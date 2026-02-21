package notifications_test

import (
	"testing"

	"github.com/bozz33/sublimego/notifications"
)

func TestInfoBuilder(t *testing.T) {
	n := notifications.Info("Test title").WithBody("Test body")
	if n.Title != "Test title" {
		t.Errorf("expected title 'Test title', got %q", n.Title)
	}
	if n.Body != "Test body" {
		t.Errorf("expected body 'Test body', got %q", n.Body)
	}
	if n.Level != notifications.LevelInfo {
		t.Errorf("expected level Info, got %v", n.Level)
	}
}

func TestSuccessBuilder(t *testing.T) {
	n := notifications.Success("Done")
	if n.Level != notifications.LevelSuccess {
		t.Errorf("expected level Success, got %v", n.Level)
	}
	if n.Icon == "" {
		t.Error("expected non-empty icon for Success")
	}
}

func TestWarningBuilder(t *testing.T) {
	n := notifications.Warning("Warn")
	if n.Level != notifications.LevelWarning {
		t.Errorf("expected level Warning, got %v", n.Level)
	}
}

func TestDangerBuilder(t *testing.T) {
	n := notifications.Danger("Error")
	if n.Level != notifications.LevelDanger {
		t.Errorf("expected level Danger, got %v", n.Level)
	}
}

func TestWithDuration(t *testing.T) {
	n := notifications.Info("title").WithDuration(10)
	if n.Duration != 10 {
		t.Errorf("expected duration 10, got %d", n.Duration)
	}
}

func TestPersistent(t *testing.T) {
	n := notifications.Info("title").Persistent()
	if n.Duration != 0 {
		t.Errorf("expected duration 0 for persistent, got %d", n.Duration)
	}
}

func TestStoreAddAndGet(t *testing.T) {
	store := notifications.NewStore(50)
	n := notifications.Info("hello").WithBody("world")
	store.Send("user1", n)

	items := store.GetAll("user1")
	if len(items) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(items))
	}
	if items[0].Title != "hello" {
		t.Errorf("expected title 'hello', got %q", items[0].Title)
	}
}

func TestStoreMarkRead(t *testing.T) {
	store := notifications.NewStore(50)
	n := notifications.Info("hello").WithBody("world")
	store.Send("user1", n)

	items := store.GetAll("user1")
	if len(items) == 0 {
		t.Fatal("expected at least 1 notification")
	}
	id := items[0].ID
	store.MarkRead("user1", id)

	updated := store.GetAll("user1")
	for _, item := range updated {
		if item.ID == id && !item.Read {
			t.Error("expected notification to be marked as read")
		}
	}
}

func TestStoreGetUnread(t *testing.T) {
	store := notifications.NewStore(50)
	store.Send("user1", notifications.Info("a"))
	store.Send("user1", notifications.Info("hello").WithBody("world"))

	items := store.GetAll("user1")
	if len(items) != 2 {
		t.Fatal("expected notifications")
	}
	// mark first as read
	store.MarkRead("user1", items[0].ID)

	unread := store.GetUnread("user1")
	for _, item := range unread {
		if item.Read {
			t.Error("GetUnread returned a read notification")
		}
	}
}
