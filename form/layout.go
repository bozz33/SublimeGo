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

// Schema sets the section components.
func (s *Section) Schema(components ...Component) *Section {
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
func (s *Section) GetComponentType() string { return "layout_section" }

// GetSchema returns the section components.
func (s *Section) GetSchema() []Component { return s.Components }

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

// Schema sets the grid components.
func (g *Grid) Schema(components ...Component) *Grid {
	g.Components = components
	return g
}

// IsVisible returns true if the grid is visible.
func (g *Grid) IsVisible() bool { return true }

// GetComponentType returns the component type.
func (g *Grid) GetComponentType() string { return "layout_grid" }

// GetSchema returns the grid components.
func (g *Grid) GetSchema() []Component { return g.Components }

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
func (t *Tabs) GetComponentType() string { return "layout_tabs" }

// GetSchema returns all components from all tabs.
func (t *Tabs) GetSchema() []Component {
	var all []Component
	for _, tab := range t.TabItems {
		all = append(all, tab.Components...)
	}
	return all
}
