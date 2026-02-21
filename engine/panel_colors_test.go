package engine_test

import (
	"testing"

	"github.com/bozz33/sublimego/color"
	"github.com/bozz33/sublimego/engine"
)

func TestDefaultColorScheme(t *testing.T) {
	scheme := engine.DefaultColorScheme()
	if scheme == nil {
		t.Fatal("expected non-nil color scheme")
	}
	if scheme.Primary == nil {
		t.Error("expected primary color")
	}
	if scheme.Danger == nil {
		t.Error("expected danger color")
	}
	if scheme.Success == nil {
		t.Error("expected success color")
	}
	if scheme.Warning == nil {
		t.Error("expected warning color")
	}
	if scheme.Info == nil {
		t.Error("expected info color")
	}
	if scheme.Secondary == nil {
		t.Error("expected secondary color")
	}
}

func TestPanelWithColors(t *testing.T) {
	c := color.Color{}
	scheme := &engine.ColorScheme{
		Primary:   c.Hex("#3b82f6"),
		Danger:    c.Hex("#ef4444"),
		Success:   c.Hex("#10b981"),
		Warning:   c.Hex("#f59e0b"),
		Info:      c.Hex("#06b6d4"),
		Secondary: c.Hex("#64748b"),
	}

	panel := engine.NewPanel("test").WithColors(scheme)
	if panel == nil {
		t.Fatal("expected non-nil panel")
	}
}

func TestPanelWithSemanticColors(t *testing.T) {
	panel := engine.NewPanel("test").WithSemanticColors(map[string]string{
		"primary":   "#3b82f6",
		"danger":    "#ef4444",
		"success":   "#10b981",
		"warning":   "#f59e0b",
		"info":      "#06b6d4",
		"secondary": "#64748b",
	})

	if panel == nil {
		t.Fatal("expected non-nil panel")
	}
}

func TestGenerateColorCSS(t *testing.T) {
	panel := engine.NewPanel("test").WithSemanticColors(map[string]string{
		"primary": "#3b82f6",
	})

	css := panel.GenerateColorCSS()
	if css == "" {
		t.Error("expected non-empty CSS")
	}
	if len(css) < 50 {
		t.Errorf("expected substantial CSS output, got %d bytes", len(css))
	}
}
