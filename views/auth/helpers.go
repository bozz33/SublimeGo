package auth

// getBasePath retourne le chemin de base pour les URLs d'authentification
func getBasePath(basePath ...string) string {
	if len(basePath) > 0 && basePath[0] != "" {
		return basePath[0]
	}
	return ""
}
