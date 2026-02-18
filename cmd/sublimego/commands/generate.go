package commands

import (
	"fmt"
	"path/filepath"

	"github.com/bozz33/sublimego/internal/scanner"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Génère le code auto-discovery pour les resources",
	Long: `Scanne le dossier internal/resources et génère automatiquement
le fichier provider_gen.go qui enregistre toutes les resou²rces.

Cette commande utilise l'analyse AST (Abstract Syntax Tree) pour découvrir
toutes les structures qui implémentent l'interface Resource.

Convention de nommage :
- Fichiers : *_resource.go (ex: user_resource.go)
- Structures : *Resource (ex: UserResource)
- Package : nom de la resource (ex: user)

Le fichier généré (provider_gen.go) contient :
- Import de toutes les resources découvertes
- Variable AllResources avec toutes les instances
- Fonction init() pour l'enregistrement automatique
- Fonctions helper pour accéder aux resources

Cette commande doit être exécutée :
- Après la création d'une nouvelle resource
- Après la suppression d'une resource
- Avant le build de production`,
	Example: `  # Générer le provider pour toutes les resources
  sublimego generate
  sublimego gen

  # Vérifier les resources découvertes (dry-run)
  sublimego generate --dry-run

  # Générer avec sortie détaillée
  sublimego generate --verbose`,
	Aliases: []string{"gen"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()

		// Configuration du scanner
		resourcesPath := filepath.Join("internal", "resources")
		outputPath := filepath.Join("internal", "registry", "provider_gen.go")

		scannerConfig := scanner.DefaultConfig()
		scannerConfig.ResourcesPath = resourcesPath
		scannerConfig.OutputPath = outputPath
		scannerConfig.Verbose = verbose
		scannerConfig.DryRun = dryRun
		scannerConfig.AutoFix = true

		fmt.Println("🔍 Wire Pattern - Auto-Discovery Intelligent")
		fmt.Println("=============================================")
		fmt.Println()

		// Créer le scanner avec configuration
		s := scanner.NewWithConfig(scannerConfig)

		// Scanner les resources avec détection de conflits
		fmt.Printf("Scanning %s...\n", resourcesPath)
		result := s.Scan()
		if !result.Success {
			return fmt.Errorf("scan failed: %s", result.Message)
		}

		if len(result.Resources) == 0 {
			fmt.Println("No resources found.")
			fmt.Println()
			fmt.Println("Make sure you have resources in internal/resources/")
			fmt.Println("Resources must:")
			fmt.Println("  - Be in files named *_resource.go")
			fmt.Println("  - Have types named *Resource (or any type)")
			fmt.Println("  - Implement the engine.Resource interface")
			return nil
		}

		// Afficher les resources découvertes
		fmt.Printf("\n📦 Discovered %d resource(s):\n", len(result.Resources))
		for _, m := range result.Resources {
			fmt.Printf("  - %s.%s (slug: %s)\n", m.PackageName, m.TypeName, m.Slug)
		}
		fmt.Println()

		// Afficher les conflits détectés
		if len(result.Conflicts) > 0 {
			detector := scanner.NewDetector(result.Resources)

			// Warnings
			warnings := detector.FilterBySeverity(result.Conflicts, "warning")
			if len(warnings) > 0 {
				fmt.Printf("%d warning(s) detected:\n", len(warnings))
				for _, warning := range warnings {
					fmt.Printf("   - %s\n", warning.Message)
					if verbose {
						fmt.Printf("     Suggestion: %s\n", warning.Suggestion)
					}
				}
				fmt.Println()
			}

			// Erreurs
			errors := detector.FilterBySeverity(result.Conflicts, "error")
			if len(errors) > 0 {
				fmt.Printf("%d error(s) detected:\n", len(errors))
				for _, err := range errors {
					fmt.Printf("   - %s\n", err.Message)
					if verbose {
						fmt.Printf("     Suggestion: %s\n", err.Suggestion)
						fmt.Printf("     Auto-fix: %v\n", err.AutoFix)
					}
				}
				fmt.Println()

				// Vérifier si on peut continuer
				if !scannerConfig.AutoFix {
					return fmt.Errorf("blocking errors detected and auto-fix disabled")
				}

				fmt.Printf("Auto-fix enabled - resolving conflicts...\n\n")
			}
		}

		// Mode dry-run
		if dryRun {
			fmt.Println("Dry-run mode: no files will be generated")
			return nil
		}

		// Générer le fichier provider_gen.go
		fmt.Printf("Generating %s...\n", outputPath)
		gen := scanner.NewGeneratorWithConfig(scannerConfig)
		genResult := gen.Generate(result)
		if !genResult.Success {
			return fmt.Errorf("generation failed: %s", genResult.Message)
		}

		fmt.Printf("%s\n", genResult.Message)
		fmt.Println()

		// Afficher les statistiques
		if verbose {
			fmt.Printf("Statistics:\n")
			fmt.Printf("   - Scan duration: %v\n", result.Duration)
			fmt.Printf("   - Generation duration: %v\n", genResult.Duration)
			fmt.Printf("   - Total duration: %v\n", result.Duration+genResult.Duration)
			fmt.Printf("   - File size: %d bytes\n", genResult.BytesWritten)
			fmt.Println()
		}

		fmt.Printf("Next steps:\n")
		fmt.Printf("  1. Review %s\n", outputPath)
		fmt.Printf("  2. Run: go build\n")
		fmt.Printf("  3. Run: sublimego serve\n")

		// Afficher un avertissement si la config est chargée
		if cfg != nil && verbose {
			fmt.Println()
			fmt.Printf("Resources will be available at: %s\n", cfg.Engine.BasePath)
		}

		return nil
	},
}

var (
	dryRun     bool
	strictMode bool
	autoFix    bool
)

func init() {
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be generated without writing files")
	generateCmd.Flags().BoolVar(&strictMode, "strict", false, "fail on warnings and errors")
	generateCmd.Flags().BoolVar(&autoFix, "auto-fix", true, "automatically fix conflicts when possible")
}
