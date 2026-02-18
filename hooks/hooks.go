package hooks

import (
	"context"
	"io"
	"sync"

	"github.com/a-h/templ"
)

// Position is a named injection point in the UI.
type Position string

const (
	BeforeContent    Position = "before_content"
	AfterContent     Position = "after_content"
	BeforeFormFields Position = "before_form_fields"
	AfterFormFields  Position = "after_form_fields"
	BeforeTableRows  Position = "before_table_rows"
	AfterTableRows   Position = "after_table_rows"
	BeforeNavigation Position = "before_navigation"
	AfterNavigation  Position = "after_navigation"
	InHead           Position = "in_head"
	InFooter         Position = "in_footer"
)

// RenderFunc is a function that returns a Templ component for a given context.
type RenderFunc func(ctx context.Context) templ.Component

// registry holds all registered render hooks.
var registry struct {
	mu    sync.RWMutex
	hooks map[Position][]RenderFunc
}

func init() {
	registry.hooks = make(map[Position][]RenderFunc)
}

// Register adds a render function at the given position.
// Multiple functions can be registered at the same position; they are
// rendered in registration order.
func Register(pos Position, fn RenderFunc) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.hooks[pos] = append(registry.hooks[pos], fn)
}

// Render returns a Templ component that renders all hooks at the given position.
// Returns nil if no hooks are registered at that position.
func Render(pos Position) templ.Component {
	registry.mu.RLock()
	fns := registry.hooks[pos]
	registry.mu.RUnlock()

	if len(fns) == 0 {
		return nil
	}

	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		for _, fn := range fns {
			c := fn(ctx)
			if c == nil {
				continue
			}
			if err := c.Render(ctx, w); err != nil {
				return err
			}
		}
		return nil
	})
}

// Clear removes all hooks at the given position. Useful in tests.
func Clear(pos Position) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	delete(registry.hooks, pos)
}

// ClearAll removes all registered hooks. Useful in tests.
func ClearAll() {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.hooks = make(map[Position][]RenderFunc)
}

// Has returns true if at least one hook is registered at the given position.
func Has(pos Position) bool {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	return len(registry.hooks[pos]) > 0
}
