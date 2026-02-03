// Package components provides composite UI components.
//
// These are higher-level components built from atoms, including
// action buttons, modals, and data tables. They integrate with
// the framework's action and table systems.
//
// Features:
//   - Action buttons with confirmation
//   - Modal dialogs
//   - Data tables with HTMX
//   - Form components
//
// Basic usage:
//
//	// Action button
//	@components.ActionBtn(action, item)
//
//	// Confirmation modal
//	@components.Modal(components.ModalProps{
//		Title:   "Confirm Delete",
//		Content: "Are you sure?",
//	})
//
//	// Data table
//	@components.Table(tableData)
package components
