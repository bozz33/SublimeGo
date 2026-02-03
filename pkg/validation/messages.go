package validation

// frenchMessages returns validation messages in French
func frenchMessages() map[string]string {
	return map[string]string{
		// Required & Presence
		"required":         "Le champ {field} est obligatoire",
		"required_if":      "Le champ {field} est obligatoire quand {param} est défini",
		"required_unless":  "Le champ {field} est obligatoire sauf si {param} est défini",
		"required_with":    "Le champ {field} est obligatoire avec {param}",
		"required_without": "Le champ {field} est obligatoire sans {param}",

		// String Length
		"min": "Le champ {field} doit contenir au minimum {param} caractères",
		"max": "Le champ {field} doit contenir au maximum {param} caractères",
		"len": "Le champ {field} doit contenir exactement {param} caractères",

		// Numeric Comparison
		"eq":  "Le champ {field} doit être égal à {param}",
		"ne":  "Le champ {field} ne doit pas être égal à {param}",
		"lt":  "Le champ {field} doit être inférieur à {param}",
		"lte": "Le champ {field} doit être inférieur ou égal à {param}",
		"gt":  "Le champ {field} doit être supérieur à {param}",
		"gte": "Le champ {field} doit être supérieur ou égal à {param}",

		// Numeric Types
		"numeric": "Le champ {field} doit être un nombre",
		"number":  "Le champ {field} doit être un nombre",
		"integer": "Le champ {field} doit être un entier",

		// String Format
		"alpha":        "Le champ {field} ne doit contenir que des lettres",
		"alphanum":     "Le champ {field} ne doit contenir que des lettres et chiffres",
		"alphaunicode": "Le champ {field} ne doit contenir que des lettres unicode",
		"ascii":        "Le champ {field} ne doit contenir que des caractères ASCII",
		"lowercase":    "Le champ {field} doit être en minuscules",
		"uppercase":    "Le champ {field} doit être en majuscules",
		"startswith":   "Le champ {field} doit commencer par {param}",
		"endswith":     "Le champ {field} doit se terminer par {param}",
		"contains":     "Le champ {field} doit contenir {param}",
		"excludes":     "Le champ {field} ne doit pas contenir {param}",

		// Email & URL
		"email":    "Le champ {field} doit être une adresse email valide",
		"url":      "Le champ {field} doit être une URL valide",
		"uri":      "Le champ {field} doit être un URI valide",
		"hostname": "Le champ {field} doit être un nom d'hôte valide",
		"fqdn":     "Le champ {field} doit être un nom de domaine complet",

		// IP
		"ip":   "Le champ {field} doit être une adresse IP valide",
		"ipv4": "Le champ {field} doit être une adresse IPv4 valide",
		"ipv6": "Le champ {field} doit être une adresse IPv6 valide",
		"mac":  "Le champ {field} doit être une adresse MAC valide",

		// UUID
		"uuid":  "Le champ {field} doit être un UUID valide",
		"uuid3": "Le champ {field} doit être un UUID v3 valide",
		"uuid4": "Le champ {field} doit être un UUID v4 valide",
		"uuid5": "Le champ {field} doit être un UUID v5 valide",

		// Date & Time
		"datetime": "Le champ {field} doit être une date/heure valide",

		// Boolean
		"boolean": "Le champ {field} doit être un booléen (true/false)",

		// JSON
		"json": "Le champ {field} doit être un JSON valide",

		// Base64
		"base64":    "Le champ {field} doit être encodé en base64",
		"base64url": "Le champ {field} doit être encodé en base64url",

		// Field Comparison
		"eqfield":  "Le champ {field} doit être égal à {param}",
		"nefield":  "Le champ {field} ne doit pas être égal à {param}",
		"gtfield":  "Le champ {field} doit être supérieur à {param}",
		"gtefield": "Le champ {field} doit être supérieur ou égal à {param}",
		"ltfield":  "Le champ {field} doit être inférieur à {param}",
		"ltefield": "Le champ {field} doit être inférieur ou égal à {param}",

		// Array/Slice
		"unique": "Le champ {field} ne doit pas contenir de doublons",
		"dive":   "Chaque élément de {field}",

		// Custom French Validators
		"phone_fr":        "Le champ {field} doit être un numéro de téléphone français valide",
		"postal_code_fr":  "Le champ {field} doit être un code postal français valide (ex: 75001)",
		"slug":            "Le champ {field} doit être un slug valide (ex: mon-article-123)",
		"siret":           "Le champ {field} doit être un numéro SIRET valide (14 chiffres)",
		"siren":           "Le champ {field} doit être un numéro SIREN valide (9 chiffres)",
		"strong_password": "Le champ {field} doit contenir au moins 8 caractères avec majuscule, minuscule et chiffre",
	}
}
