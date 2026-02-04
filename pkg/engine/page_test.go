package engine

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-h/templ"
)

func TestNewBasePage(t *testing.T) {
	page := NewBasePage("settings", "Settings")

	if page.Slug() != "settings" {
		t.Errorf("Expected slug 'settings', got '%s'", page.Slug())
	}
	if page.Label() != "Settings" {
		t.Errorf("Expected label 'Settings', got '%s'", page.Label())
	}
	if page.Icon() != "document" {
		t.Errorf("Expected default icon 'document', got '%s'", page.Icon())
	}
	if page.Group() != "" {
		t.Errorf("Expected empty group, got '%s'", page.Group())
	}
	if page.Sort() != 100 {
		t.Errorf("Expected sort 100, got %d", page.Sort())
	}
}

func TestBasePageFluentSetters(t *testing.T) {
	page := NewBasePage("reports", "Reports").
		SetIcon("chart").
		SetGroup("Analytics").
		SetSort(50)

	if page.Icon() != "chart" {
		t.Errorf("Expected icon 'chart', got '%s'", page.Icon())
	}
	if page.Group() != "Analytics" {
		t.Errorf("Expected group 'Analytics', got '%s'", page.Group())
	}
	if page.Sort() != 50 {
		t.Errorf("Expected sort 50, got %d", page.Sort())
	}
}

func TestBasePageCanAccess(t *testing.T) {
	page := NewBasePage("test", "Test")
	ctx := context.Background()

	if !page.CanAccess(ctx) {
		t.Error("Expected CanAccess to return true by default")
	}
}

func TestSimplePage(t *testing.T) {
	renderCalled := false
	renderFunc := func(ctx context.Context, r *http.Request) templ.Component {
		renderCalled = true
		return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			w.Write([]byte("<div>Test Content</div>"))
			return nil
		})
	}

	page := NewSimplePage("analytics", "Analytics", renderFunc).
		WithIcon("chart").
		WithGroup("Reports").
		WithSort(25)

	if page.Slug() != "analytics" {
		t.Errorf("Expected slug 'analytics', got '%s'", page.Slug())
	}
	if page.Label() != "Analytics" {
		t.Errorf("Expected label 'Analytics', got '%s'", page.Label())
	}
	if page.Icon() != "chart" {
		t.Errorf("Expected icon 'chart', got '%s'", page.Icon())
	}
	if page.Group() != "Reports" {
		t.Errorf("Expected group 'Reports', got '%s'", page.Group())
	}
	if page.Sort() != 25 {
		t.Errorf("Expected sort 25, got %d", page.Sort())
	}

	// Test render
	req := httptest.NewRequest("GET", "/analytics", nil)
	component := page.Render(context.Background(), req)
	if component == nil {
		t.Error("Expected component to be returned")
	}
	if !renderCalled {
		t.Error("Expected render function to be called")
	}
}

func TestSimplePageWithAccess(t *testing.T) {
	page := NewSimplePage("admin", "Admin", nil).
		WithAccess(func(ctx context.Context) bool {
			return false
		})

	ctx := context.Background()
	if page.CanAccess(ctx) {
		t.Error("Expected CanAccess to return false")
	}
}

func TestSimplePageDefaultAccess(t *testing.T) {
	page := NewSimplePage("public", "Public", nil)

	ctx := context.Background()
	if !page.CanAccess(ctx) {
		t.Error("Expected CanAccess to return true by default")
	}
}
