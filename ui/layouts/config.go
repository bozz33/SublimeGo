package layouts

// PanelConfig contains the admin panel configuration
// Filament style - centralized configuration
type PanelConfig struct {
	Name              string    // Panel name (ex: "SublimeGo")
	Path              string    // Base path (ex: "/admin")
	Logo              string    // Logo URL (optional)
	Favicon           string    // Favicon URL (optional)
	PrimaryColor      string    // Primary color (default: green)
	DarkMode          bool      // Enable dark mode by default
	Registration      bool      // Enable registration
	EmailVerification bool      // Enable email verification
	PasswordReset     bool      // Enable password reset
	Profile           bool      // Enable profile page
	Navigation        []NavItem // Navigation items
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

// panelConfig stores the global panel configuration
var panelConfig = DefaultPanelConfig()

// SetPanelConfig sets the panel configuration
func SetPanelConfig(config *PanelConfig) {
	if config != nil {
		panelConfig = config
	}
}

// GetPanelConfig returns the current configuration
func GetPanelConfig() *PanelConfig {
	return panelConfig
}
