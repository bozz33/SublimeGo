package widgets

import (
	"github.com/a-h/templ"
	"github.com/bozz33/sublimego/widget"
)

func init() {
	widget.SetStatsRenderer(func(w *widget.StatsWidget) templ.Component {
		return Stats(w)
	})
	widget.SetChartRenderer(func(w *widget.ChartWidget) templ.Component {
		return Chart(w)
	})
}
