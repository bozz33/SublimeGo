package main

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/bozz33/sublimeadmin/widget"
)

// blogWidgets builds the dashboard widgets from live blog data: headline stats,
// a posts-by-category chart and a draft/published distribution.
func blogWidgets(ctx context.Context, db *sql.DB) []widget.Widget {
	posts := scalar(ctx, db, `SELECT COUNT(*) FROM posts`)
	published := scalar(ctx, db, `SELECT COUNT(*) FROM posts WHERE status = 'published'`)
	drafts := posts - published
	categories := scalar(ctx, db, `SELECT COUNT(*) FROM categories`)
	comments := scalar(ctx, db, `SELECT COUNT(*) FROM comments`)

	stats := widget.NewStats(
		widget.Stat{Label: "Posts", Value: strconv.Itoa(posts)},
		widget.Stat{Label: "Published", Value: strconv.Itoa(published)},
		widget.Stat{Label: "Categories", Value: strconv.Itoa(categories)},
		widget.Stat{Label: "Comments", Value: strconv.Itoa(comments)},
	)

	// Posts per category (one slice per category for the circular chart).
	byCat := widget.NewChart("posts-by-category", "Posts by category", widget.Pie).
		WithDescription("How posts are distributed across categories")
	catLabels := []string{}
	if rows, err := db.QueryContext(ctx, `
		SELECT c.name, (SELECT COUNT(*) FROM posts p WHERE p.category_id = c.id)
		FROM categories c ORDER BY c.name`); err == nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			var n int
			if rows.Scan(&name, &n) == nil {
				catLabels = append(catLabels, name)
				byCat.AddSeries(name, []int{n})
			}
		}
	}
	byCat.SetLabels(catLabels)

	byStatus := widget.NewChart("posts-by-status", "Posts by status", widget.Donut).
		SetLabels([]string{"Draft", "Published"}).
		AddSeries("Draft", []int{drafts}).
		AddSeries("Published", []int{published}).
		WithDescription("Draft vs published posts")

	return []widget.Widget{stats, byCat, byStatus}
}

// scalar runs a single-value COUNT query, returning 0 on error.
func scalar(ctx context.Context, db *sql.DB, query string) int {
	var n int
	if err := db.QueryRowContext(ctx, query).Scan(&n); err != nil {
		return 0
	}
	return n
}
