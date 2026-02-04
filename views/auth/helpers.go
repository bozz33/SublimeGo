package auth

// getBasePath returns the base path for authentication URLs
func getBasePath(basePath ...string) string {
	if len(basePath) > 0 && basePath[0] != "" {
		return basePath[0]
	}
	return ""
}
