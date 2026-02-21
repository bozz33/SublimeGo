package color_test

import (
	"strings"
	"testing"

	"github.com/bozz33/sublimego/color"
)

func TestPaletteCSSVars(t *testing.T) {
	p := &color.Palette{
		Name: "test",
		Shades: []color.Shade{
			{Number: 100, Hex: "#f0f0f0"},
			{Number: 500, Hex: "#808080"},
			{Number: 900, Hex: "#101010"},
		},
	}
	css := p.CSSVars("primary")
	if !strings.Contains(css, "--color-primary-100: #f0f0f0") {
		t.Errorf("missing --color-primary-100, got:\n%s", css)
	}
	if !strings.Contains(css, "--color-primary-500: #808080") {
		t.Errorf("missing --color-primary-500, got:\n%s", css)
	}
	if !strings.Contains(css, "--color-primary-900: #101010") {
		t.Errorf("missing --color-primary-900, got:\n%s", css)
	}
}

func TestManagerRegisterAndGet(t *testing.T) {
	m := color.NewManager()
	p := &color.Palette{Name: "brand", Shades: []color.Shade{{Number: 500, Hex: "#ff0000"}}}
	m.Register("brand", p)

	got := m.Get("brand")
	if got == nil {
		t.Fatal("expected to find palette 'brand'")
	}
	if got.Name != "brand" {
		t.Errorf("expected name 'brand', got %q", got.Name)
	}
}

func TestManagerSetPrimary(t *testing.T) {
	m := color.NewManager()
	// "green" is already built-in, just set it as primary
	if err := m.SetPrimary("green"); err != nil {
		t.Fatalf("SetPrimary failed: %v", err)
	}

	css := m.PrimaryCSSVars()
	if !strings.Contains(css, "--color-primary-500") {
		t.Errorf("PrimaryCSSVars missing --color-primary-500, got:\n%s", css)
	}
}

func TestManagerAllCSSVars(t *testing.T) {
	// NewManager already includes built-in palettes (red, blue, green, etc.)
	m := color.NewManager()
	css := m.AllCSSVars()
	if !strings.Contains(css, "--color-red-500") {
		t.Errorf("AllCSSVars missing --color-red-500, got:\n%s", css)
	}
	if !strings.Contains(css, "--color-blue-500") {
		t.Errorf("AllCSSVars missing --color-blue-500, got:\n%s", css)
	}
}

func TestBuiltinPalettes(t *testing.T) {
	palettes := []*color.Palette{
		color.Green, color.Blue, color.Red, color.Purple,
		color.Orange, color.Indigo, color.Teal, color.Rose,
		color.Amber, color.Cyan,
	}
	for _, p := range palettes {
		if p.Name == "" {
			t.Error("palette has empty name")
		}
		if len(p.Shades) == 0 {
			t.Errorf("palette %q has no shades", p.Name)
		}
	}
}

func TestManagerGetNotFound(t *testing.T) {
	m := color.NewManager()
	got := m.Get("nonexistent")
	if got != nil {
		t.Error("expected nil for nonexistent palette")
	}
}

func TestColorHex(t *testing.T) {
	c := color.Color{}
	palette := c.Hex("#3b82f6")
	if palette == nil {
		t.Fatal("expected non-nil palette")
	}
	if len(palette.Shades) != 11 {
		t.Errorf("expected 11 shades, got %d", len(palette.Shades))
	}
	// Verify 500 shade exists
	found := false
	for _, shade := range palette.Shades {
		if shade.Number == 500 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected shade 500 in palette")
	}
}

func TestColorRGB(t *testing.T) {
	c := color.Color{}
	palette := c.RGB("rgb(59, 130, 246)")
	if palette == nil {
		t.Fatal("expected non-nil palette")
	}
	if len(palette.Shades) != 11 {
		t.Errorf("expected 11 shades, got %d", len(palette.Shades))
	}
}

func TestColorFromRGB(t *testing.T) {
	c := color.Color{}
	palette := c.FromRGB(59, 130, 246)
	if palette == nil {
		t.Fatal("expected non-nil palette")
	}
	if len(palette.Shades) != 11 {
		t.Errorf("expected 11 shades, got %d", len(palette.Shades))
	}
}

func TestColorInvalidHex(t *testing.T) {
	c := color.Color{}
	palette := c.Hex("invalid")
	if palette == nil {
		t.Fatal("expected fallback palette")
	}
	// Should return gray fallback
	if len(palette.Shades) != 11 {
		t.Errorf("expected 11 shades in fallback, got %d", len(palette.Shades))
	}
}
