// Package atoms provides reusable UI components built with Templ.
//
// These are the smallest building blocks of the UI, including buttons,
// inputs, cards, and other common elements. All components follow
// a consistent design system with Tailwind CSS.
//
// Features:
//   - Buttons (primary, secondary, danger, etc.)
//   - Form inputs (text, select, checkbox, radio)
//   - Cards and containers
//   - Badges and indicators
//   - Modals and popovers
//   - Tables and pagination
//   - Spinners and loading states
//
// Basic usage:
//
//	// Button
//	@atoms.Button(atoms.ButtonProps{
//		Text:    "Save",
//		Variant: "primary",
//		Type:    "submit",
//	})
//
//	// Input
//	@atoms.Input(atoms.InputProps{
//		Name:        "email",
//		Type:        "email",
//		Placeholder: "Enter email",
//	})
//
//	// Card
//	@atoms.Card() {
//		<p>Card content</p>
//	}
package atoms
