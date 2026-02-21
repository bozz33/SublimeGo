package color

import (
	"fmt"
	"strings"
	"sync"
)

// Manager manages color palettes for a SublimeGo panel.
// It supports registering custom palettes and generating CSS variable blocks.
type Manager struct {
	mu       sync.RWMutex
	palettes map[string]*Palette
	primary  string
}

// NewManager creates a Manager pre-loaded with all built-in palettes.
func NewManager() *Manager {
	m := &Manager{
		palettes: make(map[string]*Palette, len(BuiltIn)+4),
		primary:  "green",
	}
	for name, p := range BuiltIn {
		m.palettes[name] = p
	}
	return m
}

// Register adds or replaces a named palette.
func (m *Manager) Register(name string, p *Palette) *Manager {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.palettes[name] = p
	return m
}

// SetPrimary sets the active primary palette by name.
// Returns an error if the palette is not registered.
func (m *Manager) SetPrimary(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.palettes[name]; !ok {
		return fmt.Errorf("color: palette %q not registered", name)
	}
	m.primary = name
	return nil
}

// Get returns a palette by name, or nil if not found.
func (m *Manager) Get(name string) *Palette {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.palettes[name]
}

// Primary returns the active primary palette.
func (m *Manager) Primary() *Palette {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.palettes[m.primary]
}

// PrimaryName returns the name of the active primary palette.
func (m *Manager) PrimaryName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.primary
}

// PrimaryCSSVars returns a <style> block injecting the primary palette
// as --color-primary-* CSS custom properties.
func (m *Manager) PrimaryCSSVars() string {
	p := m.Primary()
	if p == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(":root {\n")
	for _, s := range p.Shades {
		sb.WriteString(fmt.Sprintf("  --color-primary-%d: %s;\n", s.Number, s.Hex))
	}
	sb.WriteString("}")
	return sb.String()
}

// AllCSSVars returns CSS custom properties for all registered palettes.
// Useful for injecting a full color system into the page.
func (m *Manager) AllCSSVars() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var sb strings.Builder
	sb.WriteString(":root {\n")
	for name, p := range m.palettes {
		for _, s := range p.Shades {
			sb.WriteString(fmt.Sprintf("  --color-%s-%d: %s;\n", name, s.Number, s.Hex))
		}
	}
	sb.WriteString("}")
	return sb.String()
}

// Hex returns the hex value for a given palette name and shade number.
// Returns empty string if not found.
func (m *Manager) Hex(paletteName string, shade int) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.palettes[paletteName]
	if !ok {
		return ""
	}
	for _, s := range p.Shades {
		if s.Number == shade {
			return s.Hex
		}
	}
	return ""
}

// Default is the global color manager instance.
var Default = NewManager()
