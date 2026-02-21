package generics

import (
	"fmt"

	"github.com/bozz33/sublimego/actions"
)

// actionModalDispatch builds the Alpine.js $dispatch call string for a confirmation modal.
func actionModalDispatch(a *actions.Action, item any) string {
	url := ""
	if a.UrlResolver != nil {
		url = a.UrlResolver(item)
	}
	method := a.Method
	if method == "" {
		method = "POST"
	}
	confirmLabel := a.ConfirmLabel
	if confirmLabel == "" {
		confirmLabel = "Confirm"
	}
	cancelLabel := a.CancelLabel
	if cancelLabel == "" {
		cancelLabel = "Cancel"
	}
	title := a.ModalTitle
	desc := a.ModalDescription
	color := a.Color
	if color == "" {
		color = "red"
	}
	return fmt.Sprintf(
		`$dispatch('open-action-modal', { url: '%s', method: '%s', title: '%s', desc: '%s', confirmLabel: '%s', cancelLabel: '%s', color: '%s' })`,
		url, method, title, desc, confirmLabel, cancelLabel, color,
	)
}

// actionButtonClass returns Tailwind classes for a full action button by color.
func actionButtonClass(color string) string {
	base := "inline-flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-xl transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 "
	switch color {
	case "primary":
		return base + "text-white bg-primary-600 hover:bg-primary-700 focus:ring-primary-500"
	case "danger", "red":
		return base + "text-white bg-red-600 hover:bg-red-700 focus:ring-red-500"
	case "success", "green":
		return base + "text-white bg-green-600 hover:bg-green-700 focus:ring-green-500"
	case "warning", "yellow":
		return base + "text-white bg-yellow-500 hover:bg-yellow-600 focus:ring-yellow-400"
	case "info", "blue":
		return base + "text-white bg-blue-600 hover:bg-blue-700 focus:ring-blue-500"
	case "secondary":
		return base + "text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700 focus:ring-gray-500"
	default:
		return base + "text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700 focus:ring-gray-500"
	}
}

// actionIconClass returns Tailwind classes for a compact icon-only action button.
func actionIconClass(color string) string {
	base := "p-1.5 rounded-lg transition-colors "
	switch color {
	case "primary":
		return base + "text-gray-500 hover:text-primary-600 hover:bg-primary-50 dark:hover:bg-primary-900/20"
	case "danger", "red":
		return base + "text-gray-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20"
	case "success", "green":
		return base + "text-gray-500 hover:text-green-600 hover:bg-green-50 dark:hover:bg-green-900/20"
	case "warning", "yellow":
		return base + "text-gray-500 hover:text-yellow-600 hover:bg-yellow-50 dark:hover:bg-yellow-900/20"
	case "info", "blue":
		return base + "text-gray-500 hover:text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/20"
	default:
		return base + "text-gray-500 hover:text-gray-700 hover:bg-gray-100 dark:hover:bg-gray-700"
	}
}

// actionIcon returns the Material Icon name for an action, with sensible defaults.
func actionIcon(a *actions.Action) string {
	if a.Icon != "" {
		return a.Icon
	}
	switch a.Name {
	case "edit":
		return "edit"
	case "delete", "force-delete":
		return "delete"
	case "view":
		return "visibility"
	case "restore":
		return "restore"
	case "export":
		return "download"
	case "import":
		return "upload"
	case "create":
		return "add"
	default:
		return "more_horiz"
	}
}
