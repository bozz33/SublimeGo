// Package generator provides code generation for SublimeGo resources and pages.
//
// It generates boilerplate code for resources, pages, forms, tables, and Ent schemas
// using Go templates. The generator is used by the CLI to scaffold new
// components quickly.
//
// Features:
//   - Resource generation (CRUD with form/table)
//   - Custom page generation (standalone views)
//   - Ent schema generation
//   - Form and table templates
//   - Migration and seeder generation
//   - Customizable templates
//   - Force overwrite and backup options
//
// Generate a Resource:
//
//	gen, err := generator.New(&generator.Options{
//		Force:   false,
//		Verbose: true,
//	})
//
//	// Generate a complete resource (resource.go, table.go, form.go, schema.go)
//	err = generator.GenerateResource(gen, "Product", projectPath)
//
// Generate a Custom Page:
//
//	// Generate a page with default options
//	err = generator.GeneratePage(gen, "Settings", projectPath)
//
//	// Generate a page with custom options (group, icon, sort order)
//	err = generator.GeneratePageWithOptions(gen, "Analytics", projectPath, "Reports", "chart", 50)
//
// Generated Page Structure:
//
//	internal/pages/settings/
//	├── page.go         # Page struct with Render(), GetForm(), GetTable()
//	└── content.templ   # Templ template for the page content
//
// Page Features:
//   - Full access to Form Builder (TextInput, Select, Checkbox, etc.)
//   - Full access to Table Builder (columns, sorting, pagination)
//   - Custom access control (CanAccess method)
//   - Navigation integration (icon, group, sort order)
//   - Automatic registration via scanner
package generator
