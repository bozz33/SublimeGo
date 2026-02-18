package scanner

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// titleCaser is a package-level caser to avoid re-allocating on every call.
var titleCaser = cases.Title(language.English)

// ConflictType represents the type of conflict detected.
type ConflictType int

const (
	ConflictNone ConflictType = iota
	ConflictDuplicateName
	ConflictGenericName
	ConflictNamingConvention
	ConflictPackageConflict
)

// Conflict detects a conflict between resources.
type Conflict struct {
	Type       ConflictType
	Severity   string // "error", "warning", "info"
	Message    string
	Suggestion string
	DocsURL    string
	Resources  []ResourceMetadata
	AutoFix    bool
}

// Detector analyzes resources to detect conflicts.
type Detector struct {
	resources []ResourceMetadata
}

// NewDetector creates a new detector.
func NewDetector(resources []ResourceMetadata) *Detector {
	return &Detector{
		resources: resources,
	}
}

// Detect analyzes all possible conflicts.
func (d *Detector) Detect() []Conflict {
	var conflicts []Conflict

	conflicts = append(conflicts, d.detectDuplicateNames()...)

	conflicts = append(conflicts, d.detectGenericNames()...)

	conflicts = append(conflicts, d.detectNamingConventions()...)

	conflicts = append(conflicts, d.detectPackageConflicts()...)

	return conflicts
}

// detectDuplicateNames detects duplicate type names.
func (d *Detector) detectDuplicateNames() []Conflict {
	var conflicts []Conflict

	grouped := lo.GroupBy(d.resources, func(r ResourceMetadata) string {
		return r.TypeName
	})

	for typeName, resources := range grouped {
		if len(resources) > 1 {
			conflict := Conflict{
				Type:       ConflictDuplicateName,
				Severity:   "error",
				Message:    fmt.Sprintf("Duplicate type name '%s' found in %d packages", typeName, len(resources)),
				Suggestion: "Rename types to be unique across all packages",
				DocsURL:    "https://docs.sublimego.dev/resources/naming",
				Resources:  resources,
				AutoFix:    true,
			}

			var aliases []string
			for _, r := range resources {
				alias := d.generateAlias(r)
				aliases = append(aliases, fmt.Sprintf("%s.%s → %s.%s", r.PackageName, r.TypeName, alias, r.TypeName))
			}
			conflict.Suggestion += fmt.Sprintf("\nAuto-fix aliases: %s", strings.Join(aliases, ", "))

			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts
}

// detectGenericNames detects generic names.
func (d *Detector) detectGenericNames() []Conflict {
	var conflicts []Conflict

	genericNames := []string{"Resource", "Entity", "Model", "Item", "Object"}

	for _, resource := range d.resources {
		if lo.Contains(genericNames, resource.TypeName) {
			conflict := Conflict{
				Type:       ConflictGenericName,
				Severity:   "warning",
				Message:    fmt.Sprintf("Generic type name '%s' should be more specific", resource.TypeName),
				Suggestion: fmt.Sprintf("Rename '%s' to '%s%s'", resource.TypeName, titleCaser.String(resource.PackageName), resource.TypeName),
				DocsURL:    "https://docs.sublimego.dev/resources/naming",
				Resources:  []ResourceMetadata{resource},
				AutoFix:    true,
			}

			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts
}

// detectNamingConventions detects convention violations.
func (d *Detector) detectNamingConventions() []Conflict {
	var conflicts []Conflict

	for _, resource := range d.resources {
		expectedName := fmt.Sprintf("%s%s", titleCaser.String(resource.PackageName), "Resource")

		if resource.TypeName != expectedName {
			conflict := Conflict{
				Type:       ConflictNamingConvention,
				Severity:   "info",
				Message:    fmt.Sprintf("Type '%s' doesn't follow naming convention", resource.TypeName),
				Suggestion: fmt.Sprintf("Consider renaming to '%s' for consistency", expectedName),
				DocsURL:    "https://docs.sublimego.dev/resources/naming",
				Resources:  []ResourceMetadata{resource},
				AutoFix:    false,
			}

			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts
}

// detectPackageConflicts detects package conflicts.
func (d *Detector) detectPackageConflicts() []Conflict {
	var conflicts []Conflict

	grouped := lo.GroupBy(d.resources, func(r ResourceMetadata) string {
		return r.PackageName
	})

	for pkgName, resources := range grouped {
		if len(resources) > 1 {
			typeNames := lo.Map(resources, func(r ResourceMetadata, _ int) string {
				return r.TypeName
			})

			if len(lo.Uniq(typeNames)) != len(typeNames) {
				conflict := Conflict{
					Type:       ConflictPackageConflict,
					Severity:   "error",
					Message:    fmt.Sprintf("Multiple resources with same name in package '%s'", pkgName),
					Suggestion: "Rename resources to be unique within their package",
					DocsURL:    "https://docs.sublimego.dev/resources/organization",
					Resources:  resources,
					AutoFix:    false,
				}

				conflicts = append(conflicts, conflict)
			}
		}
	}

	return conflicts
}

// generateAlias generates a unique alias for a resource.
func (d *Detector) generateAlias(resource ResourceMetadata) string {
	alias := fmt.Sprintf("%s_%s", resource.PackageName, strings.ToLower(resource.TypeName))

	if d.isAliasUnique(alias, resource) {
		return alias
	}

	counter := 1
	for {
		candidate := fmt.Sprintf("%s_%d", alias, counter)
		if d.isAliasUnique(candidate, resource) {
			return candidate
		}
		counter++
	}
}

// isAliasUnique checks if an alias is unique.
func (d *Detector) isAliasUnique(alias string, exclude ResourceMetadata) bool {
	for _, r := range d.resources {
		if r.PackageName == exclude.PackageName && r.TypeName == exclude.TypeName {
			continue
		}
		if fmt.Sprintf("%s_%s", r.PackageName, strings.ToLower(r.TypeName)) == alias {
			return false
		}
	}
	return true
}

// HasErrors returns true if there are blocking errors.
func (d *Detector) HasErrors(conflicts []Conflict) bool {
	return lo.SomeBy(conflicts, func(c Conflict) bool {
		return c.Severity == "error"
	})
}

// HasWarnings returns true if there are warnings.
func (d *Detector) HasWarnings(conflicts []Conflict) bool {
	return lo.SomeBy(conflicts, func(c Conflict) bool {
		return c.Severity == "warning"
	})
}

// FilterBySeverity filters conflicts by severity.
func (d *Detector) FilterBySeverity(conflicts []Conflict, severity string) []Conflict {
	return lo.Filter(conflicts, func(c Conflict, _ int) bool {
		return c.Severity == severity
	})
}

// GetAutoFixable returns auto-fixable conflicts.
func (d *Detector) GetAutoFixable(conflicts []Conflict) []Conflict {
	return lo.Filter(conflicts, func(c Conflict, _ int) bool {
		return c.AutoFix
	})
}
