package generics

import "fmt"

// getValueStr retourne la valeur sous forme de string
func getValueStr(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// isChecked retourne true si la valeur est un bool true
func isChecked(v any) bool {
	if v == nil {
		return false
	}
	if val, ok := v.(bool); ok {
		return val
	}
	return false
}

// hasValue retourne true si la valeur n'est pas nil
func hasValue(v any) bool {
	return v != nil
}
