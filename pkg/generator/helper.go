package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ResourceData contains the data to generate a resource.
type ResourceData struct {
	PackageName string // user
	Name        string // User
	TypeName    string // UserResource
	EntTypeName string // User
	Slug        string // users
	Label       string // User
	PluralLabel string // Users
	Icon        string // users
}

// NewResourceData creates the data for a resource.
func NewResourceData(name string) *ResourceData {
	packageName := ToSnakeCase(name)
	typeName := ToPascalCase(name) + "Resource"
	entTypeName := ToPascalCase(name)
	slug := Pluralize(ToSnakeCase(name))
	label := ToPascalCase(name)
	pluralLabel := Pluralize(label)

	return &ResourceData{
		PackageName: packageName,
		Name:        name,
		TypeName:    typeName,
		EntTypeName: entTypeName,
		Slug:        slug,
		Label:       label,
		PluralLabel: pluralLabel,
		Icon:        slug,
	}
}

// GenerateResource generates all files for a resource.
func GenerateResource(g *Generator, name, outputDir string) error {
	data := NewResourceData(name)

	resourceDir := filepath.Join(outputDir, "internal", "resources", data.PackageName)

	schemaDir := filepath.Join(outputDir, "internal", "ent", "schema")

	files := map[string]string{
		"resource": filepath.Join(resourceDir, "resource.go"),
		"schema":   filepath.Join(schemaDir, data.PackageName+".go"),
		"table":    filepath.Join(resourceDir, "table.go"),
		"form":     filepath.Join(resourceDir, "form.go"),
	}

	stats := struct {
		Generated int
		Skipped   int
		Failed    int
	}{}

	for templateName, outputPath := range files {
		if g.shouldSkip(templateName) {
			if g.options.Verbose {
				fmt.Printf("Skipped: %s\n", filepath.Base(outputPath))
			}
			stats.Skipped++
			continue
		}

		if err := g.Generate(templateName, outputPath, data); err != nil {
			if g.options.Verbose {
				fmt.Printf("Failed: %s - %v\n", filepath.Base(outputPath), err)
			}
			stats.Failed++
			return err
		}

		stats.Generated++
	}

	if g.options.Verbose {
		fmt.Printf("\n✨ Statistics:\n")
		fmt.Printf("   - Generated: %d files\n", stats.Generated)
		if stats.Skipped > 0 {
			fmt.Printf("   - Skipped: %d files\n", stats.Skipped)
		}
		if stats.Failed > 0 {
			fmt.Printf("   - Failed: %d files\n", stats.Failed)
		}
	}

	return nil
}

// GenerateMigration generates a migration file.
func GenerateMigration(name, outputDir string) error {
	timestamp := fmt.Sprintf("%d", timeNow().Unix())
	filename := fmt.Sprintf("%s_%s.sql", timestamp, ToSnakeCase(name))
	outputPath := filepath.Join(outputDir, "migrations", filename)

	content := fmt.Sprintf(`-- Migration: %s
-- Created at: %s

-- TODO: Add SQL commands here

-- Example:
-- CREATE TABLE IF NOT EXISTS %s (
--     id INTEGER PRIMARY KEY AUTOINCREMENT,
--     name TEXT NOT NULL,
--     created_at DATETIME DEFAULT CURRENT_TIMESTAMP
-- );
`, name, timeNow().Format("2006-01-02 15:04:05"), Pluralize(ToSnakeCase(name)))

	if err := ensureDir(filepath.Dir(outputPath)); err != nil {
		return err
	}

	return writeFile(outputPath, []byte(content))
}

// GenerateSeeder generates a seeder file.
func GenerateSeeder(name, outputDir string) error {
	packageName := ToSnakeCase(name)
	typeName := ToPascalCase(name)
	filename := fmt.Sprintf("%s_seeder.go", packageName)
	outputPath := filepath.Join(outputDir, "seeders", filename)

	content := fmt.Sprintf(`package seeders

import (
	"context"
	"log"

	"github.com/bozz33/SublimeGo/internal/ent"
)

// Seed%s inserts test data for %s
func Seed%s(ctx context.Context, client *ent.Client) error {
	log.Println("Seeding %s...")
	
	// TODO: Add seeding logic here
	// Example:
	// _, err := client.%s.Create().
	// 	SetName("Example").
	// 	Save(ctx)
	// if err != nil {
	// 	return err
	// }
	
	log.Println("%s seeded successfully")
	return nil
}
`, typeName, name, typeName, name, typeName, typeName)

	if err := ensureDir(filepath.Dir(outputPath)); err != nil {
		return err
	}

	return writeFile(outputPath, []byte(content))
}

// Internal helpers

func ensureDir(dir string) error {
	return mkdirAll(dir, 0755)
}

func writeFile(path string, content []byte) error {
	return osWriteFile(path, content, 0644)
}

// Variables to facilitate testing (dependency injection)
var (
	timeNow     = defaultTimeNow
	mkdirAll    = defaultMkdirAll
	osWriteFile = defaultWriteFile
)

func defaultTimeNow() timeInterface {
	return timeWrapper{timeNowFunc()}
}

func defaultMkdirAll(path string, perm uint32) error {
	return osMkdirAll(path, osPerm(perm))
}

func defaultWriteFile(path string, data []byte, perm uint32) error {
	return osWriteFileFunc(path, data, osPerm(perm))
}

// Interfaces for testing
type timeInterface interface {
	Unix() int64
	Format(string) string
}

type timeWrapper struct {
	t timeType
}

func (tw timeWrapper) Unix() int64 {
	return tw.t.Unix()
}

func (tw timeWrapper) Format(layout string) string {
	return tw.t.Format(layout)
}

type timeType = time.Time
type osPerm = os.FileMode

var (
	timeNowFunc     = time.Now
	osMkdirAll      = os.MkdirAll
	osWriteFileFunc = os.WriteFile
)
