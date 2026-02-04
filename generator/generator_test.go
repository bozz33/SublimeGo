package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	g, err := New(nil)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if g == nil {
		t.Fatal("New() returned nil generator")
	}

	if len(g.templates) != 6 {
		t.Errorf("Expected 6 templates, got %d", len(g.templates))
	}
}

func TestHasTemplate(t *testing.T) {
	g, _ := New(nil)

	tests := []struct {
		name     string
		template string
		want     bool
	}{
		{"resource exists", "resource", true},
		{"schema exists", "schema", true},
		{"table exists", "table", true},
		{"form exists", "form", true},
		{"page exists", "page", true},
		{"page_templ exists", "page_templ", true},
		{"unknown template", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := g.HasTemplate(tt.template); got != tt.want {
				t.Errorf("HasTemplate(%q) = %v, want %v", tt.template, got, tt.want)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.go")

	g, _ := New(&Options{})

	data := &ResourceData{
		PackageName: "test",
		Name:        "Test",
		TypeName:    "TestResource",
		EntTypeName: "Test",
		Slug:        "tests",
		Label:       "Test",
		PluralLabel: "Tests",
		Icon:        "test",
	}

	err := g.Generate("resource", outputPath, data)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Generated file does not exist")
	}
}

func TestGenerateWithForce(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.go")

	// Create an existing file
	os.WriteFile(outputPath, []byte("existing content"), 0644)

	// Without force, should fail
	g1, _ := New(&Options{Force: false})
	data := &ResourceData{
		PackageName: "test",
		Name:        "Test",
		TypeName:    "TestResource",
		EntTypeName: "Test",
		Slug:        "tests",
		Label:       "Test",
		PluralLabel: "Tests",
		Icon:        "test",
	}

	err := g1.Generate("resource", outputPath, data)
	if err == nil {
		t.Error("Expected error when file exists without force flag")
	}

	// With force, should succeed
	g2, _ := New(&Options{Force: true})
	err = g2.Generate("resource", outputPath, data)
	if err != nil {
		t.Errorf("Generate() with force failed: %v", err)
	}
}

func TestGenerateWithBackup(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.go")

	// Create an existing file
	originalContent := []byte("original content")
	os.WriteFile(outputPath, originalContent, 0644)

	// Generate with force and backup
	g, _ := New(&Options{Force: true, NoBackup: false, Verbose: true})
	data := &ResourceData{
		PackageName: "test",
		Name:        "Test",
		TypeName:    "TestResource",
		EntTypeName: "Test",
		Slug:        "tests",
		Label:       "Test",
		PluralLabel: "Tests",
		Icon:        "test",
	}

	err := g.Generate("resource", outputPath, data)
	if err != nil {
		t.Fatalf("Generate() with backup failed: %v", err)
	}

	// Verify file was overwritten
	newContent, _ := os.ReadFile(outputPath)
	if len(newContent) == len(originalContent) {
		t.Error("File should have been overwritten")
	}
}

func TestGenerateWithDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.go")

	g, _ := New(&Options{DryRun: true})
	data := &ResourceData{
		PackageName: "test",
		Name:        "Test",
		TypeName:    "TestResource",
		EntTypeName: "Test",
		Slug:        "tests",
		Label:       "Test",
		PluralLabel: "Tests",
		Icon:        "test",
	}

	err := g.Generate("resource", outputPath, data)
	if err != nil {
		t.Fatalf("Generate() in dry-run failed: %v", err)
	}

	// Le fichier ne devrait pas exister
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Error("File should not exist in dry-run mode")
	}
}

func TestGenerateWithSkip(t *testing.T) {
	g, _ := New(&Options{Skip: []string{"schema", "form"}})

	if g.shouldSkip("schema") {
		t.Log("schema correctly skipped")
	} else {
		t.Error("schema should be skipped")
	}

	if g.shouldSkip("resource") {
		t.Error("resource should not be skipped")
	}
}

func TestNewResourceData(t *testing.T) {
	data := NewResourceData("Product")

	tests := []struct {
		field string
		got   string
		want  string
	}{
		{"PackageName", data.PackageName, "product"},
		{"TypeName", data.TypeName, "ProductResource"},
		{"EntTypeName", data.EntTypeName, "Product"},
		{"Slug", data.Slug, "products"},
		{"Label", data.Label, "Product"},
		{"PluralLabel", data.PluralLabel, "Products"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.field, tt.got, tt.want)
			}
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"user", "users"},
		{"category", "categories"},
		{"person", "people"},
		{"child", "children"},
		{"mouse", "mice"},
		{"product", "products"},
		{"box", "boxes"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Pluralize(tt.input)
			if got != tt.want {
				t.Errorf("Pluralize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"user", "User"},
		{"user_profile", "UserProfile"},
		{"user-profile", "UserProfile"},
		{"user profile", "UserProfile"},
		{"UserProfile", "Userprofile"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToPascalCase(tt.input)
			if got != tt.want {
				t.Errorf("ToPascalCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"User", "user"},
		{"UserProfile", "user_profile"},
		{"HTTPServer", "h_t_t_p_server"},
		{"user", "user"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToSnakeCase(tt.input)
			if got != tt.want {
				t.Errorf("ToSnakeCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateResource(t *testing.T) {
	tmpDir := t.TempDir()

	g, _ := New(&Options{Verbose: false})

	err := GenerateResource(g, "Product", tmpDir)
	if err != nil {
		t.Fatalf("GenerateResource() failed: %v", err)
	}

	// Verify files were created
	expectedFiles := []string{
		"internal/resources/product/resource.go",
		"internal/ent/schema/product.go",
		"internal/resources/product/table.go",
		"internal/resources/product/form.go",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", file)
		}
	}
}

func BenchmarkGenerate(b *testing.B) {
	tmpDir := b.TempDir()
	g, _ := New(&Options{})
	data := NewResourceData("Product")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(tmpDir, "test.go")
		_ = g.Generate("resource", outputPath, data)
		_ = os.Remove(outputPath)
	}
}
