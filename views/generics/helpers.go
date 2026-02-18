package generics

import (
	"fmt"
	"strings"

	"github.com/bozz33/sublimego/engine"
)

// getValueStr returns the value as a string
func getValueStr(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// isChecked returns true if the value is a bool true
func isChecked(v any) bool {
	if v == nil {
		return false
	}
	if val, ok := v.(bool); ok {
		return val
	}
	return false
}

// hasValue returns true if the value is not nil
func hasValue(v any) bool { //nolint:unused
	return v != nil
}

// headerActionClass returns Tailwind classes for a header action button by color.
func headerActionClass(color string) string {
	base := "inline-flex items-center gap-1.5 px-3 py-2 text-sm font-medium rounded-xl transition-colors "
	switch color {
	case "danger":
		return base + "text-red-700 bg-red-50 hover:bg-red-100 border border-red-200 dark:text-red-400 dark:bg-red-900/20 dark:border-red-800 dark:hover:bg-red-900/40"
	case "secondary":
		return base + "border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
	default: // "primary"
		return base + "text-white bg-primary-600 hover:bg-primary-700"
	}
}

// bulkActionClass returns Tailwind classes for a bulk action button by color.
func bulkActionClass(color string) string { //nolint:unused
	switch color {
	case "danger":
		return "text-red-700 bg-red-50 hover:bg-red-100 dark:text-red-400 dark:bg-red-900/20 dark:hover:bg-red-900/40"
	case "warning":
		return "text-yellow-700 bg-yellow-50 hover:bg-yellow-100 dark:text-yellow-400 dark:bg-yellow-900/20"
	default:
		return "text-primary-700 bg-primary-50 hover:bg-primary-100 dark:text-primary-400 dark:bg-primary-900/20"
	}
}

// rowIDsJSON returns a JSON array of row IDs for Alpine.js toggleAll.
func rowIDsJSON(rows []engine.Row) string { //nolint:unused
	if len(rows) == 0 {
		return "[]"
	}
	parts := make([]string, len(rows))
	for i, r := range rows {
		parts[i] = fmt.Sprintf("%q", r.ID)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// suppressUnused silences the "declared but not used" error for loop index.
func suppressUnused(_ int) string { return "" } //nolint:unused

// min returns the smaller of two ints (Go 1.20 stdlib min is generic, keep local for compat).
func min(a, b int) int { //nolint:unused
	if a < b {
		return a
	}
	return b
}
