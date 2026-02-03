package commands

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/alexedwards/scs/v2"
	"github.com/bozz33/SublimeGo/internal/ent"
	"github.com/bozz33/SublimeGo/pkg/auth"
	"github.com/bozz33/SublimeGo/pkg/engine"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:     "serve",
	Aliases: []string{"s"},
	Short:   "Start the HTTP server",
	Long: `Start the SublimeGo HTTP server with graceful shutdown.
	
This command:
- Loads the configuration
- Connects to the database
- Initializes the admin panel
- Starts the HTTP server
- Handles graceful shutdown on SIGINT/SIGTERM`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()

		fmt.Printf("Démarrage de SublimeGo %s\n", version)
		fmt.Printf("Environnement: %s\n", cfg.Environment)
		fmt.Printf("Mode debug: %v\n", cfg.App.Debug)

		// Connexion à la base de données
		fmt.Printf("Connexion à la base de données (%s)...\n", cfg.Database.Driver)

		// Ajouter le pragma foreign_keys pour SQLite si absent
		dbURL := cfg.Database.URL
		if (cfg.Database.Driver == "sqlite3" || cfg.Database.Driver == "sqlite") && !strings.Contains(dbURL, "_fk=") {
			if strings.Contains(dbURL, "?") {
				dbURL += "&_fk=1"
			} else {
				dbURL += "?_fk=1"
			}
		}
		fmt.Printf("URL: %s\n", dbURL)

		// Ouvrir la connexion SQL directement pour pouvoir exécuter PRAGMA
		db, err := sql.Open(cfg.Database.Driver, dbURL)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}

		// Activer les foreign keys pour SQLite
		if cfg.Database.Driver == "sqlite3" || cfg.Database.Driver == "sqlite" {
			if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
				db.Close()
				return fmt.Errorf("failed to enable foreign keys: %w", err)
			}
		}

		// Créer le driver Ent à partir de la connexion SQL
		// Utiliser "sqlite3" comme dialecte pour Ent (requis par Ent)
		dialectName := cfg.Database.Driver
		if dialectName == "sqlite" {
			dialectName = "sqlite3"
		}
		drv := entsql.OpenDB(dialectName, db)
		client := ent.NewClient(ent.Driver(drv))
		defer client.Close()
		fmt.Println("✓ Connexion DB réussie")

		// Exécution des migrations si activées
		if cfg.Database.AutoMigrate {
			fmt.Println("Exécution des migrations automatiques...")
			ctx := context.Background()
			if err := client.Schema.Create(ctx); err != nil {
				return fmt.Errorf("failed to run migrations: %w", err)
			}
			fmt.Println("Migrations terminées")
		}

		// Initialisation de la session et de l'authentification
		fmt.Println("Initialisation de l'authentification...")
		sessionManager := scs.New()
		authManager := auth.NewManager(sessionManager)

		// Initialisation du panneau d'administration
		fmt.Println("Initialisation du panneau d'administration...")
		panel := engine.NewPanel("admin").
			SetPath(cfg.Engine.BasePath).
			SetBrandName(cfg.Engine.BrandName).
			SetDatabase(client).
			SetAuthManager(authManager).
			SetSession(sessionManager)

		// Création du serveur HTTP
		addr := cfg.ServerAddress()
		server := &http.Server{
			Addr:              addr,
			Handler:           panel.Router(),
			ReadTimeout:       cfg.Server.ReadTimeout,
			WriteTimeout:      cfg.Server.WriteTimeout,
			IdleTimeout:       cfg.Server.IdleTimeout,
			MaxHeaderBytes:    cfg.Server.MaxHeaderBytes,
			ReadHeaderTimeout: 10 * time.Second,
		}

		// Démarrage du serveur en goroutine
		serverErrors := make(chan error, 1)
		go func() {
			fmt.Printf("\nServeur démarré sur http://%s\n", addr)
			fmt.Printf("Panneau d'administration: http://%s%s\n", addr, cfg.Engine.BasePath)
			fmt.Println("\nAppuyez sur Ctrl+C pour arrêter")
			serverErrors <- server.ListenAndServe()
		}()

		// Attente du signal d'interruption
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		select {
		case err := <-serverErrors:
			return fmt.Errorf("server error: %w", err)
		case sig := <-shutdown:
			fmt.Printf("\n\nSignal reçu: %v\n", sig)
			fmt.Println("Arrêt gracieux en cours...")

			// Arrêt gracieux du serveur
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulShutdownTimeout)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				server.Close()
				return fmt.Errorf("failed to shutdown gracefully: %w", err)
			}

			fmt.Println("Serveur arrêté avec succès")
		}

		return nil
	},
}
