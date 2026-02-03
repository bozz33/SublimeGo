package commands

import (
	"fmt"

	"github.com/bozz33/SublimeGo/pkg/config"
	"github.com/spf13/cobra"
)

// Informations de version (injectées lors du build)

var (
	version   = "dev"
	gitCommit = "unknown"
	buildDate = "unknown"
)

// SetVersionInfo injecte les informations de build depuis main
func SetVersionInfo(v, commit, date string) {
	version = v
	gitCommit = commit
	buildDate = date
}

// État global partagé entre les commandes

var (
	// Configuration chargée (disponible pour toutes les commandes)
	cfg *config.Config

	// Fichier de configuration personnalisé
	cfgFile string

	// Active les logs détaillés
	verbose bool
)

// Commande racine (infrastructure)

var rootCmd = &cobra.Command{
	Use:   "sublimego",
	Short: "SublimeGo - The Go Admin Framework",
	Long: `SublimeGo is a production-grade admin framework for Go.
	
It provides automatic CRUD generation, beautiful UI components,
and a powerful CLI for rapid application development.

Inspired by Laravel Filament, built for Go developers.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Ignore le chargement de config pour certaines commandes
		if cmd.Name() == "version" || cmd.Name() == "help" {
			return nil
		}

		// Charge la configuration
		opts := []config.Option{}
		if cfgFile != "" {
			opts = append(opts, config.WithConfigPath(cfgFile))
		}

		var err error
		cfg, err = config.Load(opts...)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		if verbose {
			fmt.Printf("Configuration chargée (environnement: %s)\n", cfg.Environment)
		}

		return nil
	},
}

// Execute exécute la commande racine
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Flags globaux
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./config/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")

	// Ajoute toutes les sous-commandes
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(makeCmd)
	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(routesCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(resourceCmd)
}

// GetConfig retourne la configuration chargée
func GetConfig() *config.Config {
	return cfg
}
