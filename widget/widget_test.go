package widget

import (
	"testing"
)

func TestNewStats(t *testing.T) {
	stats := NewStats(
		Stat{Label: "Users", Value: "100"},
		Stat{Label: "Revenue", Value: "$1000"},
	)
	
	if stats == nil {
		t.Error("Expected stats widget to be created")
	}
	if len(stats.Stats) != 2 {
		t.Errorf("Expected 2 stats, got %d", len(stats.Stats))
	}
}

func TestStatsWidgetType(t *testing.T) {
	stats := NewStats()
	
	if stats.GetType() != "stats" {
		t.Errorf("Expected type 'stats', got '%s'", stats.GetType())
	}
}

func TestStat(t *testing.T) {
	stat := Stat{
		Label:       "Total Users",
		Value:       "1,234",
		Description: "+12% this month",
		Icon:        "users",
		Color:       "primary",
		Increase:    true,
		Chart:       []int{10, 20, 30, 40, 50},
	}
	
	if stat.Label != "Total Users" {
		t.Errorf("Expected label 'Total Users', got '%s'", stat.Label)
	}
	if stat.Value != "1,234" {
		t.Errorf("Expected value '1,234', got '%s'", stat.Value)
	}
	if !stat.Increase {
		t.Error("Expected Increase to be true")
	}
	if len(stat.Chart) != 5 {
		t.Errorf("Expected 5 chart points, got %d", len(stat.Chart))
	}
}

func TestNewChart(t *testing.T) {
	chart := NewChart("test-chart", "Test Chart", Line)
	
	if chart == nil {
		t.Error("Expected chart widget to be created")
	}
	if chart.ID != "test-chart" {
		t.Errorf("Expected ID 'test-chart', got '%s'", chart.ID)
	}
	if chart.Label != "Test Chart" {
		t.Errorf("Expected label 'Test Chart', got '%s'", chart.Label)
	}
	if chart.Type != Line {
		t.Errorf("Expected type Line, got '%s'", chart.Type)
	}
	if chart.Height != "300" {
		t.Errorf("Expected height '300', got '%s'", chart.Height)
	}
}

func TestChartWidgetType(t *testing.T) {
	chart := NewChart("test", "Test", Bar)
	
	if chart.GetType() != "chart" {
		t.Errorf("Expected type 'chart', got '%s'", chart.GetType())
	}
}

func TestChartSetLabels(t *testing.T) {
	chart := NewChart("test", "Test", Line).
		SetLabels([]string{"Jan", "Feb", "Mar"})
	
	if len(chart.Labels) != 3 {
		t.Errorf("Expected 3 labels, got %d", len(chart.Labels))
	}
	if chart.Labels[0] != "Jan" {
		t.Errorf("Expected first label 'Jan', got '%s'", chart.Labels[0])
	}
}

func TestChartAddSeries(t *testing.T) {
	chart := NewChart("test", "Test", Line).
		AddSeries("Sales", []int{100, 200, 300}).
		AddSeries("Revenue", []int{150, 250, 350})
	
	if len(chart.Series) != 2 {
		t.Errorf("Expected 2 series, got %d", len(chart.Series))
	}
	if chart.Series[0].Name != "Sales" {
		t.Errorf("Expected first series name 'Sales', got '%s'", chart.Series[0].Name)
	}
	if len(chart.Series[0].Data) != 3 {
		t.Errorf("Expected 3 data points, got %d", len(chart.Series[0].Data))
	}
}

func TestChartGetSeriesJSON(t *testing.T) {
	chart := NewChart("test", "Test", Line).
		AddSeries("Sales", []int{100, 200})
	
	json := chart.GetSeriesJSON()
	
	if json == "" {
		t.Error("Expected non-empty JSON")
	}
	if json == "null" {
		t.Error("Expected valid JSON, got null")
	}
}

func TestChartGetSeriesJSON_Donut(t *testing.T) {
	chart := NewChart("test", "Test", Donut).
		AddSeries("A", []int{30}).
		AddSeries("B", []int{70})
	
	json := chart.GetSeriesJSON()
	
	// Pour Donut, on attend un tableau simple [30, 70]
	if json != "[30,70]" {
		t.Errorf("Expected '[30,70]', got '%s'", json)
	}
}

func TestChartGetLabelsJSON(t *testing.T) {
	chart := NewChart("test", "Test", Line).
		SetLabels([]string{"A", "B", "C"})
	
	json := chart.GetLabelsJSON()
	
	if json != `["A","B","C"]` {
		t.Errorf("Expected '[\"A\",\"B\",\"C\"]', got '%s'", json)
	}
}

func TestChartGetColorsJSON(t *testing.T) {
	chart := NewChart("test", "Test", Line)
	
	json := chart.GetColorsJSON()
	
	if json == "" || json == "null" {
		t.Error("Expected default colors JSON")
	}
}

func TestChartTypes(t *testing.T) {
	if Line != "area" {
		t.Errorf("Expected Line to be 'area', got '%s'", Line)
	}
	if Bar != "bar" {
		t.Errorf("Expected Bar to be 'bar', got '%s'", Bar)
	}
	if Donut != "donut" {
		t.Errorf("Expected Donut to be 'donut', got '%s'", Donut)
	}
}
