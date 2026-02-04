package widget

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

// Widget is a generic interface.
type Widget interface {
	GetType() string
}

func (s *StatsWidget) GetType() string { return "stats" }
