// Package icons provides embedded SVG icons for complete independence.
//
// It replaces external icon libraries with local SVG icons, ensuring
// the framework works without external dependencies. Icons are organized
// by category and can be customized with CSS classes and sizes.
//
// Features:
//   - Navigation icons (home, dashboard, users, settings)
//   - Action icons (plus, edit, delete, search)
//   - Status icons (check, warning, error, info)
//   - UI icons (menu, close, chevrons)
//   - Business icons (food, heart, location)
//   - Variable size spinners
//
// Basic usage:
//
//	// Get icon by name
//	icon := icons.Get("users")
//
//	// Render as string
//	html := icon.String()
//
//	// With custom class
//	html := icon.WithClass("w-6 h-6 text-blue-500")
//
//	// With specific size
//	html := icon.WithSize("8") // w-8 h-8
//
//	// In templates
//	@icons.IconComponent("users")
//	@icons.IconWithSize("settings", "6")
package icons
