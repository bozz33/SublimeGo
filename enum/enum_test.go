package enum_test

import (
	"testing"

	"github.com/bozz33/sublimego/enum"
)

// --- test enum type ---

type Status int

const (
	StatusActive   Status = iota
	StatusInactive Status = iota
	StatusPending  Status = iota
)

func (s Status) Label() string {
	switch s {
	case StatusActive:
		return "Active"
	case StatusInactive:
		return "Inactive"
	case StatusPending:
		return "Pending"
	}
	return "Unknown"
}

func (s Status) Color() string {
	switch s {
	case StatusActive:
		return "green"
	case StatusInactive:
		return "gray"
	case StatusPending:
		return "yellow"
	}
	return ""
}

func (s Status) Icon() string {
	switch s {
	case StatusActive:
		return "check_circle"
	case StatusInactive:
		return "cancel"
	case StatusPending:
		return "schedule"
	}
	return ""
}

func (s Status) String() string { return s.Label() }

var allStatuses = []Status{StatusActive, StatusInactive, StatusPending}

func TestOptions(t *testing.T) {
	opts := enum.Options(allStatuses)
	if len(opts) != 3 {
		t.Fatalf("expected 3 options, got %d", len(opts))
	}
	if opts[0].Label != "Active" {
		t.Errorf("expected label 'Active', got %q", opts[0].Label)
	}
	if opts[1].Value != "Inactive" {
		t.Errorf("expected value 'Inactive', got %q", opts[1].Value)
	}
}

func TestLabels(t *testing.T) {
	labels := enum.Labels(allStatuses)
	if labels["Active"] != "Active" {
		t.Errorf("expected 'Active' -> 'Active', got %q", labels["Active"])
	}
	if len(labels) != 3 {
		t.Errorf("expected 3 labels, got %d", len(labels))
	}
}

func TestColors(t *testing.T) {
	colors := enum.Colors(allStatuses)
	if colors["Active"] != "green" {
		t.Errorf("expected 'Active' -> 'green', got %q", colors["Active"])
	}
	if colors["Inactive"] != "gray" {
		t.Errorf("expected 'Inactive' -> 'gray', got %q", colors["Inactive"])
	}
}

func TestIcons(t *testing.T) {
	icons := enum.Icons(allStatuses)
	if icons["Active"] != "check_circle" {
		t.Errorf("expected 'Active' -> 'check_circle', got %q", icons["Active"])
	}
}

func TestBadgeColor(t *testing.T) {
	// BadgeColor(values []T, value string, defaultColor string) string
	got := enum.BadgeColor(allStatuses, "Active", "gray")
	if got != "green" {
		t.Errorf("BadgeColor Active = %q, want 'green'", got)
	}
	got = enum.BadgeColor(allStatuses, "Inactive", "gray")
	if got != "gray" {
		t.Errorf("BadgeColor Inactive = %q, want 'gray'", got)
	}
	got = enum.BadgeColor(allStatuses, "nonexistent", "default")
	if got != "default" {
		t.Errorf("BadgeColor nonexistent = %q, want 'default'", got)
	}
}

func TestOptionsFromStringer(t *testing.T) {
	opts := enum.OptionsFromStringer(allStatuses)
	if len(opts) != 3 {
		t.Fatalf("expected 3 options, got %d", len(opts))
	}
	if opts[0].Value != "Active" {
		t.Errorf("expected value 'Active', got %q", opts[0].Value)
	}
}
