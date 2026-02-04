// Package config provides configuration management with YAML support.
//
// It handles loading, parsing, and hot-reloading of configuration files.
// The package supports environment-specific configurations and provides
// a watcher for automatic reload on file changes.
//
// Features:
//   - YAML configuration files
//   - Environment-specific overrides
//   - Hot-reload with file watching
//   - Type-safe configuration structs
//   - Default values
//
// Basic usage:
//
//	cfg, err := config.Load("config.yaml")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Access configuration
//	port := cfg.Server.Port
//	dbURL := cfg.Database.URL
//
//	// Watch for changes
//	config.Watch("config.yaml", func(newCfg *config.Config) {
//		// Handle config update
//	})
package config
