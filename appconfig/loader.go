package appconfig

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Loader manages configuration loading and validation.
type Loader struct {
	v         *viper.Viper
	validator *validator.Validate
	options   *LoadOptions
}

// LoadOptions contains loading options.
type LoadOptions struct {
	ConfigPaths       []string
	ConfigName        string
	ConfigType        string
	EnvPrefix         string
	RequireConfigFile bool
}

// NewLoader creates a new Loader with default options.
func NewLoader(opts ...Option) *Loader {
	options := &LoadOptions{
		ConfigPaths: []string{
			".",
			"./config",
			"$HOME/.sublimego",
			"/etc/sublimego",
		},
		ConfigName:        "config",
		ConfigType:        "yaml",
		EnvPrefix:         "SUBLIMEGO",
		RequireConfigFile: false,
	}

	for _, opt := range opts {
		opt(options)
	}

	return &Loader{
		v:         viper.New(),
		validator: validator.New(),
		options:   options,
	}
}

// Load loads and validates the complete configuration.
func (l *Loader) Load() (*Config, error) {
	if err := l.configure(); err != nil {
		return nil, fmt.Errorf("failed to configure loader: %w", err)
	}

	l.setDefaults()

	if err := l.readConfigFile(); err != nil {
		return nil, err
	}

	l.bindEnvironmentVariables()

	cfg := &Config{}
	if err := l.v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := l.validateTags(cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	return cfg, nil
}

// configure initializes Viper with options.
func (l *Loader) configure() error {
	l.v.SetConfigName(l.options.ConfigName)
	l.v.SetConfigType(l.options.ConfigType)

	for _, path := range l.options.ConfigPaths {
		l.v.AddConfigPath(path)
	}

	return nil
}

// setDefaults sets default values.
func (l *Loader) setDefaults() {
	l.v.SetDefault("environment", "development")
	l.v.SetDefault("app.name", "SublimeGo")
	l.v.SetDefault("app.version", "1.0.0")
	l.v.SetDefault("app.debug", false)

	l.v.SetDefault("server.host", "localhost")
	l.v.SetDefault("server.port", 8080)
	l.v.SetDefault("server.read_timeout", 15*time.Second)
	l.v.SetDefault("server.write_timeout", 15*time.Second)
	l.v.SetDefault("server.idle_timeout", 60*time.Second)
	l.v.SetDefault("server.graceful_shutdown_timeout", 30*time.Second)
	l.v.SetDefault("server.max_header_bytes", 1<<20) // 1 MB

	l.v.SetDefault("database.driver", "sqlite")
	l.v.SetDefault("database.url", "file:sublimego.db?cache=shared&_fk=1")
	l.v.SetDefault("database.max_open_conns", 25)
	l.v.SetDefault("database.max_idle_conns", 5)
	l.v.SetDefault("database.conn_max_lifetime", 5*time.Minute)
	l.v.SetDefault("database.auto_migrate", true)
	l.v.SetDefault("database.log_level", "warn")

	l.v.SetDefault("engine.base_path", "/admin")
	l.v.SetDefault("engine.brand_name", "SublimeGo Admin")
	l.v.SetDefault("engine.assets_path", "/assets")
	l.v.SetDefault("engine.default_page_size", 15)
	l.v.SetDefault("engine.max_page_size", 100)
	l.v.SetDefault("engine.enable_auto_discovery", true)
	l.v.SetDefault("engine.resources_path", "internal/resources")

	l.v.SetDefault("logging.level", "info")
	l.v.SetDefault("logging.format", "json")
	l.v.SetDefault("logging.output", "stdout")
	l.v.SetDefault("logging.max_size", 100) // 100 MB
	l.v.SetDefault("logging.max_backups", 3)
	l.v.SetDefault("logging.max_age", 28) // 28 days
	l.v.SetDefault("logging.enable_caller", false)
	l.v.SetDefault("logging.enable_stacktrace", false)

	l.v.SetDefault("security.enable_csrf", true)
	l.v.SetDefault("security.csrf_token_length", 32)
	l.v.SetDefault("security.enable_cors", false)
	l.v.SetDefault("security.allowed_origins", []string{})
	l.v.SetDefault("security.allowed_methods", []string{"GET", "POST", "PUT", "DELETE"})
	l.v.SetDefault("security.rate_limit_enabled", true)
	l.v.SetDefault("security.rate_limit_requests", 100)
	l.v.SetDefault("security.rate_limit_window", 1*time.Minute)
	l.v.SetDefault("security.secret_key", generateSecureSecret())

	l.v.SetDefault("features.enable_hot_reload", false)
	l.v.SetDefault("features.enable_metrics", false)
	l.v.SetDefault("features.enable_profiling", false)
	l.v.SetDefault("features.enable_swagger", false)
}

// readConfigFile reads the configuration file.
func (l *Loader) readConfigFile() error {
	err := l.v.ReadInConfig()

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if l.options.RequireConfigFile {
				return fmt.Errorf("config file required but not found in paths: %v",
					l.options.ConfigPaths)
			}
			return nil
		}
		return fmt.Errorf("error reading config file: %w", err)
	}

	return nil
}

// bindEnvironmentVariables automatically binds environment variables.
func (l *Loader) bindEnvironmentVariables() {
	l.v.AutomaticEnv()
	l.v.SetEnvPrefix(l.options.EnvPrefix)
	l.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	criticalVars := []string{
		"environment",
		"server.port",
		"server.host",
		"database.driver",
		"database.url",
		"security.secret_key",
		"logging.level",
	}

	for _, key := range criticalVars {
		envKey := l.options.EnvPrefix + "_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		_ = l.v.BindEnv(key, envKey)
	}
}

// validateTags validates validation tags.
func (l *Loader) validateTags(cfg *Config) error {
	if err := l.validator.Struct(cfg); err != nil {
		return l.formatValidationError(err)
	}
	return nil
}

// formatValidationError formats validation errors in a readable way.
func (l *Loader) formatValidationError(err error) error {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	var errors []string
	for _, e := range validationErrors {
		field := e.Field()
		tag := e.Tag()
		param := e.Param()

		var msg string
		switch tag {
		case "required":
			msg = fmt.Sprintf("'%s' is required", field)
		case "oneof":
			msg = fmt.Sprintf("'%s' must be one of: %s", field, param)
		case "min":
			msg = fmt.Sprintf("'%s' must be at least %s", field, param)
		case "max":
			msg = fmt.Sprintf("'%s' must be at most %s", field, param)
		case "required_if":
			msg = fmt.Sprintf("'%s' is required when %s", field, param)
		default:
			msg = fmt.Sprintf("'%s' failed validation '%s'", field, tag)
		}

		errors = append(errors, msg)
	}

	return fmt.Errorf("validation failed:\n  - %s", strings.Join(errors, "\n  - "))
}

// generateSecureSecret generates a cryptographically secure secret.
func generateSecureSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "CHANGE-ME-IN-PRODUCTION-MINIMUM-32-CHARS-REQUIRED"
	}
	return hex.EncodeToString(bytes)
}

// Load is a helper to load configuration with default options.
func Load(opts ...Option) (*Config, error) {
	loader := NewLoader(opts...)
	return loader.Load()
}

// MustLoad loads configuration and panics on error.
func MustLoad(opts ...Option) *Config {
	cfg, err := Load(opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}
	return cfg
}

// GetViper returns the Viper instance for hot reload.
func (l *Loader) GetViper() *viper.Viper {
	return l.v
}
