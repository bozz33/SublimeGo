package form

// Section represents a form section.
type Section struct {
	Heading     string
	Description string
	Components  []Component
	CanCollapse bool
	Collapsed   bool
}

// NewSection creates a new section.
func NewSection(heading string) *Section {
	return &Section{
		Heading:    heading,
		Components: make([]Component, 0),
	}
}

// SetSchema sets the section components.
func (s *Section) SetSchema(components ...Component) *Section {
	s.Components = components
	return s
}

// Desc sets the description.
func (s *Section) Desc(desc string) *Section {
	s.Description = desc
	return s
}

// Collapsible makes the section collapsible.
func (s *Section) Collapsible() *Section {
	s.CanCollapse = true
	return s
}

// IsVisible returns true if the section is visible.
func (s *Section) IsVisible() bool { return true }

// GetComponentType returns the component type.
func (s *Section) ComponentType() string { return "layout_section" }

// Schema returns the section components.
func (s *Section) Schema() []Component { return s.Components }

// Grid represents a column grid.
type Grid struct {
	Columns    int
	Components []Component
}

// NewGrid creates a new grid.
func NewGrid(columns int) *Grid {
	return &Grid{
		Columns:    columns,
		Components: make([]Component, 0),
	}
}

// SetSchema sets the grid components.
func (g *Grid) SetSchema(components ...Component) *Grid {
	g.Components = components
	return g
}

// IsVisible returns true if the grid is visible.
func (g *Grid) IsVisible() bool { return true }

// GetComponentType returns the component type.
func (g *Grid) ComponentType() string { return "layout_grid" }

// Schema returns the grid components.
func (g *Grid) Schema() []Component { return g.Components }

// Tabs represents a tab system.
type Tabs struct {
	TabItems []Tab
}

// Tab represents a tab.
type Tab struct {
	Label      string
	Icon       string
	Components []Component
}

// NewTabs creates a new tab system.
func NewTabs() *Tabs {
	return &Tabs{
		TabItems: make([]Tab, 0),
	}
}

// AddTab adds a tab.
func (t *Tabs) AddTab(label string, components ...Component) *Tabs {
	t.TabItems = append(t.TabItems, Tab{
		Label:      label,
		Components: components,
	})
	return t
}

// IsVisible returns true if the tabs are visible.
func (t *Tabs) IsVisible() bool { return true }

// GetComponentType returns the component type.
func (t *Tabs) ComponentType() string { return "layout_tabs" }

// Schema returns all components from all tabs.
func (t *Tabs) Schema() []Component {
	var all []Component
	for _, tab := range t.TabItems {
		all = append(all, tab.Components...)
	}
	return all
}

// WizardStep represents a single step in a Wizard.
type WizardStep struct {
	Label       string
	Description string
	Icon        string
	Components  []Component
}

// Wizard represents a multi-step form wizard.
type Wizard struct {
	Steps []*WizardStep
}

// NewWizard creates a new Wizard.
func NewWizard() *Wizard {
	return &Wizard{Steps: make([]*WizardStep, 0)}
}

// AddStep adds a step to the wizard.
func (w *Wizard) AddStep(label string, components ...Component) *Wizard {
	w.Steps = append(w.Steps, &WizardStep{
		Label:      label,
		Components: components,
	})
	return w
}

// WithDescription sets the description of the last added step.
func (w *Wizard) WithDescription(desc string) *Wizard {
	if len(w.Steps) > 0 {
		w.Steps[len(w.Steps)-1].Description = desc
	}
	return w
}

// WithIcon sets the icon of the last added step.
func (w *Wizard) WithIcon(icon string) *Wizard {
	if len(w.Steps) > 0 {
		w.Steps[len(w.Steps)-1].Icon = icon
	}
	return w
}

// IsVisible returns true.
func (w *Wizard) IsVisible() bool { return true }

// ComponentType returns the component type.
func (w *Wizard) ComponentType() string { return "layout_wizard" }

// Schema returns all components from all steps.
func (w *Wizard) Schema() []Component {
	var all []Component
	for _, step := range w.Steps {
		all = append(all, step.Components...)
	}
	return all
}

// CalloutColor defines the color/intent of a Callout.
type CalloutColor string

const (
	CalloutInfo    CalloutColor = "info"
	CalloutSuccess CalloutColor = "success"
	CalloutWarning CalloutColor = "warning"
	CalloutDanger  CalloutColor = "danger"
)

// Callout renders an informational banner inside a form.
type Callout struct {
	Heading    string
	Body       string
	Icon       string
	Color      CalloutColor
	Components []Component
}

// NewCallout creates a new Callout.
func NewCallout(heading string) *Callout {
	return &Callout{
		Heading:    heading,
		Color:      CalloutInfo,
		Components: make([]Component, 0),
	}
}

// WithBody sets the callout body text.
func (c *Callout) WithBody(body string) *Callout {
	c.Body = body
	return c
}

// WithIcon sets the callout icon.
func (c *Callout) WithIcon(icon string) *Callout {
	c.Icon = icon
	return c
}

// WithColor sets the callout color.
func (c *Callout) WithColor(color CalloutColor) *Callout {
	c.Color = color
	return c
}

// IsVisible returns true.
func (c *Callout) IsVisible() bool { return true }

// ComponentType returns the component type.
func (c *Callout) ComponentType() string { return "layout_callout" }

// Schema returns the callout's nested components.
func (c *Callout) Schema() []Component { return c.Components }
