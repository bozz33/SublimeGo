package generator

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"text/template"
)

//go:embed stubs/widget.go.tmpl
var widgetTemplate string

//go:embed stubs/action.go.tmpl
var actionTemplate string

//go:embed stubs/enum.go.tmpl
var enumTemplate string

// registerExtraTemplates adds widget/action/enum templates to an existing Generator.
func registerExtraTemplates(g *Generator) error {
	funcMap := template.FuncMap{
		"lower":  toLower,
		"upper":  toUpper,
		"pascal": ToPascalCase,
		"snake":  ToSnakeCase,
		"camel":  ToCamelCase,
		"plural": Pluralize,
	}
	extras := map[string]string{
		"widget": widgetTemplate,
		"action": actionTemplate,
		"enum":   enumTemplate,
	}
	for name, content := range extras {
		tmpl, err := template.New(name).Funcs(funcMap).Parse(content)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}
		g.templates[name] = tmpl
	}
	return nil
}

func toLower(s string) string { return ToSnakeCase(s) }
func toUpper(s string) string {
	r := []byte(s)
	for i, b := range r {
		if b >= 'a' && b <= 'z' {
			r[i] = b - 32
		}
	}
	return string(r)
}

// GenerateWidget generates a widget file.
func GenerateWidget(g *Generator, name, outputDir string) error {
	if err := registerExtraTemplates(g); err != nil {
		return err
	}
	data := NewResourceData(name)
	pkgDir := filepath.Join(outputDir, "internal", "widgets", data.PackageName)
	outputPath := filepath.Join(pkgDir, "widget.go")
	if err := g.Generate("widget", outputPath, data); err != nil {
		return fmt.Errorf("failed to generate widget: %w", err)
	}
	fmt.Printf("Widget '%s' generated: %s\n", name, outputPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit widget.go to implement your data provider")
	fmt.Println("  2. Register the widget in your panel: panel.WithWidgets(New" + data.TypeName + "Widget())")
	return nil
}

// GenerateAction generates an action file.
func GenerateAction(g *Generator, name, outputDir string) error {
	if err := registerExtraTemplates(g); err != nil {
		return err
	}
	data := NewResourceData(name)
	pkgDir := filepath.Join(outputDir, "internal", "actions")
	outputPath := filepath.Join(pkgDir, ToSnakeCase(name)+"_action.go")
	if err := g.Generate("action", outputPath, data); err != nil {
		return fmt.Errorf("failed to generate action: %w", err)
	}
	fmt.Printf("Action '%s' generated: %s\n", name, outputPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit the action handler to implement your logic")
	fmt.Println("  2. Add the action to your resource: .HeaderActions(" + data.TypeName + "Action)")
	return nil
}

// GenerateEnum generates an enum file.
func GenerateEnum(g *Generator, name, outputDir string) error {
	if err := registerExtraTemplates(g); err != nil {
		return err
	}
	data := NewResourceData(name)
	pkgDir := filepath.Join(outputDir, "internal", "enums")
	outputPath := filepath.Join(pkgDir, ToSnakeCase(name)+".go")
	if err := g.Generate("enum", outputPath, data); err != nil {
		return fmt.Errorf("failed to generate enum: %w", err)
	}
	fmt.Printf("Enum '%s' generated: %s\n", name, outputPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Rename Option1/Option2/Option3 to your actual values")
	fmt.Println("  2. Use with form: form.Select(\"status\").Options(enum.Options(All" + data.TypeName + "Values))")
	fmt.Println("  3. Use with table: table.Badge(\"status\").Colors(enum.Colors(All" + data.TypeName + "Values))")
	return nil
}

// GenerateInfolist generates an infolist view file for an existing resource.
func GenerateInfolist(g *Generator, name, outputDir string) error {
	data := NewResourceData(name)
	pkgDir := filepath.Join(outputDir, "internal", "resources", data.PackageName)
	outputPath := filepath.Join(pkgDir, "view.go")
	if err := g.Generate("infolist", outputPath, data); err != nil {
		return fmt.Errorf("failed to generate infolist: %w", err)
	}
	fmt.Printf("Infolist view '%s' generated: %s\n", name, outputPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit view.go to add your infolist entries")
	fmt.Println("  2. Implement engine.ResourceViewable on your resource")
	return nil
}
