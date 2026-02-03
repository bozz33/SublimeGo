package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:     "build",
	Aliases: []string{"b"},
	Short:   "Build the application for production",
	Long: `Build the SublimeGo application for production.

This command runs a complete build pipeline:
1. Generate Templ templates (templ generate)
2. Build Tailwind CSS (if configured)
3. Compile Go binary (go build -o bin/sublimego)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Build de SublimeGo pour la production...")

		// Génération des templates Templ
		fmt.Println("\nÉtape 1/3: Génération des templates Templ...")
		templCmd := exec.Command("templ", "generate")
		templCmd.Stdout = os.Stdout
		templCmd.Stderr = os.Stderr
		if err := templCmd.Run(); err != nil {
			return fmt.Errorf("failed to generate templ templates: %w", err)
		}
		fmt.Println("Templates Templ générés")

		// Build CSS (optionnel)
		fmt.Println("\nÉtape 2/3: Build CSS...")
		fmt.Println("Ignoré (Tailwind non configuré)")

		// Compilation du binaire Go
		fmt.Println("\nÉtape 3/3: Compilation du binaire Go...")

		// Création du répertoire bin si nécessaire
		binDir := "bin"
		if err := os.MkdirAll(binDir, 0755); err != nil {
			return fmt.Errorf("failed to create bin directory: %w", err)
		}

		// Build avec informations de version
		outputPath := filepath.Join(binDir, "sublimego")
		ldflags := fmt.Sprintf("-X main.Version=%s -X main.GitCommit=%s -X main.BuildDate=%s",
			version, gitCommit, buildDate)

		buildCmd := exec.Command("go", "build",
			"-ldflags", ldflags,
			"-o", outputPath,
			"./cmd/sublimego")
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr

		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("failed to build binary: %w", err)
		}

		fmt.Printf("\nBuild terminé\n")
		fmt.Printf("Binaire: %s\n", outputPath)
		fmt.Printf("Exécuter avec: ./%s serve\n", outputPath)

		return nil
	},
}
