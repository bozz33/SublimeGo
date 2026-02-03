package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	configFormat string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Affiche la configuration finale chargée en mémoire",
	Long: `Affiche la configuration finale chargée en mémoire.

Cette commande affiche le résultat fusionné de :
- Valeurs par défaut (config/defaults.yaml)
- Fichier de configuration (config.yaml)
- Variables d'environnement (.env)
- Flags de ligne de commande

Utile pour déboguer les problèmes de configuration et vérifier
les surcharges appliquées.

Formats de sortie supportés :
- YAML (par défaut) : Format lisible et structuré
- JSON : Format pour intégration avec d'autres outils`,
	Example: `  # Afficher la configuration en YAML
  sublimego config

  # Afficher la configuration en JSON
  sublimego config --format json
  sublimego config -f json

  # Vérifier une valeur spécifique (avec jq)
  sublimego config -f json | jq '.database.driver'

  # Comparer avec les valeurs par défaut
  diff <(cat config/defaults.yaml) <(sublimego config)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		fmt.Println("Configuration SublimeGo")
		fmt.Println("=======================")
		fmt.Println()

		var output []byte
		var err error

		switch configFormat {
		case "json":
			output, err = json.MarshalIndent(cfg, "", "  ")
		case "yaml":
			output, err = yaml.Marshal(cfg)
		default:
			return fmt.Errorf("unsupported format: %s (use json or yaml)", configFormat)
		}

		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		fmt.Println(string(output))

		return nil
	},
}

func init() {
	configCmd.Flags().StringVarP(&configFormat, "format", "f", "yaml", "output format (json or yaml)")
}
