// Package widget provides dashboard widget components.
//
// It includes stats cards and chart widgets for building admin dashboards.
// Widgets are rendered as Templ components and support various configurations.
//
// Features:
//   - Stats cards with icons and trends
//   - Chart widgets (line, bar, pie) using ApexCharts
//   - Customizable colors and sizes
//   - Trend indicators (up/down)
//   - Responsive design
//
// Basic usage:
//
//	// Stats card
//	stats := widget.NewStats().
//		SetTitle("Total Users").
//		SetValue("1,234").
//		SetIcon("users").
//		SetTrend("+12%", "up")
//
//	// Chart widget
//	chart := widget.NewChart().
//		SetType("line").
//		SetTitle("Revenue").
//		SetData(revenueData).
//		SetHeight(300)
//
//	// Render widgets
//	stats.Render(ctx)
//	chart.Render(ctx)
package widget
