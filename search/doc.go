// Package search provides global search functionality for SublimeGo.
//
// Global search allows searching across all registered resources from a single
// search input. Resources implement the Searchable interface to participate
// in global search.
//
// Example usage:
//
//	// Register a searchable resource
//	search.Register(
//		search.NewSearchable("Users").
//			SetIcon("users").
//			SetPriority(1).
//			SetFields("name", "email").
//			WithSearcher(func(ctx context.Context, query string, limit int) ([]search.Result, error) {
//				users, err := db.User.Query().
//					Where(user.Or(
//						user.NameContains(query),
//						user.EmailContains(query),
//					)).
//					Limit(limit).
//					All(ctx)
//				if err != nil {
//					return nil, err
//				}
//
//				results := make([]search.Result, len(users))
//				for i, u := range users {
//					results[i] = search.Result{
//						ID:           fmt.Sprintf("%d", u.ID),
//						Title:        u.Name,
//						Subtitle:     u.Email,
//						URL:          fmt.Sprintf("/users/%d", u.ID),
//						Icon:         "user",
//						ResourceType: "Users",
//						Score:        search.CalculateScore(query, u.Name),
//					}
//				}
//				return results, nil
//			}),
//	)
//
//	// Perform a global search
//	results, err := search.QuickSearch(ctx, "john")
package search
