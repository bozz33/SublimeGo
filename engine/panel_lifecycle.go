package engine

import "fmt"

// BootHook is a function called during panel boot lifecycle.
type BootHook func(p *Panel) error

// WithBeforeBoot registers a hook called before Router() boots the panel.
// Multiple hooks are called in registration order.
// Return a non-nil error to abort boot.
func (p *Panel) WithBeforeBoot(fn BootHook) *Panel {
	p.beforeBootHooks = append(p.beforeBootHooks, fn)
	return p
}

// WithAfterBoot registers a hook called after Router() has finished booting.
// Multiple hooks are called in registration order.
func (p *Panel) WithAfterBoot(fn BootHook) *Panel {
	p.afterBootHooks = append(p.afterBootHooks, fn)
	return p
}

// runBeforeBoot executes all registered BeforeBoot hooks.
func (p *Panel) runBeforeBoot() error {
	for i, fn := range p.beforeBootHooks {
		if err := fn(p); err != nil {
			return fmt.Errorf("panel %s: before_boot hook %d: %w", p.ID, i, err)
		}
	}
	return nil
}

// runAfterBoot executes all registered AfterBoot hooks.
func (p *Panel) runAfterBoot() error {
	for i, fn := range p.afterBootHooks {
		if err := fn(p); err != nil {
			return fmt.Errorf("panel %s: after_boot hook %d: %w", p.ID, i, err)
		}
	}
	return nil
}
