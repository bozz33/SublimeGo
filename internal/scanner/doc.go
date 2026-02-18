// Package scanner provides automatic resource discovery and code generation.
//
// It scans the project directory for resource definitions and generates
// the provider registration code. This enables automatic resource discovery
// without manual registration.
//
// Features:
//   - Automatic resource scanning
//   - Provider code generation
//   - Conflict detection (duplicate names, etc.)
//   - Import management
//   - Template-based generation
//
// Basic usage:
//
//	scanner := scanner.New(&scanner.Config{
//		ResourcesDir: "internal/resources",
//		OutputFile:   "internal/registry/provider_gen.go",
//	})
//
//	// Scan and generate
//	result, err := scanner.ScanAndGenerate()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Found %d resources\n", len(result.Resources))
package scanner
