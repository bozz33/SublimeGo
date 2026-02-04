package config

import (
	"fmt"
	"time"
)

// Config is the main configuration structure.
// It follows the 12-Factor App pattern and is immutable after loading.
type Config struct {
	Environment string         `mapstructure:"environment" validate:"required,oneof=development staging production"`
	App         AppConfig      `mapstructure:"app" validate:"required"`
	Server      ServerConfig   `mapstructure:"server" validate:"required"`
	Database    DatabaseConfig `mapstructure:"database" validate:"required"`
	Engine      EngineConfig   `mapstructure:"engine" validate:"required"`
	Logging     LoggingConfig  `mapstructure:"logging" validate:"required"`
	Security    SecurityConfig `mapstructure:"security" validate:"required"`
	Features    FeaturesConfig `mapstructure:"features"`
}

// AppConfig holds application metadata.
type AppConfig struct {
	Name    string `mapstructure:"name" validate:"required"`
	Version string `mapstructure:"version" validate:"required"`
	Debug   bool   `mapstructure:"debug"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host                    string        `mapstructure:"host" validate:"required"`
	Port                    int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	ReadTimeout             time.Duration `mapstructure:"read_timeout" validate:"required"`
	WriteTimeout            time.Duration `mapstructure:"write_timeout" validate:"required"`
	IdleTimeout             time.Duration `mapstructure:"idle_timeout" validate:"required"`
	GracefulShutdownTimeout time.Duration `mapstructure:"graceful_shutdown_timeout" validate:"required"`
	MaxHeaderBytes          int           `mapstructure:"max_header_bytes" validate:"required,min=1024"`
	TLS                     *TLSConfig    `mapstructure:"tls"`
}

// TLSConfig holds TLS/HTTPS configuration.
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file" validate:"required_if=Enabled true"`
	KeyFile  string `mapstructure:"key_file" validate:"required_if=Enabled true"`
}

// Address returns the full server address (host:port).
func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver" validate:"required,oneof=sqlite sqlite3 postgres mysql"`
	URL             string        `mapstructure:"url" validate:"required"`
	MaxOpenConns    int           `mapstructure:"max_open_conns" validate:"required,min=1"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" validate:"required,min=1"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" validate:"required"`
	AutoMigrate     bool          `mapstructure:"auto_migrate"`
	LogLevel        string        `mapstructure:"log_level" validate:"required,oneof=debug info warn error silent"`
}

// EngineConfig holds SublimeGo admin panel settings.
type EngineConfig struct {
	BasePath            string `mapstructure:"base_path" validate:"required"`
	BrandName           string `mapstructure:"brand_name" validate:"required"`
	AssetsPath          string `mapstructure:"assets_path" validate:"required"`
	DefaultPageSize     int    `mapstructure:"default_page_size" validate:"required,min=1,max=100"`
	MaxPageSize         int    `mapstructure:"max_page_size" validate:"required,min=1,max=1000"`
	EnableAutoDiscovery bool   `mapstructure:"enable_auto_discovery"`
	ResourcesPath       string `mapstructure:"resources_path"`
}

// LoggingConfig holds logging settings.
type LoggingConfig struct {
	Level            string `mapstructure:"level" validate:"required,oneof=debug info warn error"`
	Format           string `mapstructure:"format" validate:"required,oneof=json text"`
	Output           string `mapstructure:"output" validate:"required,oneof=stdout stderr file"`
	FilePath         string `mapstructure:"file_path" validate:"required_if=Output file"`
	MaxSize          int    `mapstructure:"max_size" validate:"min=1"`
	MaxBackups       int    `mapstructure:"max_backups" validate:"min=0"`
	MaxAge           int    `mapstructure:"max_age" validate:"min=0"`
	EnableCaller     bool   `mapstructure:"enable_caller"`
	EnableStacktrace bool   `mapstructure:"enable_stacktrace"`
}

// SecurityConfig holds security settings.
type SecurityConfig struct {
	EnableCSRF      bool `mapstructure:"enable_csrf"`
	CSRFTokenLength int  `mapstructure:"csrf_token_length" validate:"min=16"`
	EnableCORS      bool `mapstructure:"enable_cors"`

	AllowedOrigins    []string      `mapstructure:"allowed_origins"`
	AllowedMethods    []string      `mapstructure:"allowed_methods"`
	RateLimitEnabled  bool          `mapstructure:"rate_limit_enabled"`
	RateLimitRequests int           `mapstructure:"rate_limit_requests" validate:"min=1"`
	RateLimitWindow   time.Duration `mapstructure:"rate_limit_window" validate:"required"`
	SecretKey         string        `mapstructure:"secret_key" validate:"required,min=32"`
}

// FeaturesConfig holds feature flags.
type FeaturesConfig struct {
	EnableHotReload bool `mapstructure:"enable_hot_reload"`
	EnableMetrics   bool `mapstructure:"enable_metrics"`
	EnableProfiling bool `mapstructure:"enable_profiling"`
	EnableSwagger   bool `mapstructure:"enable_swagger"`
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsStaging returns true if running in staging mode.
func (c *Config) IsStaging() bool {
	return c.Environment == "staging"
}

// Validate checks configuration consistency across all sections.
func (c *Config) Validate() error {
	// Hot reload only allowed in development
	if c.Features.EnableHotReload && c.IsProduction() {
		return fmt.Errorf("hot reload cannot be enabled in production environment")
	}

	// Debug mode only allowed in development
	if c.App.Debug && c.IsProduction() {
		return fmt.Errorf("debug mode cannot be enabled in production environment")
	}

	// Auto-migrate only allowed in development
	if c.Database.AutoMigrate && c.IsProduction() {
		return fmt.Errorf("auto-migrate cannot be enabled in production environment")
	}

	// Ensure MaxIdleConns doesn't exceed MaxOpenConns
	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return fmt.Errorf("database max_idle_conns (%d) cannot be greater than max_open_conns (%d)",
			c.Database.MaxIdleConns, c.Database.MaxOpenConns)
	}

	// Ensure DefaultPageSize doesn't exceed MaxPageSize
	if c.Engine.DefaultPageSize > c.Engine.MaxPageSize {
		return fmt.Errorf("engine default_page_size (%d) cannot be greater than max_page_size (%d)",
			c.Engine.DefaultPageSize, c.Engine.MaxPageSize)
	}

	return nil
}

// String returns a string representation of the config (without secrets).
func (c *Config) String() string {
	return fmt.Sprintf("Config{Environment=%s, App=%s, Server=%s:%d, Database=%s}",
		c.Environment, c.App.Name, c.Server.Host, c.Server.Port, c.Database.Driver)
}

// ServerAddress returns the full server address (host:port).
func (c *Config) ServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
