package providers

import (
	"context"

	"github.com/bozz33/sublimego/internal/ent"
	"github.com/bozz33/sublimego/widget"
)

// GetDashboardStats generates widgets for the dashboard page.
// By default, the dashboard is empty. Developers can add their own widgets here.
func GetDashboardStats(ctx context.Context, client *ent.Client) []widget.Widget {
	// Empty dashboard by default.
	// To add widgets, uncomment and customize the code below:

	/*
		var widgets []widget.Widget

		userCount, err := client.User.Query().Count(ctx)
		if err != nil {
			userCount = 0
		}

		stats := widget.NewStats(
			widget.Stat{
				Label:       "Total Users",
				Value:       fmt.Sprintf("%d", userCount),
				Description: "+12% this month",
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
