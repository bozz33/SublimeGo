package widget

import "github.com/a-h/templ"

// Widget is the interface all dashboard widgets must implement.
type Widget interface {
	GetType() string
	Render() templ.Component
}

// Stat represents a single statistic card.
type Stat struct {
	Label       string
	Value       string
	Description string
	Icon        string
	Color       string
	Chart       []int
	Increase    bool
}

// StatsWidget is a container for multiple stats.
type StatsWidget struct {
	Stats []Stat
}

// NewStats creates a new statistics widget.
func NewStats(stats ...Stat) *StatsWidget {
	return &StatsWidget{Stats: stats}
}

func (s *StatsWidget) GetType() string { return "stats" }

// renderFunc is set by the views/widgets package to avoid import cycles.
var statsRenderFunc func(*StatsWidget) templ.Component

// SetStatsRenderer registers the render function (called from views/widgets init).
func SetStatsRenderer(fn func(*StatsWidget) templ.Component) {
	statsRenderFunc = fn
}

func (s *StatsWidget) Render() templ.Component {
	if statsRenderFunc != nil {
		return statsRenderFunc(s)
	}
	return templ.NopComponent
}
