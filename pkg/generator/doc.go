// Package generator provides code generation for SublimeGo resources.
//
// It generates boilerplate code for resources, forms, tables, and Ent schemas
// using Go templates. The generator is used by the CLI to scaffold new
// components quickly.
//
// Features:
//   - Resource generation (CRUD)
//   - Ent schema generation
//   - Form and table templates
//   - Migration and seeder generation
//   - Customizable templates
//   - Force overwrite and backup options
//
// Basic usage:
//
//	gen, err := generator.New(&generator.Options{
//		Force:   false,
//		Verbose: true,
//	})
//
//	// Generate a complete resource
//	err = generator.GenerateResource(gen, "Product", projectPath)
//
//	// Generate individual components
//	err = gen.Generate("resource", outputPath, data)
package generator
