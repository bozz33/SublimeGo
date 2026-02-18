package form

import (
	"testing"
)

func TestNewForm(t *testing.T) {
	f := New()

	if f == nil {
		t.Error("Expected form to be created")
	}
	if len(f.Schema) != 0 {
		t.Errorf("Expected empty schema, got %d components", len(f.Schema))
	}
	if f.State == nil {
		t.Error("Expected State map to be initialized")
	}
}

func TestFormSetSchema(t *testing.T) {
	f := New().SetSchema(
		Text("name"),
		Email("email"),
	)

	if len(f.Schema) != 2 {
		t.Errorf("Expected 2 components, got %d", len(f.Schema))
	}
}

func TestFormBind(t *testing.T) {
	type User struct {
		Name  string
		Email string
	}

	user := &User{Name: "John", Email: "john@example.com"}
	f := New().Bind(user)

	if f.Model != user {
		t.Error("Expected model to be bound")
	}
}

func TestTextInput(t *testing.T) {
	field := Text("username")

	if field.Name() != "username" {
		t.Errorf("Expected name 'username', got '%s'", field.Name())
	}
	if field.LabelStr != "username" {
		t.Errorf("Expected default label 'username', got '%s'", field.LabelStr)
	}
	if field.Type != "text" {
		t.Errorf("Expected type 'text', got '%s'", field.Type)
	}
}

func TestTextInputFluent(t *testing.T) {
	field := Text("name").
		Label("Full Name").
		WithPlaceholder("Enter your name").
		HelperText("Your legal name").
		Required()

	if field.LabelStr != "Full Name" {
		t.Errorf("Expected label 'Full Name', got '%s'", field.LabelStr)
	}
	if field.Placeholder() != "Enter your name" {
		t.Errorf("Expected placeholder, got '%s'", field.Placeholder())
	}
	if field.HelpText != "Your legal name" {
		t.Errorf("Expected help text, got '%s'", field.HelpText)
	}
	if !field.BaseField.Required {
		t.Error("Expected Required to be true")
	}
}

func TestEmailInput(t *testing.T) {
	field := Email("email")

	if field.Type != "email" {
		t.Errorf("Expected type 'email', got '%s'", field.Type)
	}
	if len(field.Rules()) == 0 || field.Rules()[0] != "email" {
		t.Error("Expected 'email' rule")
	}
}

func TestPasswordInput(t *testing.T) {
	field := Password("password")

	if field.Type != "password" {
		t.Errorf("Expected type 'password', got '%s'", field.Type)
	}
}

func TestSelect(t *testing.T) {
	field := NewSelect("role").
		Label("User Role").
		Options(map[string]string{
			"admin": "Administrator",
			"user":  "Regular User",
		}).
		Required()

	if field.Name() != "role" {
		t.Errorf("Expected name 'role', got '%s'", field.Name())
	}
	if field.LabelStr != "User Role" {
		t.Errorf("Expected label 'User Role', got '%s'", field.LabelStr)
	}
	if len(field.SelectOptions()) != 2 {
		t.Errorf("Expected 2 options, got %d", len(field.SelectOptions()))
	}
	if !field.BaseField.Required {
		t.Error("Expected Required to be true")
	}
}

func TestCheckbox(t *testing.T) {
	field := NewCheckbox("active").
		Label("Is Active").
		Default(true)

	if field.Name() != "active" {
		t.Errorf("Expected name 'active', got '%s'", field.Name())
	}
	if field.Value() != true {
		t.Error("Expected default value true")
	}
}

func TestTextarea(t *testing.T) {
	field := NewTextarea("description").
		Label("Description").
		Rows(5).
		Required()

	if field.Name() != "description" {
		t.Errorf("Expected name 'description', got '%s'", field.Name())
	}
	if field.RowCount != 5 {
		t.Errorf("Expected 5 rows, got %d", field.RowCount)
	}
	if !field.BaseField.Required {
		t.Error("Expected Required to be true")
	}
}

func TestFileUpload(t *testing.T) {
	field := NewFileUpload("avatar").
		Label("Profile Picture").
		Accept("image/*").
		MaxSize(5 * 1024 * 1024).
		Required()

	if field.Name() != "avatar" {
		t.Errorf("Expected name 'avatar', got '%s'", field.Name())
	}
	if field.AcceptTypes != "image/*" {
		t.Errorf("Expected accept 'image/*', got '%s'", field.AcceptTypes)
	}
	if field.MaxFileSize != 5*1024*1024 {
		t.Errorf("Expected max size 5MB, got %d", field.MaxFileSize)
	}
}

func TestFieldVisibility(t *testing.T) {
	field := Text("hidden_field")
	field.Hidden = true

	if field.IsVisible() {
		t.Error("Expected field to be hidden")
	}

	field.Hidden = false
	if !field.IsVisible() {
		t.Error("Expected field to be visible")
	}
}

func TestFieldComponentType(t *testing.T) {
	field := Text("test")

	if field.ComponentType() != "field" {
		t.Errorf("Expected component type 'field', got '%s'", field.ComponentType())
	}
}
