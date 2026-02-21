package engine

import (
	"github.com/bozz33/sublimego/color"
)

// ColorScheme defines all semantic colors used in the panel.
// Matches the template's CSS color system (primary, danger, success, warning, info, secondary).
type ColorScheme struct {
	Primary   *color.Palette
	Danger    *color.Palette
	Success   *color.Palette
	Warning   *color.Palette
	Info      *color.Palette
	Secondary *color.Palette
}

// DefaultColorScheme returns the template's default color scheme (green primary).
func DefaultColorScheme() *ColorScheme {
	c := color.Color{}
	return &ColorScheme{
		Primary:   c.Hex("#22c55e"), // Green-500
		Danger:    c.Hex("#ef4444"), // Red-500
		Success:   c.Hex("#10b981"), // Emerald-500
		Warning:   c.Hex("#f59e0b"), // Amber-500
		Info:      c.Hex("#3b82f6"), // Blue-500
		Secondary: c.Hex("#64748b"), // Slate-500
	}
}

// WithColors sets a complete custom color scheme for the panel.
// Example:
//   panel.WithColors(&engine.ColorScheme{
//       Primary:   color.Color{}.Hex("#3b82f6"),
//       Danger:    color.Color{}.Hex("#ef4444"),
//       Success:   color.Color{}.Hex("#10b981"),
//       Warning:   color.Color{}.Hex("#f59e0b"),
//       Info:      color.Color{}.Hex("#06b6d4"),
//       Secondary: color.Color{}.Hex("#6b7280"),
//   })
func (p *Panel) WithColors(scheme *ColorScheme) *Panel {
	// Store in panel for later CSS generation
	if p.colorScheme == nil {
		p.colorScheme = scheme
	}
	return p
}

// WithSemanticColors allows setting individual semantic colors.
// Example:
//   panel.WithSemanticColors(map[string]string{
//       "primary":   "#3b82f6",
//       "danger":    "#ef4444",
//       "success":   "#10b981",
//   })
func (p *Panel) WithSemanticColors(colors map[string]string) *Panel {
	c := color.Color{}
	scheme := DefaultColorScheme()

	if primary, ok := colors["primary"]; ok {
		scheme.Primary = c.Hex(primary)
	}
	if danger, ok := colors["danger"]; ok {
		scheme.Danger = c.Hex(danger)
	}
	if success, ok := colors["success"]; ok {
		scheme.Success = c.Hex(success)
	}
	if warning, ok := colors["warning"]; ok {
		scheme.Warning = c.Hex(warning)
	}
	if info, ok := colors["info"]; ok {
		scheme.Info = c.Hex(info)
	}
	if secondary, ok := colors["secondary"]; ok {
		scheme.Secondary = c.Hex(secondary)
	}

	p.colorScheme = scheme
	return p
}

// GenerateColorCSS generates CSS variables for all semantic colors.
// This is called automatically during panel initialization.
func (p *Panel) GenerateColorCSS() string {
	scheme := p.colorScheme
	if scheme == nil {
		scheme = DefaultColorScheme()
	}

	css := "@theme {\n"
	
	// Primary
	if scheme.Primary != nil {
		for _, shade := range scheme.Primary.Shades {
			css += "  --color-primary-" + shade.Hex + ";\n"
		}
	}
	
	// Danger
	if scheme.Danger != nil {
		for _, shade := range scheme.Danger.Shades {
			css += "  --color-danger-" + shade.Hex + ";\n"
		}
	}
	
	// Success
	if scheme.Success != nil {
		for _, shade := range scheme.Success.Shades {
			css += "  --color-success-" + shade.Hex + ";\n"
		}
	}
	
	// Warning
	if scheme.Warning != nil {
		for _, shade := range scheme.Warning.Shades {
			css += "  --color-warning-" + shade.Hex + ";\n"
		}
	}
	
	// Info
	if scheme.Info != nil {
		for _, shade := range scheme.Info.Shades {
			css += "  --color-info-" + shade.Hex + ";\n"
		}
	}
	
	// Secondary
	if scheme.Secondary != nil {
		for _, shade := range scheme.Secondary.Shades {
			css += "  --color-secondary-" + shade.Hex + ";\n"
		}
	}
	
	css += "}\n"
	return css
}
