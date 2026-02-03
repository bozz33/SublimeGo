package providers

import (
	"context"

	"github.com/bozz33/SublimeGo/internal/ent"
	"github.com/bozz33/SublimeGo/pkg/widget"
)

// GetDashboardStats genere les widgets pour la page d'accueil
// Par defaut, le dashboard est vide. Les developpeurs peuvent ajouter leurs propres widgets ici.
func GetDashboardStats(ctx context.Context, client *ent.Client) []widget.Widget {
	// Dashboard vide par defaut
	// Pour ajouter des widgets, decommentez et personnalisez le code ci-dessous:

	/*
		var widgets []widget.Widget

		userCount, err := client.User.Query().Count(ctx)
		if err != nil {
			userCount = 0
		}

		stats := widget.NewStats(
			widget.Stat{
				Label:       "Utilisateurs Totaux",
				Value:       fmt.Sprintf("%d", userCount),
				Description: "+12% ce mois-ci",
				Icon:        "users",
				Color:       "primary",
				Increase:    true,
				Chart:       []int{10, 25, 40, 30, 45, 50, 70, 85},
			},
		)
		widgets = append(widgets, stats)

		return widgets
	*/

	return []widget.Widget{}
}
