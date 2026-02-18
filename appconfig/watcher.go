package appconfig

import (
	"fmt"
	"log"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Watcher monitors and reloads configuration automatically.
type Watcher struct {
	cfg      *Config
	v        *viper.Viper
	mu       sync.RWMutex // Protection thread-safe
	handlers []ChangeHandler
	active   bool
	stopCh   chan struct{}
}

// ChangeHandler is a function called when the config changes.
type ChangeHandler func(old, new *Config) error

// NewWatcher creates a new Watcher.
func NewWatcher(cfg *Config, v *viper.Viper) *Watcher {
	return &Watcher{
		cfg:      cfg,
		v:        v,
		handlers: make([]ChangeHandler, 0),
		active:   false,
		stopCh:   make(chan struct{}),
	}
}

// OnChange registers a handler called on config change.
func (w *Watcher) OnChange(handler ChangeHandler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handlers = append(w.handlers, handler)
}

// Start begins watching for config changes (dev mode only).
func (w *Watcher) Start() error {
	w.mu.Lock()
	if w.active {
		w.mu.Unlock()
		return fmt.Errorf("watcher already started")
	}
	w.active = true
	w.mu.Unlock()

	w.v.WatchConfig()
	w.v.OnConfigChange(w.handleConfigChange)

	log.Println("[Config] Hot reload enabled (development mode)")
	return nil
}

// Stop stops watching for config changes.
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.active {
		return
	}

	w.active = false
	close(w.stopCh)
	log.Println("[Config] Hot reload stopped")
}

// handleConfigChange is called automatically by Viper on config change.
func (w *Watcher) handleConfigChange(e fsnotify.Event) {
	log.Printf("[Config] Configuration file modified: %s", e.Name)

	w.mu.RLock()
	oldCfg := w.copyConfig(w.cfg)
	w.mu.RUnlock()

	newCfg := &Config{}
	if err := w.v.Unmarshal(newCfg); err != nil {
		log.Printf("[Config] Failed to unmarshal new config: %v", err)
		return
	}

	loader := NewLoader()
	if err := loader.validateTags(newCfg); err != nil {
		log.Printf("[Config] Invalid configuration after reload: %v", err)
		log.Println("[Config] Keeping previous configuration")
		return
	}

	if err := newCfg.Validate(); err != nil {
		log.Printf("[Config] Validation error after reload: %v", err)
		log.Println("[Config] Keeping previous configuration")
		return
	}

	w.mu.Lock()
	w.cfg = newCfg
	w.mu.Unlock()

	log.Println("[Config] Configuration reloaded successfully")

	w.notifyHandlers(oldCfg, newCfg)
}

// notifyHandlers calls all registered handlers.
func (w *Watcher) notifyHandlers(old, new *Config) {
	w.mu.RLock()
	handlers := make([]ChangeHandler, len(w.handlers))
	copy(handlers, w.handlers)
	w.mu.RUnlock()

	for i, handler := range handlers {
		if err := handler(old, new); err != nil {
			log.Printf("[Config] Handler #%d error: %v", i+1, err)
		}
	}
}

// GetConfig returns the current config in a thread-safe manner.
func (w *Watcher) GetConfig() *Config {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.copyConfig(w.cfg)
}

// copyConfig creates a deep copy of the config.
func (w *Watcher) copyConfig(cfg *Config) *Config {
	if cfg == nil {
		return nil
	}

	copied := *cfg

	if cfg.Security.AllowedOrigins != nil {
		copied.Security.AllowedOrigins = make([]string, len(cfg.Security.AllowedOrigins))
		copy(copied.Security.AllowedOrigins, cfg.Security.AllowedOrigins)
	}

	if cfg.Security.AllowedMethods != nil {
		copied.Security.AllowedMethods = make([]string, len(cfg.Security.AllowedMethods))
		copy(copied.Security.AllowedMethods, cfg.Security.AllowedMethods)
	}

	return &copied
}

// IsActive returns true if the watcher is active.
func (w *Watcher) IsActive() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.active
}

// Watch is a helper to enable hot reload in development only.
func Watch(cfg *Config) *Watcher {
	if cfg.Environment != "development" {
		log.Println("[Config] Hot reload disabled (not in development mode)")
		return nil
	}

	if !cfg.Features.EnableHotReload {
		log.Println("[Config] Hot reload disabled (feature flag disabled)")
		return nil
	}

	loader := NewLoader()
	if err := loader.configure(); err != nil {
		log.Printf("[Config] Failed to configure loader for hot reload: %v", err)
		return nil
	}

	watcher := NewWatcher(cfg, loader.GetViper())
	if err := watcher.Start(); err != nil {
		log.Printf("[Config] Failed to start hot reload: %v", err)
		return nil
	}

	return watcher
}

// LogChangeHandler is a handler that logs important changes.
func LogChangeHandler(old, new *Config) error {
	log.Println("[Config] Configuration changes detected:")

	changes := 0

	// Server
	if old.Server.Port != new.Server.Port {
		log.Printf("  • Server port: %d → %d", old.Server.Port, new.Server.Port)
		changes++
	}
	if old.Server.Host != new.Server.Host {
		log.Printf("  • Server host: %s → %s", old.Server.Host, new.Server.Host)
		changes++
	}

	// Database
	if old.Database.Driver != new.Database.Driver {
		log.Printf("  • Database driver: %s → %s", old.Database.Driver, new.Database.Driver)
		changes++
	}
	if old.Database.URL != new.Database.URL {
		log.Println("  • Database URL changed")
		changes++
	}

	// Logging
	if old.Logging.Level != new.Logging.Level {
		log.Printf("  • Log level: %s → %s", old.Logging.Level, new.Logging.Level)
		changes++
	}

	// Engine
	if old.Engine.DefaultPageSize != new.Engine.DefaultPageSize {
		log.Printf("  • Page size: %d → %d", old.Engine.DefaultPageSize, new.Engine.DefaultPageSize)
		changes++
	}

	if changes == 0 {
		log.Println("  • No significant changes detected")
	}

	return nil
}

// RestartRequiredHandler detects changes requiring a restart.
func RestartRequiredHandler(old, new *Config) error {
	restartRequired := false

	// Critical changes requiring a restart
	if old.Server.Port != new.Server.Port {
		log.Println("[Config] Server port changed - RESTART REQUIRED")
		restartRequired = true
	}

	if old.Server.Host != new.Server.Host {
		log.Println("[Config] Server host changed - RESTART REQUIRED")
		restartRequired = true
	}

	if old.Database.Driver != new.Database.Driver {
		log.Println("[Config] Database driver changed - RESTART REQUIRED")
		restartRequired = true
	}

	if old.Database.URL != new.Database.URL {
		log.Println("[Config] Database URL changed - RESTART REQUIRED")
		restartRequired = true
	}

	if restartRequired {
		log.Println("[Config] Please restart the application to apply changes")
	}

	return nil
}
