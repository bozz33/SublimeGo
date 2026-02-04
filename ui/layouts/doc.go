// Package layouts provides page layout components for the admin panel.
//
// It includes the base HTML layout, sidebar navigation, topbar, and
// flash message display. Layouts are built with Templ and integrate
// with Alpine.js for interactivity.
//
// Features:
//   - Base HTML layout with Tailwind CSS
//   - Responsive sidebar with collapsible groups
//   - Topbar with user menu and notifications
//   - Flash message display
//   - Dark mode support
//   - Panel configuration
//
// Basic usage:
//
//	// Configure panel
//	layouts.SetPanelConfig(&layouts.PanelConfig{
//		Name:         "My Admin",
//		PrimaryColor: "blue",
//		DarkMode:     true,
//	})
//
//	// Set navigation
//	layouts.SetNavItems([]layouts.NavItem{
//		{Slug: "users", Label: "Users", Icon: "users"},
//		{Slug: "products", Label: "Products", Icon: "shopping-bag"},
//	})
//
//	// Render page
//	layouts.Page("Dashboard", content)
package layouts
