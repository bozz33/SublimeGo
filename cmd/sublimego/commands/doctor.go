package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/bozz33/sublimego/internal/ent"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:     "doctor",
	Aliases: []string{"doc"},
	Short:   "Check system health and requirements",
	Long: `Run diagnostic checks on your SublimeGo installation.

This command verifies:
- Go version
- Required tools (templ, air)
- Database connectivity
- Configuration validity
- Environment variables`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Diagnostic SublimeGo")
		fmt.Println("====================")
		fmt.Println()

		allOK := true

		// Vérification de la version Go
		fmt.Printf("Vérification de la version Go... ")
		goVersion := runtime.Version()
		fmt.Printf("%s ", goVersion)
		if strings.HasPrefix(goVersion, "go1.2") || strings.HasPrefix(goVersion, "go1.1") {
			fmt.Println("OK")
		} else {
			fmt.Println("ATTENTION (Recommandé: Go 1.18+)")
			allOK = false
		}

		// Vérification de Templ
		fmt.Printf("Vérification de templ... ")
		if _, err := exec.LookPath("templ"); err == nil {
			templCmd := exec.Command("templ", "version")
			output, _ := templCmd.CombinedOutput()
			fmt.Printf("%s OK\n", strings.TrimSpace(string(output)))
		} else {
			fmt.Println("NON TROUVÉ")
			fmt.Println("   Installer avec: go install github.com/a-h/templ/cmd/templ@latest")
			allOK = false
		}

		// Vérification de Air (optionnel)
		fmt.Printf("Vérification de air (optionnel)... ")
		if _, err := exec.LookPath("air"); err == nil {
			fmt.Println("OK")
		} else {
			fmt.Println("Non trouvé (optionnel pour le hot reload)")
		}

		// Vérification de la connectivité DB
		fmt.Printf("Vérification de la connectivité DB... ")
		cfg := GetConfig()
		if cfg != nil {
			client, err := ent.Open(cfg.Database.Driver, cfg.Database.URL)
			if err != nil {
				fmt.Printf("ÉCHEC: %v\n", err)
				allOK = false
			} else {
				defer client.Close()
				ctx := context.Background()
				if err := client.Schema.Create(ctx); err != nil {
					fmt.Printf("Problème de schéma: %v\n", err)
				} else {
					fmt.Println("OK")
				}
			}
		} else {
			fmt.Println("Ignoré (pas de config)")
		}

		// Vérification du fichier .env
		fmt.Printf("Vérification du fichier .env... ")
		if _, err := os.Stat(".env"); err == nil {
			fmt.Println("OK")
		} else {
			fmt.Println("Non trouvé (valeurs par défaut utilisées)")
		}

		// Vérification du fichier config
		fmt.Printf("Vérification de config.yaml... ")
		if _, err := os.Stat("config/config.yaml"); err == nil {
			fmt.Println("OK")
		} else {
			fmt.Println("Non trouvé (valeurs par défaut utilisées)")
		}

		// Résumé
		fmt.Println("\n====================")
		if allOK {
			fmt.Println("Tous les tests sont passés. Système prêt.")
		} else {
			fmt.Println("Problèmes détectés. Veuillez vérifier ci-dessus.")
		}

		return nil
	},
}
