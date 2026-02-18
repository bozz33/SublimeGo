package plugin

import (
	"fmt"
	"sync"
)

// Plugin is the interface that all SublimeGo plugins must implement.
type Plugin interface {
	// Name returns the unique identifier of the plugin.
	Name() string
	// Boot is called once when the panel starts, after all plugins are registered.
	Boot() error
}

// registry holds all registered plugins.
var (
	mu      sync.RWMutex
	plugins []Plugin
)

// Register adds a plugin to the global registry.
// Typically called from a plugin's init() function.
func Register(p Plugin) {
	mu.Lock()
	defer mu.Unlock()
	plugins = append(plugins, p)
}

// All returns a copy of all registered plugins.
func All() []Plugin {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Plugin, len(plugins))
	copy(out, plugins)
	return out
}

// Boot calls Boot() on every registered plugin in registration order.
// Returns the first error encountered.
func Boot() error {
	mu.RLock()
	list := make([]Plugin, len(plugins))
	copy(list, plugins)
	mu.RUnlock()

	for _, p := range list {
		if err := p.Boot(); err != nil {
			return fmt.Errorf("plugin %q: boot failed: %w", p.Name(), err)
		}
	}
	return nil
}

// Get returns the plugin with the given name, or nil if not found.
func Get(name string) Plugin {
	mu.RLock()
	defer mu.RUnlock()
	for _, p := range plugins {
		if p.Name() == name {
			return p
		}
	}
	return nil
}
