package actions

import (
	"testing"
)

// MockEntity pour les tests
type MockEntity struct {
	ID   int
	Name string
}

func (m *MockEntity) GetID() int {
	return m.ID
}

func TestNew(t *testing.T) {
	action := New("test")

	if action.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", action.Name)
	}
	if action.Label != "test" {
		t.Errorf("Expected label 'test', got '%s'", action.Label)
	}
	if action.Type != Link {
		t.Errorf("Expected type Link, got '%s'", action.Type)
	}
	if action.Method != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", action.Method)
	}
	if action.Color != "gray" {
		t.Errorf("Expected color 'gray', got '%s'", action.Color)
	}
}

func TestFluentAPI(t *testing.T) {
	action := New("test").
		SetLabel("Test Label").
		SetIcon("star").
		SetColor("primary")

	if action.Label != "Test Label" {
		t.Errorf("Expected label 'Test Label', got '%s'", action.Label)
	}
	if action.Icon != "star" {
		t.Errorf("Expected icon 'star', got '%s'", action.Icon)
	}
	if action.Color != "primary" {
		t.Errorf("Expected color 'primary', got '%s'", action.Color)
	}
}

func TestRequiresDialog(t *testing.T) {
	action := New("delete").RequiresDialog("Delete?", "This action is irreversible")

	if !action.RequiresConfirmation {
		t.Error("Expected RequiresConfirmation to be true")
	}
	if action.ModalTitle != "Delete?" {
		t.Errorf("Expected ModalTitle 'Delete?', got '%s'", action.ModalTitle)
	}
	if action.ModalDescription != "This action is irreversible" {
		t.Errorf("Expected ModalDescription, got '%s'", action.ModalDescription)
	}
	if action.Type != Button {
		t.Errorf("Expected type Button, got '%s'", action.Type)
	}
}

func TestEditAction(t *testing.T) {
	action := EditAction("/users")

	if action.Name != "edit" {
		t.Errorf("Expected name 'edit', got '%s'", action.Name)
	}
	if action.Label != "Edit" {
		t.Errorf("Expected label 'Edit', got '%s'", action.Label)
	}
	if action.Color != "primary" {
		t.Errorf("Expected color 'primary', got '%s'", action.Color)
	}
	if action.Icon != "pencil" {
		t.Errorf("Expected icon 'pencil', got '%s'", action.Icon)
	}
}

func TestDeleteAction(t *testing.T) {
	action := DeleteAction("/users")

	if action.Name != "delete" {
		t.Errorf("Expected name 'delete', got '%s'", action.Name)
	}
	if action.Label != "Delete" {
		t.Errorf("Expected label 'Delete', got '%s'", action.Label)
	}
	if action.Color != "danger" {
		t.Errorf("Expected color 'danger', got '%s'", action.Color)
	}
	if !action.RequiresConfirmation {
		t.Error("Expected RequiresConfirmation to be true")
	}
	if action.Method != "DELETE" {
		t.Errorf("Expected method 'DELETE', got '%s'", action.Method)
	}
}

func TestGetItemID_WithIdentifiable(t *testing.T) {
	entity := &MockEntity{ID: 42, Name: "Test"}

	id := GetItemID(entity)

	if id != "42" {
		t.Errorf("Expected ID '42', got '%s'", id)
	}
}

func TestGetItemID_WithoutIdentifiable(t *testing.T) {
	item := "simple-string"

	id := GetItemID(item)

	if id != "simple-string" {
		t.Errorf("Expected 'simple-string', got '%s'", id)
	}
}

func TestSetUrl(t *testing.T) {
	entity := &MockEntity{ID: 123, Name: "Test"}

	action := New("test").SetUrl(func(item any) string {
		if e, ok := item.(*MockEntity); ok {
			return "/users/" + GetItemID(e) + "/edit"
		}
		return ""
	})

	url := action.UrlResolver(entity)

	if url != "/users/123/edit" {
		t.Errorf("Expected '/users/123/edit', got '%s'", url)
	}
}
