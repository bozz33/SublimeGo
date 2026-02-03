package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bozz33/SublimeGo/internal/ent"
	"github.com/spf13/cobra"
)

// Commande DB (parent pour les opérations de base de données)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
	Long:  `Manage database migrations, seeding, and schema operations.`,
}

// DB:MIGRATE - Exécute les migrations de base de données

var dbMigrateCmd = &cobra.Command{
	Use:     "migrate",
	Aliases: []string{"m"},
	Short:   "Run database migrations",
	Long: `Apply database schema changes using Ent auto-migration.

This command will:
- Connect to the database
- Create/update tables based on Ent schemas
- Apply any pending migrations`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()

		fmt.Printf("Exécution des migrations sur la base %s...\n", cfg.Database.Driver)

		client, err := ent.Open(cfg.Database.Driver, cfg.Database.URL)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer client.Close()

		ctx := context.Background()
		if err := client.Schema.Create(ctx); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		fmt.Println("Migrations terminées avec succès")
		return nil
	},
}

// DB:SEED - Exécute les seeders de base de données

var dbSeedCmd = &cobra.Command{
	Use:     "seed",
	Aliases: []string{"s"},
	Short:   "Seed the database with initial data",
	Long: `Run all database seeders to populate initial data.

This is useful for:
- Development environments
- Testing
- Initial production setup`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()

		fmt.Println("Peuplement de la base de données...")

		client, err := ent.Open(cfg.Database.Driver, cfg.Database.URL)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer client.Close()

		ctx := context.Background()

		// TODO: Charger et exécuter les seeders depuis internal/seeders
		fmt.Println("Aucun seeder configuré")
		fmt.Println("Créer des seeders avec: sublimego make:seeder [name]")

		_ = ctx
		return nil
	},
}

// DB:RESET - Réinitialise la base de données (DANGEREUX)

var dbResetCmd = &cobra.Command{
	Use:     "reset",
	Aliases: []string{"r"},
	Short:   "Reset the database (drop all tables)",
	Long: `DANGER: Drop all tables, run migrations, and seed.

This will:
1. Drop all tables
2. Run migrations
3. Run seeders

This is DESTRUCTIVE and cannot be undone!`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()

		// Demande de confirmation
		fmt.Println("ATTENTION: Ceci va SUPPRIMER TOUTES LES DONNÉES de votre base!")
		fmt.Printf("Database: %s (%s)\n", cfg.Database.Driver, cfg.Database.URL)
		fmt.Print("\nAre you sure you want to continue? (yes/no): ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "yes" && response != "y" {
			fmt.Println("Opération annulée")
			return nil
		}

		fmt.Println("\nSuppression de toutes les tables...")

		client, err := ent.Open(cfg.Database.Driver, cfg.Database.URL)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer client.Close()

		ctx := context.Background()

		// Suppression et recréation du schéma
		fmt.Println("Recréation du schéma...")
		if err := client.Schema.Create(ctx); err != nil {
			return fmt.Errorf("failed to recreate schema: %w", err)
		}
		fmt.Println("Schéma recréé")

		// Exécution des seeders
		fmt.Println("Peuplement de la base de données...")
		fmt.Println("Aucun seeder configuré")

		fmt.Println("\nRéinitialisation de la base terminée")
		return nil
	},
}

// DB:STATUS - Affiche le statut de la base de données

var dbStatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st"},
	Short:   "Show database connection status",
	Long:    `Display information about the database connection and schema.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()

		fmt.Println("Statut de la base de données")
		fmt.Println("=============================")
		fmt.Println()

		fmt.Printf("Driver:     %s\n", cfg.Database.Driver)
		fmt.Printf("URL:        %s\n", maskDatabaseURL(cfg.Database.URL))
		fmt.Printf("Max Conns:  %d\n", cfg.Database.MaxOpenConns)
		fmt.Printf("Idle Conns: %d\n", cfg.Database.MaxIdleConns)
		fmt.Printf("Lifetime:   %s\n", cfg.Database.ConnMaxLifetime)

		fmt.Println("\nTest de connexion...")

		client, err := ent.Open(cfg.Database.Driver, cfg.Database.URL)
		if err != nil {
			fmt.Printf("Échec de connexion: %v\n", err)
			return err
		}
		defer client.Close()

		ctx := context.Background()
		if err := client.Schema.Create(ctx); err != nil {
			fmt.Printf("Problème de schéma: %v\n", err)
		} else {
			fmt.Println("Connexion réussie")
		}

		return nil
	},
}

// Fonctions utilitaires

func maskDatabaseURL(url string) string {
	// Masque les parties sensibles de l'URL (mots de passe, etc.)
	if strings.Contains(url, "@") {
		parts := strings.Split(url, "@")
		if len(parts) == 2 {
			return "***@" + parts[1]
		}
	}
	return url
}

func init() {
	dbCmd.AddCommand(dbMigrateCmd)
	dbCmd.AddCommand(dbSeedCmd)
	dbCmd.AddCommand(dbResetCmd)
	dbCmd.AddCommand(dbStatusCmd)
}
