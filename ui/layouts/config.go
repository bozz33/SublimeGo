package layouts

import "context"

// PanelConfig contains the admin panel configuration — Filament-style.
type PanelConfig struct {
	Name         string // Panel name (ex: "SublimeGo")
	Path         string // Base path (ex: "/admin")
	Logo         string // Logo URL (optional)
	Favicon      string // Favicon URL (optional)
	PrimaryColor string // Accent color: green, blue, red, purple, orange, pink, indigo
	DarkMode     bool   // Enable dark mode by default

	Registration      bool // Enable /register route
	EmailVerification bool // Enable email verification flow
	PasswordReset     bool // Enable /forgot-password route
	Profile           bool // Enable /profile page
	Notifications     bool // Enable notification bell + SSE

	Navigation []NavItem // Navigation items
}

// DefaultPanelConfig returns the default configuration
func DefaultPanelConfig() *PanelConfig {
	return &PanelConfig{
		Name:              "SublimeGo",
		Path:              "/admin",
		PrimaryColor:      "green",
		DarkMode:          false,
		Registration:      true,
		EmailVerification: false,
		PasswordReset:     true,
		Profile:           true,
		Navigation:        []NavItem{},
	}
}

// panelConfig stores the global/default panel configuration.
// For multi-panel setups, use WithPanelConfig(ctx, cfg) to inject per-request config.
var panelConfig = DefaultPanelConfig()

// SetPanelConfig sets the global panel configuration.
func SetPanelConfig(config *PanelConfig) {
	if config != nil {
		panelConfig = config
	}
}

type panelConfigKey struct{}

// WithPanelConfig returns a new context carrying the given PanelConfig.
// Use this in multi-panel setups to inject per-panel config into each request.
func WithPanelConfig(ctx context.Context, cfg *PanelConfig) context.Context {
	return context.WithValue(ctx, panelConfigKey{}, cfg)
}

// GetPanelConfig returns the PanelConfig for the current context.
// Falls back to the global config if none is set in the context.
func GetPanelConfigFromContext(ctx context.Context) *PanelConfig {
	if cfg, ok := ctx.Value(panelConfigKey{}).(*PanelConfig); ok && cfg != nil {
		return cfg
	}
	return panelConfig
}

// GetPanelConfig returns the global panel configuration.
// Templates call this; for context-aware access use GetPanelConfigFromContext.
func GetPanelConfig() *PanelConfig {
	return panelConfig
}

// initAppData returns the Alpine.js x-data string for the root html element,
// respecting the panel's default dark mode setting.
func initAppData(defaultDark bool) string {
	if defaultDark {
		return `{ darkMode: localStorage.getItem('theme') === 'dark' || (!localStorage.getItem('theme') && true), sidebarOpen: window.innerWidth >= 1024 }`
	}
	return `{ darkMode: localStorage.getItem('theme') === 'dark' || (!localStorage.getItem('theme') && window.matchMedia('(prefers-color-scheme: dark)').matches), sidebarOpen: window.innerWidth >= 1024 }`
}

// primaryColorPalette returns a Tailwind-compatible JS color object for the given color name.
func primaryColorPalette(color string) string {
	palettes := map[string]string{
		"green":  `{ 50:'#f0fdf4',100:'#dcfce7',200:'#bbf7d0',300:'#86efac',400:'#4ade80',500:'#22c55e',600:'#16a34a',700:'#15803d',800:'#166534',900:'#14532d' }`,
		"blue":   `{ 50:'#eff6ff',100:'#dbeafe',200:'#bfdbfe',300:'#93c5fd',400:'#60a5fa',500:'#3b82f6',600:'#2563eb',700:'#1d4ed8',800:'#1e40af',900:'#1e3a8a' }`,
		"red":    `{ 50:'#fef2f2',100:'#fee2e2',200:'#fecaca',300:'#fca5a5',400:'#f87171',500:'#ef4444',600:'#dc2626',700:'#b91c1c',800:'#991b1b',900:'#7f1d1d' }`,
		"purple": `{ 50:'#faf5ff',100:'#f3e8ff',200:'#e9d5ff',300:'#d8b4fe',400:'#c084fc',500:'#a855f7',600:'#9333ea',700:'#7e22ce',800:'#6b21a8',900:'#581c87' }`,
		"orange": `{ 50:'#fff7ed',100:'#ffedd5',200:'#fed7aa',300:'#fdba74',400:'#fb923c',500:'#f97316',600:'#ea580c',700:'#c2410c',800:'#9a3412',900:'#7c2d12' }`,
		"pink":   `{ 50:'#fdf2f8',100:'#fce7f3',200:'#fbcfe8',300:'#f9a8d4',400:'#f472b6',500:'#ec4899',600:'#db2777',700:'#be185d',800:'#9d174d',900:'#831843' }`,
		"indigo": `{ 50:'#eef2ff',100:'#e0e7ff',200:'#c7d2fe',300:'#a5b4fc',400:'#818cf8',500:'#6366f1',600:'#4f46e5',700:'#4338ca',800:'#3730a3',900:'#312e81' }`,
	}
	if p, ok := palettes[color]; ok {
		return p
	}
	return palettes["green"]
}
