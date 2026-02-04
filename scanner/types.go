package scanner

import "time"

// ImportInfo represents an import with alias management.
type ImportInfo struct {
	Path       string // "github.com/bozz33/sublimego/internal/resources/user"
	Alias      string // "resource_user" (empty if no alias)
	NeedsAlias bool   // true if an alias is needed
	Package    string // "user"
}

// ResourceInfo represents a resource for generation.
type ResourceInfo struct {
	Reference string // "user.UserResource" or "resource_user.Resource"
	Source    string // "internal/resources/user/resource.go"
	Alias     string // Alias used if needed
	Conflict  bool   // True if this resource has a conflict
}

// PageInfo represents a page for generation.
type PageInfo struct {
	Reference string // "settings.SettingsPage"
	Source    string // "internal/pages/settings/page.go"
	Alias     string // Alias used if needed
	Conflict  bool   // True if this page has a conflict
}

// PageMetadata contains metadata for a discovered page.
type PageMetadata struct {
	TypeName    string
	PackageName string
	FilePath    string
	Slug        string
}

// TemplateData contains all data for the template.
type TemplateData struct {
	Timestamp   string         // "2024-01-30 11:53:00"
	Count       int            // Number of resources
	PageCount   int            // Number of pages
	Imports     []ImportInfo   // Required imports
	PageImports []ImportInfo   // Page imports
	Resources   []ResourceInfo // Resources to generate
	Pages       []PageInfo     // Pages to generate
	Warnings    []string       // Educational warnings
	Conflicts   []Conflict     // Detected conflicts
	Generated   time.Time      // Generation date
}

// ScannerConfig contains the scanner configuration.
type ScannerConfig struct {
	ResourcesPath   string   // Path to resources
	PagesPath       string   // Path to pages
	OutputPath      string   // Path to generated file
	TemplatePath    string   // Path to template
	StrictMode      bool     // Strict mode (error on warnings)
	AutoFix         bool     // Auto-fix conflicts
	Verbose         bool     // Detailed output
	DryRun          bool     // Dry-run mode
	ExcludePatterns []string // Patterns to exclude
}

// DefaultConfig returns the default configuration.
func DefaultConfig() ScannerConfig {
	return ScannerConfig{
		ResourcesPath:   "internal/resources",
		PagesPath:       "internal/pages",
		OutputPath:      "internal/registry/provider_gen.go",
		TemplatePath:    "templates/provider.go.tmpl",
		StrictMode:      false,
		AutoFix:         true,
		Verbose:         false,
		DryRun:          false,
		ExcludePatterns: []string{"*_test.go", "*_gen.go"},
	}
}

// ScanResult contains the scan result.
type ScanResult struct {
	Resources []ResourceMetadata
	Pages     []PageMetadata
	Conflicts []Conflict
	Success   bool
	Message   string
	Duration  time.Duration
}

// GenerationResult contains the generation result.
type GenerationResult struct {
	FilePath     string
	BytesWritten int
	Success      bool
	Message      string
	Warnings     []string
	Conflicts    []Conflict
	Duration     time.Duration
}
