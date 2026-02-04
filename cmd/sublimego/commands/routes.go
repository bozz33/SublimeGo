package commands

import (
	"fmt"
	"strings"

	"github.com/bozz33/sublimego/internal/ent"
	"github.com/bozz33/sublimego/engine"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var routesCmd = &cobra.Command{
	Use:   "route:list",
	Short: "Affiche toutes les routes HTTP enregistrées dans l'application",
	Long: `Affiche un tableau détaillé de toutes les routes HTTP enregistrées.

Pour chaque route, affiche :
- Méthode HTTP (GET, POST, PUT, DELETE, etc.)
- Chemin de la route (/admin/users, /api/posts, etc.)
- Handler associé (nom de la fonction)
- Middleware appliqués (auth, cors, etc.)

Utile pour :
- Déboguer les problèmes de routage
- Documenter l'API
- Vérifier les permissions et middleware
- Auditer la sécurité des endpoints`,
	Example: `  # Afficher toutes les routes
  sublimego route:list
  sublimego routes

  # Filtrer les routes par méthode (avec grep)
  sublimego route:list | grep GET

  # Filtrer les routes par chemin
  sublimego route:list | grep /admin

  # Compter le nombre de routes
  sublimego route:list | wc -l

  # Exporter vers un fichier
  sublimego route:list > routes.txt`,
	Aliases: []string{"routes"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()

		fmt.Println("Routes SublimeGo")
		fmt.Println("================")
		fmt.Println()

		// Création d'un client temporaire pour l'initialisation des resources
		client, err := ent.Open(cfg.Database.Driver, cfg.Database.URL)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer client.Close()

		// Initialisation du panel pour récupérer les resources
		panel := engine.NewPanel("admin").
			SetPath(cfg.Engine.BasePath)

		// Affichage des routes de base
		fmt.Println("Routes de base:")
		fmt.Println("─────────────────────────────────────────────")
		printRoute("GET", "/", "Dashboard")
		fmt.Println()

		// Affichage des routes des resources
		fmt.Println("Routes des resources:")
		fmt.Println("─────────────────────────────────────────────")

		routes := lo.FlatMap(panel.Resources, func(res engine.Resource, _ int) []RouteInfo {
			slug := res.Slug()
			basePath := cfg.Engine.BasePath

			return []RouteInfo{
				{Method: "GET", Path: basePath + "/" + slug, Description: fmt.Sprintf("%s - List", res.PluralLabel())},
				{Method: "GET", Path: basePath + "/" + slug + "/create", Description: fmt.Sprintf("%s - Create Form", res.Label())},
				{Method: "POST", Path: basePath + "/" + slug, Description: fmt.Sprintf("%s - Store", res.Label())},
				{Method: "GET", Path: basePath + "/" + slug + "/{id}/edit", Description: fmt.Sprintf("%s - Edit Form", res.Label())},
				{Method: "POST", Path: basePath + "/" + slug + "/{id}", Description: fmt.Sprintf("%s - Update", res.Label())},
				{Method: "DELETE", Path: basePath + "/" + slug + "/{id}", Description: fmt.Sprintf("%s - Delete", res.Label())},
			}
		})

		for _, route := range routes {
			printRoute(route.Method, route.Path, route.Description)
		}

		fmt.Printf("\nTotal routes: %d\n", len(routes)+1)

		return nil
	},
}

type RouteInfo struct {
	Method      string
	Path        string
	Description string
}

func printRoute(method, path, description string) {
	methodColor := getMethodColor(method)
	methodPadded := padRight(method, 7)
	pathPadded := padRight(path, 40)

	fmt.Printf("%s%s\x1b[0m  %s  %s\n", methodColor, methodPadded, pathPadded, description)
}

func getMethodColor(method string) string {
	switch method {
	case "GET":
		return "\x1b[32m" // Green
	case "POST":
		return "\x1b[33m" // Yellow
	case "PUT", "PATCH":
		return "\x1b[34m" // Blue
	case "DELETE":
		return "\x1b[31m" // Red
	default:
		return "\x1b[37m" // White
	}
}

func padRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-len(s))
}
