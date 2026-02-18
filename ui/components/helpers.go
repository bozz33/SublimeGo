package components

// getBadgeClasses returns the CSS classes for a badge based on its color
func getBadgeClasses(color string) string {
	classes := "px-2.5 py-0.5 text-xs font-medium rounded-full "
	switch color {
	case "primary":
		return classes + "bg-primary-100 text-primary-800 dark:bg-primary-900 dark:text-primary-300"
	case "danger":
		return classes + "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300"
	case "success":
		return classes + "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300"
	case "warning":
		return classes + "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300"
	default:
		return classes + "bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300"
	}
}

// getActionClasses returns the CSS classes for an action based on its color
func getActionClasses(color string) string {
	base := "inline-flex items-center justify-center w-8 h-8 rounded-lg transition-colors "
	switch color {
	case "primary":
		return base + "text-primary-600 hover:bg-primary-100 dark:text-primary-400 dark:hover:bg-primary-900"
	case "danger":
		return base + "text-red-600 hover:bg-red-100 dark:text-red-400 dark:hover:bg-red-900"
	case "secondary":
		return base + "text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700"
	default:
		return base + "text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700"
	}
}

// getBulkActionClasses returns CSS classes for a bulk action button.
func getBulkActionClasses(color string) string {
	base := "inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors "
	switch color {
	case "danger":
		return base + "bg-red-50 text-red-700 hover:bg-red-100 dark:bg-red-900/20 dark:text-red-400 dark:hover:bg-red-900/40"
	case "primary":
		return base + "bg-primary-50 text-primary-700 hover:bg-primary-100 dark:bg-primary-900/20 dark:text-primary-400 dark:hover:bg-primary-900/40"
	default:
		return base + "bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
	}
}
