package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bozz33/sublimego/generator"
	"github.com/spf13/cobra"
)

// Commande MAKE (parent pour les générateurs)

var makeCmd = &cobra.Command{
	Use:   "make",
	Short: "Generate code scaffolding",
	Long:  `Generate resources, migrations, seeders, and other code scaffolding.`,
}

// MAKE:RESOURCE - Génère une resource CRUD complète

var (
	forceFlag    bool
	dryRunFlag   bool
	skipFlag     []string
	noBackupFlag bool
	verboseFlag  bool
)

var makeResourceCmd = &cobra.Command{
	Use:     "resource [name]",
	Aliases: []string{"r"},
	Short:   "Generate a new CRUD resource",
	Long: `Generate a complete CRUD resource with:
- Resource file (resource.go)
- Schema file (schema.go)
- Table view (table.go)
- Form view (form.go)

Flags:
  --force      Overwrite existing files
  --dry-run    Preview without creating files
  --skip       Skip specific files (resource,schema,table,form)
  --no-backup  Disable automatic backups
  --verbose    Show detailed output

Example: sublimego make:resource Product
Example: sublimego make:resource Product --force --skip=form`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string

		// Mode interactif si aucun argument fourni
		if len(args) == 0 {
			fmt.Print("Nom de la resource (ex: Product): ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			name = strings.TrimSpace(input)
		} else {
			name = args[0]
		}

		if name == "" {
			return fmt.Errorf("resource name is required")
		}

		// Créer le générateur avec options
		g, err := generator.New(&generator.Options{
			Force:    forceFlag,
			DryRun:   dryRunFlag,
			Skip:     skipFlag,
			NoBackup: noBackupFlag,
			Verbose:  verboseFlag,
		})
		if err != nil {
			return fmt.Errorf("failed to create generator: %w", err)
		}

		// Afficher le mode
		if dryRunFlag {
			fmt.Println("Dry-run mode: no files will be generated")
		}

		fmt.Printf("Génération de la resource: %s\n", name)

		// Générer la resource complète
		if err := generator.GenerateResource(g, name, "."); err != nil {
			return fmt.Errorf("generation failed: %w", err)
		}

		if !dryRunFlag {
			fmt.Printf("\nResource '%s' générée avec succès\n", name)
			fmt.Printf("\nProchaines étapes:\n")
			fmt.Printf("   1. Éditer les fichiers générés pour personnaliser\n")
			fmt.Printf("   2. Exécuter: go generate ./internal/ent\n")
			fmt.Printf("   3. Exécuter: sublimego generate (auto-discovery)\n")
		}

		return nil
	},
}

// MAKE:MIGRATION - Génère une migration de base de données

var makeMigrationCmd = &cobra.Command{
	Use:     "migration [name]",
	Aliases: []string{"m"},
	Short:   "Generate a new database migration",
	Long: `Generate a versioned SQL migration file.

Example: sublimego make:migration add_status_to_users`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string

		if len(args) == 0 {
			fmt.Print("Nom de la migration: ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			name = strings.TrimSpace(input)
		} else {
			name = args[0]
		}

		if name == "" {
			return fmt.Errorf("migration name is required")
		}

		// Générer la migration
		if err := generator.GenerateMigration(name, "."); err != nil {
			return fmt.Errorf("failed to generate migration: %w", err)
		}

		fmt.Printf("Migration créée: %s\n", name)
		fmt.Printf("Éditez le fichier pour ajouter vos changements SQL\n")

		return nil
	},
}

// MAKE:SEEDER - Génère un seeder de base de données

var makeSeederCmd = &cobra.Command{
	Use:     "seeder [name]",
	Aliases: []string{"s"},
	Short:   "Generate a new database seeder",
	Long: `Generate a seeder file for populating initial data.

Example: sublimego make:seeder UserSeeder`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string

		if len(args) == 0 {
			fmt.Print("Nom du seeder: ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			name = strings.TrimSpace(input)
		} else {
			name = args[0]
		}

		if name == "" {
			return fmt.Errorf("seeder name is required")
		}

		// Générer le seeder
		if err := generator.GenerateSeeder(name, "."); err != nil {
			return fmt.Errorf("failed to generate seeder: %w", err)
		}

		fmt.Printf("Seeder créé: %s\n", name)

		return nil
	},
}

// Fonctions utilitaires supprimées - maintenant dans pkg/generator

func init() {
	// Ajouter les flags pour make:resource
	makeResourceCmd.Flags().BoolVar(&forceFlag, "force", false, "Overwrite existing files")
	makeResourceCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Preview without creating files")
	makeResourceCmd.Flags().StringSliceVar(&skipFlag, "skip", []string{}, "Skip specific files (resource,schema,table,form)")
	makeResourceCmd.Flags().BoolVar(&noBackupFlag, "no-backup", false, "Disable automatic backups")
	makeResourceCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Show detailed output")

	makeCmd.AddCommand(makeResourceCmd)
	makeCmd.AddCommand(makeMigrationCmd)
	makeCmd.AddCommand(makeSeederCmd)
}
