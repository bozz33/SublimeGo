package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/bozz33/SublimeGo/internal/ent"
	"github.com/bozz33/SublimeGo/views/dashboard"
	"github.com/joho/godotenv"

	// Imports pour faire le "Pont" Ent <-> ModernC
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	_ "modernc.org/sqlite"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Info: Pas de fichier .env trouvé.")
	}

	// Config
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbDriver := os.Getenv("DB_DRIVER")
	if dbDriver == "" {
		dbDriver = "sqlite"
	}

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		dbUrl = "file:dev.db?cache=shared&_fk=1"
	}

	// ---------------------------------------------------------
	// 🔌 CONNEXION AVANCÉE (Le Pont)
	// ---------------------------------------------------------

	// 1. On ouvre une connexion SQL standard avec le driver "sqlite"
	db, err := sql.Open(dbDriver, dbUrl)
	if err != nil {
		log.Fatalf("❌ Erreur ouverture SQL: %v", err)
	}

	// 2. On crée un "Driver Ent" à partir de cette connexion
	// On force le dialecte à "sqlite3" (dialect.SQLite) pour qu'Ent génère le bon SQL
	drv := entsql.OpenDB(dialect.SQLite, db)

	// 3. On initialise le client Ent avec ce driver
	client := ent.NewClient(ent.Driver(drv))
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Erreur fermeture client DB: %v", err)
		}
	}()

	// ---------------------------------------------------------

	// Migration Automatique
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("❌ Erreur migration DB: %v", err)
	}
	fmt.Printf("✅ Base de données connectée (%s via ModernC)\n", dbDriver)

	// Démarrage
	if err := run(port); err != nil {
		fmt.Fprintf(os.Stderr, "erreur serveur: %v\n", err)
		os.Exit(1)
	}
}

func run(port string) error {
	mux := http.NewServeMux()
	filesDir := http.Dir("./pkg/ui/assets")
	fileServer := http.FileServer(filesDir)
	mux.Handle("/assets/", http.StripPrefix("/assets/", fileServer))

	homeComponent := dashboard.Index()
	mux.Handle("/", templ.Handler(homeComponent))

	fmt.Printf("🚀 SublimeGo Engine: Serveur démarré sur http://localhost:%s\n", port)
	return http.ListenAndServe(":"+port, mux)
}
