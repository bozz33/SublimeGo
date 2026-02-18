package table

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBulkDelete(t *testing.T) {
	ba := BulkDelete()
	assert.Equal(t, "Delete selected", ba.GetLabel())
	assert.Equal(t, "trash", ba.GetIcon())
	assert.Equal(t, "danger", ba.GetColor())
	assert.True(t, ba.RequireConf)
	assert.NotEmpty(t, ba.ConfTitle)
	assert.NotEmpty(t, ba.ConfDesc)
}

func TestBulkExport(t *testing.T) {
	ba := BulkExport()
	assert.Equal(t, "Export selected", ba.GetLabel())
	assert.Equal(t, "download", ba.GetIcon())
	assert.Equal(t, "secondary", ba.GetColor())
	assert.False(t, ba.RequireConf)
}

func TestNewBulkAction(t *testing.T) {
	ba := NewBulkAction("Archive", "archive", "warning")
	assert.Equal(t, "Archive", ba.GetLabel())
	assert.Equal(t, "archive", ba.GetIcon())
	assert.Equal(t, "warning", ba.GetColor())
}

func TestBulkActionWithHandler(t *testing.T) {
	called := false
	var receivedIDs []string

	ba := NewBulkAction("Test", "check", "primary").
		WithHandler(func(ctx context.Context, ids []string) error {
			called = true
			receivedIDs = ids
			return nil
		})

	err := ba.Handler(context.Background(), []string{"1", "2", "3"})
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, []string{"1", "2", "3"}, receivedIDs)
}

func TestBulkActionWithHandlerError(t *testing.T) {
	ba := NewBulkAction("Fail", "x", "danger").
		WithHandler(func(ctx context.Context, ids []string) error {
			return errors.New("bulk failed")
		})

	err := ba.Handler(context.Background(), []string{"1"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bulk failed")
}

func TestBulkActionRequireConfirmation(t *testing.T) {
	ba := NewBulkAction("Purge", "trash", "danger").
		RequireConfirmation("Purge records", "This cannot be undone.")

	assert.True(t, ba.RequireConf)
	assert.Equal(t, "Purge records", ba.ConfTitle)
	assert.Equal(t, "This cannot be undone.", ba.ConfDesc)
}

func TestBulkActionIsVisible(t *testing.T) {
	ba := NewBulkAction("Test", "x", "primary")
	assert.True(t, ba.IsVisible(context.Background()))

	ba.VisibleWhen(func(ctx context.Context) bool { return false })
	assert.False(t, ba.IsVisible(context.Background()))
}

func TestTableWithBulkActions(t *testing.T) {
	tbl := New([]any{}).
		WithBulkActions(BulkDelete(), BulkExport())

	assert.Len(t, tbl.BulkActions, 2)
}
