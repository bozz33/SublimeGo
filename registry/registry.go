package registry

import (
	"fmt"
	"sync"

	"github.com/bozz33/sublimego/engine"
	"github.com/samber/lo"
)

// Registry manages resource registration and discovery.
type Registry struct {
	resources map[string]engine.Resource
	mu        sync.RWMutex
}

// New creates a new Registry instance.
func New() *Registry {
	return &Registry{
		resources: make(map[string]engine.Resource),
	}
}

// Register registers a resource in the registry.
func (r *Registry) Register(resource engine.Resource) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	slug := resource.Slug()
	if _, exists := r.resources[slug]; exists {
		return fmt.Errorf("resource '%s' already registered", slug)
	}

	r.resources[slug] = resource
	return nil
}

// RegisterMany registers multiple resources at once.
func (r *Registry) RegisterMany(resources ...engine.Resource) error {
	for _, resource := range resources {
		if err := r.Register(resource); err != nil {
			return err
		}
	}
	return nil
}

// Get retrieves a resource by its name.
func (r *Registry) Get(name string) (engine.Resource, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	resource, exists := r.resources[name]
	return resource, exists
}

// Has checks if a resource exists.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.resources[name]
	return exists
}

// All returns all registered resources.
func (r *Registry) All() []engine.Resource {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return lo.Values(r.resources)
}

// Names returns the names of all resources.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return lo.Keys(r.resources)
}

// Count returns the number of registered resources.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.resources)
}

// Filter returns resources that match the predicate.
func (r *Registry) Filter(predicate func(engine.Resource) bool) []engine.Resource {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return lo.Filter(r.All(), func(res engine.Resource, _ int) bool {
		return predicate(res)
	})
}

// GroupByCategory groups resources by category.
func (r *Registry) GroupByCategory() map[string][]engine.Resource {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return lo.GroupBy(r.All(), func(res engine.Resource) string {
		if meta, ok := res.(interface{ Category() string }); ok {
			return meta.Category()
		}
		return "General"
	})
}

// Clear empties the registry (useful for tests).
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.resources = make(map[string]engine.Resource)
}

// Global registry instance
var global = New()

// Global returns the global registry instance.
func Global() *Registry {
	return global
}

// Register registers a resource in the global registry.
func Register(resource engine.Resource) error {
	return global.Register(resource)
}

// RegisterMany registers multiple resources in the global registry.
func RegisterMany(resources ...engine.Resource) error {
	return global.RegisterMany(resources...)
}

// Get retrieves a resource from the global registry.
func Get(name string) (engine.Resource, bool) {
	return global.Get(name)
}

// Has checks if a resource exists in the global registry.
func Has(name string) bool {
	return global.Has(name)
}

// All returns all resources from the global registry.
func All() []engine.Resource {
	return global.All()
}

// Names returns the names of all resources from the global registry.
func Names() []string {
	return global.Names()
}

// Count returns the number of resources in the global registry.
func Count() int {
	return global.Count()
}
