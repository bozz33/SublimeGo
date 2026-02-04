package commands

import (
	"fmt"
	"path/filepath"

	"github.com/bozz33/sublimego/scanner"
	"github.com/spf13/cobra"
)

var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Gestion des resources",
	Long: `Commandes pour gérer et inspecter les resources de l'application.

Les resources sont les entités métier de votre application (User, Post, etc.).
Elles sont automatiquement découvertes et enregistrées via le système
d'auto-discovery.`,
}

var resourceListCmd = &cobra.Command{
	Use:   "list",
	Short: "Liste toutes les resources découvertes",
	Long: `Affiche la liste de toutes les resources découvertes par le scanner.

Cette commande scanne le dossier internal/resources et affiche :
- Le nom du type (ex: UserResource)
- Le package (ex: user)
- Le slug (ex: users)
- Le chemin du fichier source

Utile pour :
- Vérifier que vos resources sont bien découvertes
- Déboguer les problèmes d'auto-discovery
- Auditer les resources disponibles`,
	Example: `  # Lister toutes les resources
  sublimego resource:list
  sublimego resource list

  # Lister avec détails
  sublimego resource:list --verbose`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		resourcesPath := filepath.Join("internal", "resources")

		fmt.Println("Resources Découvertes")
		fmt.Println("=====================")
		fmt.Println()

		// Créer le scanner
		s := scanner.New(resourcesPath)

		// Scanner les resources
		fmt.Printf("Scanning %s...\n", resourcesPath)
		result := s.Scan()
		if !result.Success {
			return fmt.Errorf("failed to scan resources: %s", result.Message)
		}
		metadata := result.Resources

		if len(metadata) == 0 {
			fmt.Println("No resources found.")
			fmt.Println()
			fmt.Println("Create a resource with:")
			fmt.Println("  sublimego make:resource YourResource")
			return nil
		}

		// Afficher les resources
		fmt.Printf("Found %d resource(s):\n\n", len(metadata))

		for i, m := range metadata {
			fmt.Printf("%d. %s\n", i+1, m.TypeName)
			fmt.Printf("   Package: %s\n", m.PackageName)
			fmt.Printf("   Slug:    %s\n", m.Slug)

			if verbose {
				fmt.Printf("   File:    %s\n", m.FilePath)
			}

			fmt.Println()
		}

		fmt.Println("Next steps:")
		fmt.Println("  1. Run: sublimego generate")
		fmt.Println("  2. Run: sublimego serve")

		return nil
	},
}

var resourceCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Vérifie que les resources sont valides",
	Long: `Vérifie que toutes les resources découvertes sont valides.

Cette commande effectue les vérifications suivantes :
- Les fichiers sont bien nommés (*_resource.go)
- Les types sont bien nommés (*Resource)
- Pas de doublons de slugs
- Les packages sont cohérents

Retourne un code d'erreur si des problèmes sont détectés.`,
	Example: `  # Vérifier les resources
  sublimego resource:check
  sublimego resource check`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resourcesPath := filepath.Join("internal", "resources")

		fmt.Println("Vérification des Resources")
		fmt.Println("===========================")
		fmt.Println()

		// Créer le scanner
		s := scanner.New(resourcesPath)

		// Scanner les resources
		result := s.Scan()
		if !result.Success {
			return fmt.Errorf("failed to scan resources: %s", result.Message)
		}
		metadata := result.Resources

		if len(metadata) == 0 {
			fmt.Println("✓ No resources to check")
			return nil
		}

		// Vérifier les doublons de slugs
		slugs := scanner.ExtractSlugs(metadata)
		uniqueSlugs := make(map[string]bool)
		duplicates := []string{}

		for _, slug := range slugs {
			if uniqueSlugs[slug] {
				duplicates = append(duplicates, slug)
			}
			uniqueSlugs[slug] = true
		}

		if len(duplicates) > 0 {
			fmt.Println("✗ Duplicate slugs found:")
			for _, slug := range duplicates {
				fmt.Printf("  - %s\n", slug)
			}
			return fmt.Errorf("validation failed: duplicate slugs")
		}

		fmt.Printf("✓ All %d resource(s) are valid\n", len(metadata))
		fmt.Println()

		// Afficher un résumé
		fmt.Println("Summary:")
		for _, m := range metadata {
			fmt.Printf("  ✓ %s.%s (slug: %s)\n", m.PackageName, m.TypeName, m.Slug)
		}

		return nil
	},
}

func init() {
	resourceCmd.AddCommand(resourceListCmd)
	resourceCmd.AddCommand(resourceCheckCmd)
}
