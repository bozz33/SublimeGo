package engine

import (
	"fmt"
	"net/http"
	"sync"
)

// Registry manages multiple Panel instances, each mounted at a distinct path.
// Useful for applications that expose several admin panels (e.g. /admin, /vendor, /api-admin).
type Registry struct {
	mu     sync.RWMutex
	panels map[string]*Panel // keyed by Panel.ID
}

// globalRegistry is the default singleton registry.
var globalRegistry = &Registry{
	panels: make(map[string]*Panel),
}

// Register adds a panel to the global registry.
// Panics if a panel with the same ID is already registered.
func Register(p *Panel) {
	globalRegistry.Register(p)
}

// Get returns a panel from the global registry by ID, or nil if not found.
func Get(id string) *Panel {
	return globalRegistry.Get(id)
}

// All returns all panels registered in the global registry.
func All() []*Panel {
	return globalRegistry.All()
}

// Register adds a panel to the registry.
// Panics if a panel with the same ID is already registered.
func (r *Registry) Register(p *Panel) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.panels[p.ID]; exists {
		panic(fmt.Sprintf("engine: panel %q already registered", p.ID))
	}
	r.panels[p.ID] = p
}

// Get returns a panel by ID, or nil if not found.
func (r *Registry) Get(id string) *Panel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.panels[id]
}

// All returns all registered panels.
func (r *Registry) All() []*Panel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*Panel, 0, len(r.panels))
	for _, p := range r.panels {
		result = append(result, p)
	}
	return result
}

// MountAll mounts all registered panels onto the given ServeMux.
// Each panel is mounted at its configured Path (e.g. "/admin").
// Panics if a panel has no Path set.
func (r *Registry) MountAll(mux *http.ServeMux) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.panels {
		if p.Path == "" {
			panic(fmt.Sprintf("engine: panel %q has no Path set — call WithPath() before MountAll()", p.ID))
		}
		handler := p.Router()
		mux.Handle(p.Path+"/", http.StripPrefix(p.Path, handler))
		mux.Handle(p.Path, http.StripPrefix(p.Path, handler))
	}
}

// MountAll is a convenience wrapper on the global registry.
func MountAll(mux *http.ServeMux) {
	globalRegistry.MountAll(mux)
}
