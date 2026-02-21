// Package color provides dynamic color palette management for SublimeGo panels.
// Colors are defined as Tailwind CSS shade maps and can be injected as CSS variables.
package color

import "fmt"

// Shade maps a Tailwind shade number to a hex color value.
type Shade struct {
	Number int
	Hex    string
}

// Palette is a named set of color shades (50–950).
type Palette struct {
	Name   string
	Shades []Shade
}

// CSSVars returns the palette as CSS custom property declarations.
// Example: --color-primary-500: #22c55e;
func (p *Palette) CSSVars(prefix string) string {
	out := ""
	for _, s := range p.Shades {
		out += fmt.Sprintf("  --%s-%s-%d: %s;\n", "color", prefix, s.Number, s.Hex)
	}
	return out
}

// ---------------------------------------------------------------------------
// Built-in palettes (Tailwind v3 defaults)
// ---------------------------------------------------------------------------

// Green is the default primary palette (Tailwind green).
var Green = &Palette{
	Name: "green",
	Shades: []Shade{
		{50, "#f0fdf4"}, {100, "#dcfce7"}, {200, "#bbf7d0"},
		{300, "#86efac"}, {400, "#4ade80"}, {500, "#22c55e"},
		{600, "#16a34a"}, {700, "#15803d"}, {800, "#166534"},
		{900, "#14532d"}, {950, "#052e16"},
	},
}

// Blue palette.
var Blue = &Palette{
	Name: "blue",
	Shades: []Shade{
		{50, "#eff6ff"}, {100, "#dbeafe"}, {200, "#bfdbfe"},
		{300, "#93c5fd"}, {400, "#60a5fa"}, {500, "#3b82f6"},
		{600, "#2563eb"}, {700, "#1d4ed8"}, {800, "#1e40af"},
		{900, "#1e3a8a"}, {950, "#172554"},
	},
}

// Purple palette.
var Purple = &Palette{
	Name: "purple",
	Shades: []Shade{
		{50, "#faf5ff"}, {100, "#f3e8ff"}, {200, "#e9d5ff"},
		{300, "#d8b4fe"}, {400, "#c084fc"}, {500, "#a855f7"},
		{600, "#9333ea"}, {700, "#7e22ce"}, {800, "#6b21a8"},
		{900, "#581c87"}, {950, "#3b0764"},
	},
}

// Red palette.
var Red = &Palette{
	Name: "red",
	Shades: []Shade{
		{50, "#fef2f2"}, {100, "#fee2e2"}, {200, "#fecaca"},
		{300, "#fca5a5"}, {400, "#f87171"}, {500, "#ef4444"},
		{600, "#dc2626"}, {700, "#b91c1c"}, {800, "#991b1b"},
		{900, "#7f1d1d"}, {950, "#450a0a"},
	},
}

// Orange palette.
var Orange = &Palette{
	Name: "orange",
	Shades: []Shade{
		{50, "#fff7ed"}, {100, "#ffedd5"}, {200, "#fed7aa"},
		{300, "#fdba74"}, {400, "#fb923c"}, {500, "#f97316"},
		{600, "#ea580c"}, {700, "#c2410c"}, {800, "#9a3412"},
		{900, "#7c2d12"}, {950, "#431407"},
	},
}

// Indigo palette.
var Indigo = &Palette{
	Name: "indigo",
	Shades: []Shade{
		{50, "#eef2ff"}, {100, "#e0e7ff"}, {200, "#c7d2fe"},
		{300, "#a5b4fc"}, {400, "#818cf8"}, {500, "#6366f1"},
		{600, "#4f46e5"}, {700, "#4338ca"}, {800, "#3730a3"},
		{900, "#312e81"}, {950, "#1e1b4b"},
	},
}

// Teal palette.
var Teal = &Palette{
	Name: "teal",
	Shades: []Shade{
		{50, "#f0fdfa"}, {100, "#ccfbf1"}, {200, "#99f6e4"},
		{300, "#5eead4"}, {400, "#2dd4bf"}, {500, "#14b8a6"},
		{600, "#0d9488"}, {700, "#0f766e"}, {800, "#115e59"},
		{900, "#134e4a"}, {950, "#042f2e"},
	},
}

// Rose palette.
var Rose = &Palette{
	Name: "rose",
	Shades: []Shade{
		{50, "#fff1f2"}, {100, "#ffe4e6"}, {200, "#fecdd3"},
		{300, "#fda4af"}, {400, "#fb7185"}, {500, "#f43f5e"},
		{600, "#e11d48"}, {700, "#be123c"}, {800, "#9f1239"},
		{900, "#881337"}, {950, "#4c0519"},
	},
}

// Amber palette.
var Amber = &Palette{
	Name: "amber",
	Shades: []Shade{
		{50, "#fffbeb"}, {100, "#fef3c7"}, {200, "#fde68a"},
		{300, "#fcd34d"}, {400, "#fbbf24"}, {500, "#f59e0b"},
		{600, "#d97706"}, {700, "#b45309"}, {800, "#92400e"},
		{900, "#78350f"}, {950, "#451a03"},
	},
}

// Cyan palette.
var Cyan = &Palette{
	Name: "cyan",
	Shades: []Shade{
		{50, "#ecfeff"}, {100, "#cffafe"}, {200, "#a5f3fc"},
		{300, "#67e8f9"}, {400, "#22d3ee"}, {500, "#06b6d4"},
		{600, "#0891b2"}, {700, "#0e7490"}, {800, "#155e75"},
		{900, "#164e63"}, {950, "#083344"},
	},
}

// BuiltIn is the registry of all built-in palettes by name.
var BuiltIn = map[string]*Palette{
	"green":  Green,
	"blue":   Blue,
	"purple": Purple,
	"red":    Red,
	"orange": Orange,
	"indigo": Indigo,
	"teal":   Teal,
	"rose":   Rose,
	"amber":  Amber,
	"cyan":   Cyan,
}
