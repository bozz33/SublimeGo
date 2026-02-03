package scanner

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/samber/lo"
)

// ResourceMetadata contains metadata for a discovered resource.
type ResourceMetadata struct {
	TypeName    string
	PackageName string
	FilePath    string
	Slug        string
}

// Scanner analyzes source code to discover resources.
type Scanner struct {
	config ScannerConfig
	fset   *token.FileSet
}

// New creates a new scanner with default configuration.
func New(resourcesPath string) *Scanner {
	config := DefaultConfig()
	config.ResourcesPath = resourcesPath
	return &Scanner{
		config: config,
		fset:   token.NewFileSet(),
	}
}

// NewWithConfig creates a new scanner with custom configuration.
func NewWithConfig(config ScannerConfig) *Scanner {
	return &Scanner{
		config: config,
		fset:   token.NewFileSet(),
	}
}

// Scan analyzes all Go files with conflict detection.
func (s *Scanner) Scan() ScanResult {
	start := time.Now()
	var allMetadata []ResourceMetadata

	err := filepath.Walk(s.config.ResourcesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		for _, pattern := range s.config.ExcludePatterns {
			if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
				return nil
			}
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			metadata, err := s.scanFile(path)
			if err != nil {
				return fmt.Errorf("failed to scan %s: %w", path, err)
			}
			allMetadata = append(allMetadata, metadata...)
		}

		return nil
	})

	if err != nil {
		return ScanResult{
			Success:  false,
			Message:  fmt.Sprintf("Scan failed: %v", err),
			Duration: time.Since(start),
		}
	}

	detector := NewDetector(allMetadata)
	conflicts := detector.Detect()

	hasErrors := detector.HasErrors(conflicts)
	if hasErrors && s.config.StrictMode {
		return ScanResult{
			Success:   false,
			Message:   "Strict mode: blocking errors detected",
			Resources: allMetadata,
			Conflicts: conflicts,
			Duration:  time.Since(start),
		}
	}

	message := fmt.Sprintf("Scanned %d resources", len(allMetadata))
	if len(conflicts) > 0 {
		message += fmt.Sprintf(" (%d conflicts detected)", len(conflicts))
	}

	return ScanResult{
		Success:   true,
		Message:   message,
		Resources: allMetadata,
		Conflicts: conflicts,
		Duration:  time.Since(start),
	}
}

// scanFile analyzes a Go file to find resources.
func (s *Scanner) scanFile(filePath string) ([]ResourceMetadata, error) {
	node, err := parser.ParseFile(s.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	var metadata []ResourceMetadata
	packageName := node.Name.Name

	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			typeName := typeSpec.Name.Name

			if s.isPotentialResource(typeName) {
				slug := s.extractSlug(typeName)

				metadata = append(metadata, ResourceMetadata{
					TypeName:    typeName,
					PackageName: packageName,
					FilePath:    filePath,
					Slug:        slug,
				})
			}
		}
	}

	return metadata, nil
}

// isPotentialResource detects if a type could be a resource.
func (s *Scanner) isPotentialResource(typeName string) bool {
	if strings.HasSuffix(typeName, "Resource") {
		return true
	}

	// Generic names (detected with warning)
	genericNames := []string{"Resource", "Entity", "Model", "Item", "Object"}
	if lo.Contains(genericNames, typeName) {
		return true
	}

	return true
}

// extractSlug extracts the slug from the type name.
func (s *Scanner) extractSlug(typeName string) string {
	name := strings.TrimSuffix(typeName, "Resource")
	slug := strings.ToLower(name)

	switch {
	case strings.HasSuffix(slug, "y"):
		slug = slug[:len(slug)-1] + "ies"
	case strings.HasSuffix(slug, "s"):
		// Already plural
	case strings.HasSuffix(slug, "x") || strings.HasSuffix(slug, "ch") || strings.HasSuffix(slug, "sh"):
		slug += "es"
	default:
		slug += "s"
	}

	return slug
}

// BuildTemplateData builds data for the template.
func (s *Scanner) BuildTemplateData(result ScanResult) TemplateData {
	imports := s.buildImports(result.Resources, result.Conflicts)
	resources := s.buildResources(result.Resources, result.Conflicts)
	warnings := s.extractWarnings(result.Conflicts)

	return TemplateData{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Count:     len(result.Resources),
		Imports:   imports,
		Resources: resources,
		Warnings:  warnings,
		Conflicts: result.Conflicts,
		Generated: time.Now(),
	}
}

// buildImports builds the import list with aliases.
func (s *Scanner) buildImports(resources []ResourceMetadata, conflicts []Conflict) []ImportInfo {
	var imports []ImportInfo
	aliasMap := make(map[string]string)
	for _, conflict := range conflicts {
		if conflict.Type == ConflictDuplicateName && conflict.AutoFix {
			for _, resource := range conflict.Resources {
				alias := s.generateAlias(resource)
				aliasMap[resource.PackageName] = alias
			}
		}
	}

	for _, resource := range resources {
		importPath := fmt.Sprintf("github.com/bozz33/SublimeGo/internal/resources/%s", resource.PackageName)
		alias, needsAlias := aliasMap[resource.PackageName]

		imports = append(imports, ImportInfo{
			Path:       importPath,
			Alias:      alias,
			NeedsAlias: needsAlias,
			Package:    resource.PackageName,
		})
	}

	return lo.UniqBy(imports, func(i ImportInfo) string {
		return i.Path
	})
}

// buildResources builds the resource list with aliases.
func (s *Scanner) buildResources(resources []ResourceMetadata, conflicts []Conflict) []ResourceInfo {
	var result []ResourceInfo
	aliasMap := make(map[string]string)
	for _, conflict := range conflicts {
		if conflict.Type == ConflictDuplicateName && conflict.AutoFix {
			for _, resource := range conflict.Resources {
				alias := s.generateAlias(resource)
				key := fmt.Sprintf("%s.%s", resource.PackageName, resource.TypeName)
				aliasMap[key] = alias
			}
		}
	}

	for _, resource := range resources {
		key := fmt.Sprintf("%s.%s", resource.PackageName, resource.TypeName)
		alias, hasConflict := aliasMap[key]

		reference := fmt.Sprintf("%s.%s", resource.PackageName, resource.TypeName)
		if hasConflict {
			reference = fmt.Sprintf("%s.%s", alias, resource.TypeName)
		}

		result = append(result, ResourceInfo{
			Reference: reference,
			Source:    resource.FilePath,
			Alias:     alias,
			Conflict:  hasConflict,
		})
	}

	return result
}

// extractWarnings extracts warning messages from conflicts.
func (s *Scanner) extractWarnings(conflicts []Conflict) []string {
	var warnings []string

	for _, conflict := range conflicts {
		if conflict.Severity == "warning" || conflict.Severity == "info" {
			warnings = append(warnings, conflict.Message)
		}
	}

	return warnings
}

// generateAlias generates a unique alias for a resource.
func (s *Scanner) generateAlias(resource ResourceMetadata) string {
	alias := fmt.Sprintf("%s_%s", resource.PackageName, strings.ToLower(resource.TypeName))
	if s.isAliasUnique(alias, resource) {
		return alias
	}

	counter := 1
	for {
		candidate := fmt.Sprintf("%s_%d", alias, counter)
		if s.isAliasUnique(candidate, resource) {
			return candidate
		}
		counter++
	}
}

// isAliasUnique checks if an alias is unique.
func (s *Scanner) isAliasUnique(alias string, exclude ResourceMetadata) bool {
	return true
}

// GroupByPackage groups metadata by package.
func GroupByPackage(metadata []ResourceMetadata) map[string][]ResourceMetadata {
	return lo.GroupBy(metadata, func(m ResourceMetadata) string {
		return m.PackageName
	})
}

// FilterByPackage filters metadata by package.
func FilterByPackage(metadata []ResourceMetadata, packageName string) []ResourceMetadata {
	return lo.Filter(metadata, func(m ResourceMetadata, _ int) bool {
		return m.PackageName == packageName
	})
}

// ExtractTypeNames extracts type names.
func ExtractTypeNames(metadata []ResourceMetadata) []string {
	return lo.Map(metadata, func(m ResourceMetadata, _ int) string {
		return m.TypeName
	})
}

// ExtractSlugs extracts slugs.
func ExtractSlugs(metadata []ResourceMetadata) []string {
	return lo.Map(metadata, func(m ResourceMetadata, _ int) string {
		return m.Slug
	})
}
