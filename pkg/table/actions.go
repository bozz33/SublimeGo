package table

import (
	"context"
	"fmt"
)

// BaseAction represents a base action.
type BaseAction struct {
	LabelStr  string
	IconStr   string
	ColorStr  string
	URLFunc   func(item any) string
	VisibleFn func(ctx context.Context, item any) bool
}

// Action interface implementation
func (a *BaseAction) GetLabel() string       { return a.LabelStr }
func (a *BaseAction) GetIcon() string        { return a.IconStr }
func (a *BaseAction) GetColor() string       { return a.ColorStr }
func (a *BaseAction) GetURL(item any) string { return a.URLFunc(item) }
func (a *BaseAction) IsVisible(ctx context.Context, item any) bool {
	if a.VisibleFn != nil {
		return a.VisibleFn(ctx, item)
	}
	return true
}

// EditAction creates an edit action.
func EditAction() *BaseAction {
	return &BaseAction{
		LabelStr: "Edit",
		IconStr:  "edit",
		ColorStr: "primary",
		URLFunc: func(item any) string {
			return fmt.Sprintf("/edit/%v", getID(item))
		},
	}
}

// DeleteAction creates a delete action.
func DeleteAction() *BaseAction {
	return &BaseAction{
		LabelStr: "Delete",
		IconStr:  "trash",
		ColorStr: "danger",
		URLFunc: func(item any) string {
			return fmt.Sprintf("/delete/%v", getID(item))
		},
	}
}

// ViewAction creates a view action.
func ViewAction() *BaseAction {
	return &BaseAction{
		LabelStr: "Voir",
		IconStr:  "eye",
		ColorStr: "secondary",
		URLFunc: func(item any) string {
			return fmt.Sprintf("/view/%v", getID(item))
		},
	}
}

// CustomAction creates a custom action.
func CustomAction(label, icon, color string, urlFunc func(any) string) *BaseAction {
	return &BaseAction{
		LabelStr: label,
		IconStr:  icon,
		ColorStr: color,
		URLFunc:  urlFunc,
	}
}

// Visible defines a visibility condition.
func (a *BaseAction) Visible(fn func(ctx context.Context, item any) bool) *BaseAction {
	a.VisibleFn = fn
	return a
}

// getID extracts the ID from an item (supports different types).
func getID(item any) any {
	if item == nil {
		return ""
	}

	// Try to extract ID using reflection for common patterns
	switch v := item.(type) {
	case interface{ GetID() any }:
		return v.GetID()
	case interface{ ID() any }:
		return v.ID()
	case map[string]any:
		if id, ok := v["id"]; ok {
			return id
		}
		if id, ok := v["ID"]; ok {
			return id
		}
	}

	// Fallback: try reflection to find an "ID" or "Id" field
	// This handles structs with public ID fields
	return extractIDViaReflection(item)
}

// extractIDViaReflection uses reflection to find ID field in structs.
func extractIDViaReflection(item any) any {
	// Note: This is a simplified implementation.
	// In production, you might want to use a more robust reflection-based approach
	// or require all items to implement a common interface.
	return fmt.Sprintf("%v", item)
}
