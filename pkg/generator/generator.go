package generator

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/samber/lo"
)

//go:embed stubs/resource.go.tmpl
var resourceTemplate string

//go:embed stubs/schema.go.tmpl
var schemaTemplate string

//go:embed stubs/table.go.tmpl
var tableTemplate string

//go:embed stubs/form.go.tmpl
var formTemplate string

//go:embed stubs/page.go.tmpl
var pageTemplate string

//go:embed stubs/page_templ.go.tmpl
var pageTemplTemplate string

// Generator handles code generation with embedded templates.
type Generator struct {
	templates map[string]*template.Template
	options   *Options
}

// Options configures the generator behavior.
type Options struct {
	Force     bool
	DryRun    bool
	Skip      []string
	NoBackup  bool
	Verbose   bool
	OutputDir string
}

// New creates a new generator with embedded templates.
func New(opts *Options) (*Generator, error) {
	if opts == nil {
		opts = &Options{}
	}

	g := &Generator{
		templates: make(map[string]*template.Template),
		options:   opts,
	}

	funcMap := template.FuncMap{
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"pascal":     ToPascalCase,
		"snake":      ToSnakeCase,
		"camel":      ToCamelCase,
		"plural":     Pluralize,
		"singular":   Singularize,
		"title":      strings.Title,
		"now":        time.Now,
		"formatTime": func(t time.Time) string { return t.Format("2006-01-02 15:04:05") },
	}

	templates := map[string]string{
		"resource":   resourceTemplate,
		"schema":     schemaTemplate,
		"table":      tableTemplate,
		"form":       formTemplate,
		"page":       pageTemplate,
		"page_templ": pageTemplTemplate,
	}

	for name, content := range templates {
		tmpl, err := template.New(name).Funcs(funcMap).Parse(content)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
		}
		g.templates[name] = tmpl
	}

	return g, nil
}

// HasTemplate checks if a template exists.
func (g *Generator) HasTemplate(name string) bool {
	_, exists := g.templates[name]
	return exists
}

// Generate creates a file from a template.
func (g *Generator) Generate(templateName, outputPath string, data interface{}) error {
	if fileExists(outputPath) && !g.options.Force {
		return fmt.Errorf("file already exists: %s (use --force to overwrite)", outputPath)
	}

	if g.options.DryRun {
		if g.options.Verbose {
			fmt.Printf("📄 Would generate: %s\n", outputPath)
		}
		return nil
	}

	if fileExists(outputPath) && !g.options.NoBackup {
		if err := g.backup(outputPath); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
	}

	tmpl, exists := g.templates[templateName]
	if !exists {
		return fmt.Errorf("template not found: %s", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("template execution failed: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		formatted = buf.Bytes()
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(outputPath, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if g.options.Verbose {
		fmt.Printf("Generated: %s (%d bytes)\n", outputPath, len(formatted))
	}

	return nil
}

// backup creates a backup copy of a file.
func (g *Generator) backup(path string) error {
	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s.backup_%s", path, timestamp)

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return err
	}

	if g.options.Verbose {
		fmt.Printf("💾 Backup: %s\n", filepath.Base(backupPath))
	}

	return nil
}

// shouldSkip checks if a file should be skipped.
func (g *Generator) shouldSkip(filename string) bool {
	return lo.Contains(g.options.Skip, filename)
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ToPascalCase converts a string to PascalCase.
func ToPascalCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})

	result := ""
	for _, word := range words {
		if len(word) > 0 {
			result += strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	return result
}

// ToSnakeCase converts a string to snake_case.
func ToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// ToCamelCase converts a string to camelCase.
func ToCamelCase(s string) string {
	pascal := ToPascalCase(s)
	if len(pascal) == 0 {
		return ""
	}
	return strings.ToLower(pascal[:1]) + pascal[1:]
}

// Pluralize converts a word to its plural form.
func Pluralize(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	special := map[string]string{
		"person": "people",
		"child":  "children",
		"mouse":  "mice",
		"tooth":  "teeth",
		"foot":   "feet",
		"man":    "men",
		"woman":  "women",
	}

	lower := strings.ToLower(s)
	if plural, ok := special[lower]; ok {
		return plural
	}

	if strings.HasSuffix(lower, "y") && !isVowel(rune(lower[len(lower)-2])) {
		return s[:len(s)-1] + "ies"
	}

	if strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "x") ||
		strings.HasSuffix(lower, "z") || strings.HasSuffix(lower, "ch") ||
		strings.HasSuffix(lower, "sh") {
		return s + "es"
	}

	return s + "s"
}

// Singularize converts a word to its singular form.
func Singularize(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	special := map[string]string{
		"people":   "person",
		"children": "child",
		"mice":     "mouse",
		"teeth":    "tooth",
		"feet":     "foot",
		"men":      "man",
		"women":    "woman",
	}

	lower := strings.ToLower(s)
	if singular, ok := special[lower]; ok {
		return singular
	}

	if strings.HasSuffix(lower, "ies") {
		return s[:len(s)-3] + "y"
	}

	if strings.HasSuffix(lower, "es") {
		return s[:len(s)-2]
	}

	if strings.HasSuffix(lower, "s") {
		return s[:len(s)-1]
	}

	return s
}

// isVowel checks if a character is a vowel.
func isVowel(r rune) bool {
	vowels := "aeiouAEIOU"
	return strings.ContainsRune(vowels, r)
}
