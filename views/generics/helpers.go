package generics

import "fmt"

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
func hasValue(v any) bool {
	return v != nil
}
