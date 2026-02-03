package config

// Option is a function that modifies LoadOptions.
type Option func(*LoadOptions)

// WithConfigPath adds an additional search path.
func WithConfigPath(path string) Option {
	return func(opts *LoadOptions) {
		opts.ConfigPaths = append(opts.ConfigPaths, path)
	}
}

// WithConfigPaths replaces all search paths.
func WithConfigPaths(paths []string) Option {
	return func(opts *LoadOptions) {
		opts.ConfigPaths = paths
	}
}

// PrependConfigPath adds a path at the beginning of the list (high priority).
func PrependConfigPath(path string) Option {
	return func(opts *LoadOptions) {
		opts.ConfigPaths = append([]string{path}, opts.ConfigPaths...)
	}
}

// WithConfigName sets the config file name (without extension).
func WithConfigName(name string) Option {
	return func(opts *LoadOptions) {
		opts.ConfigName = name
	}
}

// WithConfigType sets the config file type.
func WithConfigType(typ string) Option {
	return func(opts *LoadOptions) {
		opts.ConfigType = typ
	}
}

// WithConfigFile sets an exact config file path.
func WithConfigFile(file string) Option {
	return func(opts *LoadOptions) {
		opts.ConfigName = file
		opts.ConfigPaths = []string{"."} // Viper requires at least one path
		opts.RequireConfigFile = true    // If an exact file is specified, it must exist
	}
}

// WithEnvPrefix sets the environment variable prefix.
func WithEnvPrefix(prefix string) Option {
	return func(opts *LoadOptions) {
		opts.EnvPrefix = prefix
	}
}

// RequireConfigFile indicates that a config file is required.
func RequireConfigFile() Option {
	return func(opts *LoadOptions) {
		opts.RequireConfigFile = true
	}
}

// OptionalConfigFile indicates that a config file is optional.
func OptionalConfigFile() Option {
	return func(opts *LoadOptions) {
		opts.RequireConfigFile = false
	}
}

// WithDevelopmentDefaults applies recommended options for development.
func WithDevelopmentDefaults() Option {
	return func(opts *LoadOptions) {
		opts.ConfigPaths = []string{
			".",
			"./config",
		}
		opts.ConfigName = "config"
		opts.ConfigType = "yaml"
		opts.RequireConfigFile = false
	}
}

// WithProductionDefaults applies recommended options for production.
func WithProductionDefaults() Option {
	return func(opts *LoadOptions) {
		opts.ConfigPaths = []string{
			"/etc/sublimego",
			"./config",
		}
		opts.ConfigName = "config"
		opts.ConfigType = "yaml"
		opts.RequireConfigFile = true
	}
}

// WithTestingDefaults applies recommended options for testing.
func WithTestingDefaults() Option {
	return func(opts *LoadOptions) {
		opts.ConfigPaths = []string{"."}
		opts.ConfigName = "config"
		opts.ConfigType = "yaml"
		opts.RequireConfigFile = false
	}
}

// WithDockerDefaults applies recommended options for Docker/Kubernetes.
func WithDockerDefaults() Option {
	return func(opts *LoadOptions) {
		opts.ConfigPaths = []string{}
		opts.RequireConfigFile = false
		opts.EnvPrefix = "SUBLIMEGO"
	}
}
