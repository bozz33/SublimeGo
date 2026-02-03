package commands

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bozz33/SublimeGo/internal/ent"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise le projet (Env, DB, Admin, Panels)",
	Long: `Bootstrap complet de l'application SublimeGo :
1. Configuration de l'environnement (.env)
2. Generation du Schema User Systeme (IsSystem)
3. Generation du code ORM (Ent)
4. Migration de la Base de Donnees
5. Configuration Multi-Panel
6. Creation de l'Administrateur Systeme`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Demarrage de l'initialisation de SublimeGo...")

		if err := setupEnv(); err != nil {
			return err
		}

		if err := setupConfig(); err != nil {
			return err
		}

		if err := generateSystemSchema(); err != nil {
			return err
		}

		fmt.Println("Generation du code Ent...")
		if err := runGoGenerate(); err != nil {
			return fmt.Errorf("echec de go generate: %w", err)
		}

		fmt.Println("Migration de la base de donnees...")
		if err := runMigration(); err != nil {
			return fmt.Errorf("echec de la migration: %w", err)
		}

		if err := createSystemAdmin(); err != nil {
			return err
		}

		fmt.Println("\nInitialisation terminee avec succes !")
		fmt.Println("Lancez 'sublimego serve' pour demarrer.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func setupEnv() error {
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		fmt.Println("Creation du fichier .env...")
		content := "APP_ENV=local\nAPP_KEY=" + generateSecureKey() + "\nDB_DRIVER=sqlite3\nDB_URL=file:sublimego.db?cache=shared&_fk=1\n"
		return os.WriteFile(".env", []byte(content), 0644)
	}
	fmt.Println("Fichier .env existant detecte.")
	return nil
}

func setupConfig() error {
	configPath := filepath.Join("config", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("Generation de la configuration Multi-Panel...")
		os.MkdirAll("config", 0755)

		content := `app:
  name: "SublimeGo App"
  debug: true
  port: 8080

# Configuration Multi-Panel
panels:
  - id: "admin"
    path: "/admin"
    label: "Administration"
    color: "primary"
    middleware: ["auth", "system_guard"]

  - id: "app"
    path: "/app"
    label: "Espace Client"
    color: "blue"
    middleware: ["auth"]
`
		return os.WriteFile(configPath, []byte(content), 0644)
	}
	return nil
}

func generateSystemSchema() error {
	schemaPath := filepath.Join("internal", "ent", "schema", "user.go")
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		fmt.Println("Generation du Schema User Systeme...")
		os.MkdirAll(filepath.Dir(schemaPath), 0755)

		content := `package schema

import (
	"time"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// User schema
type User struct {
	ent.Schema
}

// Fields du User
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("email").Unique(),
		field.String("password").Sensitive(),
		field.String("role").Default("user"),
		
		// Protection systeme
		field.Bool("is_system").Default(false).Comment("True si cet user ne peut pas etre supprime"),
		
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

// Edges du User
func (User) Edges() []ent.Edge {
	return nil
}
`
		return os.WriteFile(schemaPath, []byte(content), 0644)
	}
	fmt.Println("Schema User existant detecte.")
	return nil
}

func runGoGenerate() error {
	cmd := exec.Command("go", "generate", "./internal/ent")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runMigration() error {
	client, err := ent.Open("sqlite3", "file:sublimego.db?cache=shared&_fk=1")
	if err != nil {
		return fmt.Errorf("failed opening connection: %w", err)
	}
	defer client.Close()

	if err := client.Schema.Create(context.Background()); err != nil {
		return fmt.Errorf("failed creating schema: %w", err)
	}

	return nil
}

func createSystemAdmin() error {
	fmt.Println("\nConfiguration de l'Administrateur Systeme")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Email [admin@sublimego.dev]: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)
	if email == "" {
		email = "admin@sublimego.dev"
	}

	fmt.Print("Mot de passe [password]: ")
	pass, _ := reader.ReadString('\n')
	pass = strings.TrimSpace(pass)
	if pass == "" {
		pass = "password"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	client, err := ent.Open("sqlite3", "file:sublimego.db?cache=shared&_fk=1")
	if err != nil {
		return fmt.Errorf("failed opening connection: %w", err)
	}
	defer client.Close()

	_, err = client.User.Create().
		SetName("System Admin").
		SetEmail(email).
		SetPassword(string(hash)).
		SetRole("admin").
		SetIsSystem(true).
		Save(context.Background())

	if err != nil {
		fmt.Printf("Attention: L'admin existe deja ou erreur: %v\n", err)
	} else {
		fmt.Println("Admin cree avec succes")
	}

	return nil
}

func generateSecureKey() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "insecure-default-key-please-change"
	}
	return hex.EncodeToString(bytes)
}
