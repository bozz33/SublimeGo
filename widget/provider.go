package widget

import (
	"context"
	"sort"
	"sync"
)

// Provider is the interface for declarative dashboard widget providers.
type Provider interface {
	// GetWidgets returns widgets for the dashboard.
	GetWidgets(ctx context.Context) []Widget
	// GetPriority returns the display priority (lower = first).
	GetPriority() int
	// IsEnabled returns whether this provider is enabled.
	IsEnabled(ctx context.Context) bool
	// GetID returns a unique identifier for this provider.
	GetID() string
}

// BaseProvider provides default implementations for Provider.
type BaseProvider struct {
	id       string
	priority int
	enabled  bool
	widgets  func(ctx context.Context) []Widget
}

// NewProvider creates a new widget provider.
func NewProvider(id string) *BaseProvider {
	return &BaseProvider{
		id:       id,
		priority: 100,
		enabled:  true,
	}
}

func (p *BaseProvider) GetID() string                       { return p.id }
func (p *BaseProvider) GetPriority() int                    { return p.priority }
func (p *BaseProvider) IsEnabled(ctx context.Context) bool  { return p.enabled }

func (p *BaseProvider) GetWidgets(ctx context.Context) []Widget {
	if p.widgets != nil {
		return p.widgets(ctx)
	}
	return []Widget{}
}

// SetPriority sets the display priority.
func (p *BaseProvider) SetPriority(priority int) *BaseProvider {
	p.priority = priority
	return p
}

// SetEnabled sets whether the provider is enabled.
func (p *BaseProvider) SetEnabled(enabled bool) *BaseProvider {
	p.enabled = enabled
	return p
}

// WithWidgets sets the widget generator function.
func (p *BaseProvider) WithWidgets(fn func(ctx context.Context) []Widget) *BaseProvider {
	p.widgets = fn
	return p
}

// ProviderRegistry manages widget providers globally.
type ProviderRegistry struct {
	mu        sync.RWMutex
	providers []Provider
}

var globalRegistry = &ProviderRegistry{
	providers: make([]Provider, 0),
}

// Register registers a widget provider globally.
func Register(p Provider) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.providers = append(globalRegistry.providers, p)
}

// GetProviders returns all registered providers sorted by priority.
func GetProviders() []Provider {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	
	sorted := make([]Provider, len(globalRegistry.providers))
	copy(sorted, globalRegistry.providers)
	
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GetPriority() < sorted[j].GetPriority()
	})
	
	return sorted
}

// GetAllWidgets returns all widgets from all enabled providers.
func GetAllWidgets(ctx context.Context) []Widget {
	providers := GetProviders()
	var allWidgets []Widget
	
	for _, p := range providers {
		if p.IsEnabled(ctx) {
			allWidgets = append(allWidgets, p.GetWidgets(ctx)...)
		}
	}
	
	return allWidgets
}

// Unregister removes a provider by ID.
func Unregister(id string) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	
	filtered := make([]Provider, 0)
	for _, p := range globalRegistry.providers {
		if p.GetID() != id {
			filtered = append(filtered, p)
		}
	}
	globalRegistry.providers = filtered
}

// Clear removes all providers.
func Clear() {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.providers = make([]Provider, 0)
}

// ProviderCount returns the number of registered providers.
func ProviderCount() int {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	return len(globalRegistry.providers)
}
