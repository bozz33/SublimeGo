package table

import "context"

// BulkAction represents an action that applies to multiple selected rows.
type BulkAction struct {
	LabelStr    string
	IconStr     string
	ColorStr    string
	RequireConf bool
	ConfTitle   string
	ConfDesc    string
	Handler     func(ctx context.Context, ids []string) error
	VisibleFn   func(ctx context.Context) bool
}

// BulkDelete creates a pre-configured bulk delete action.
func BulkDelete() *BulkAction {
	return &BulkAction{
		LabelStr:    "Delete selected",
		IconStr:     "trash",
		ColorStr:    "danger",
		RequireConf: true,
		ConfTitle:   "Delete selected records",
		ConfDesc:    "Are you sure you want to delete the selected records? This action cannot be undone.",
	}
}

// BulkExport creates a pre-configured bulk export action.
func BulkExport() *BulkAction {
	return &BulkAction{
		LabelStr: "Export selected",
		IconStr:  "download",
		ColorStr: "secondary",
	}
}

// NewBulkAction creates a custom bulk action.
func NewBulkAction(label, icon, color string) *BulkAction {
	return &BulkAction{
		LabelStr: label,
		IconStr:  icon,
		ColorStr: color,
	}
}

// WithHandler sets the server-side handler for the bulk action.
func (b *BulkAction) WithHandler(fn func(ctx context.Context, ids []string) error) *BulkAction {
	b.Handler = fn
	return b
}

// RequireConfirmation enables a confirmation modal before executing.
func (b *BulkAction) RequireConfirmation(title, desc string) *BulkAction {
	b.RequireConf = true
	b.ConfTitle = title
	b.ConfDesc = desc
	return b
}

// VisibleWhen sets a visibility condition for the action.
func (b *BulkAction) VisibleWhen(fn func(ctx context.Context) bool) *BulkAction {
	b.VisibleFn = fn
	return b
}

// IsVisible returns whether the action should be shown.
func (b *BulkAction) IsVisible(ctx context.Context) bool {
	if b.VisibleFn != nil {
		return b.VisibleFn(ctx)
	}
	return true
}

// GetLabel returns the action label.
func (b *BulkAction) GetLabel() string { return b.LabelStr }

// GetIcon returns the action icon name.
func (b *BulkAction) GetIcon() string { return b.IconStr }

// GetColor returns the action color variant.
func (b *BulkAction) GetColor() string { return b.ColorStr }
