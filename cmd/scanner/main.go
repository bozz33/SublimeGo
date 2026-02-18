package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bozz33/sublimego/internal/scanner"
)

func main() {
	// Parse command line flags
	var (
		outputPath    = flag.String("output", "internal/registry/provider_gen.go", "Output file path")
		resourcesPath = flag.String("resources", "internal/resources", "Resources directory path")
		verbose       = flag.Bool("verbose", false, "Verbose output")
		dryRun        = flag.Bool("dry-run", false, "Dry run mode")
		autoFix       = flag.Bool("auto-fix", true, "Auto-fix conflicts")
		strict        = flag.Bool("strict", false, "Strict mode")
	)
	flag.Parse()

	// Configuration
	config := scanner.DefaultConfig()
	config.OutputPath = *outputPath
	config.ResourcesPath = *resourcesPath
	config.Verbose = *verbose
	config.DryRun = *dryRun
	config.AutoFix = *autoFix
	config.StrictMode = *strict

	// Create the scanner
	s := scanner.NewWithConfig(config)

	// Scan resources
	fmt.Printf("Scanning %s...\n", config.ResourcesPath)
	result := s.Scan()
	if !result.Success {
		log.Fatalf("Scan failed: %s", result.Message)
	}

	if len(result.Resources) == 0 {
		fmt.Println("No resources found")
		os.Exit(0)
	}

	fmt.Printf("Found %d resource(s)\n", len(result.Resources))
	if *verbose {
		for _, resource := range result.Resources {
			fmt.Printf("   - %s.%s (slug: %s)\n", resource.PackageName, resource.TypeName, resource.Slug)
		}
	}

	// Display conflicts
	if len(result.Conflicts) > 0 {
		detector := scanner.NewDetector(result.Resources)
		warnings := detector.FilterBySeverity(result.Conflicts, "warning")
		errors := detector.FilterBySeverity(result.Conflicts, "error")

		if len(warnings) > 0 {
			fmt.Printf("%d warning(s)\n", len(warnings))
		}
		if len(errors) > 0 {
			fmt.Printf("%d error(s)\n", len(errors))
			if !*autoFix {
				log.Fatalf("Blocking errors detected and auto-fix disabled")
			}
		}
	}

	fmt.Printf("Generating %s...\n", config.OutputPath)
	generator := scanner.NewGeneratorWithConfig(config)
	genResult := generator.Generate(result)
	if !genResult.Success {
		log.Fatalf("Generation failed: %s", genResult.Message)
	}

	fmt.Printf("%s\n", genResult.Message)
	if *verbose {
		fmt.Printf("Scan: %v, Generation: %v, Total: %v\n",
			result.Duration, genResult.Duration, result.Duration+genResult.Duration)
	}
}
