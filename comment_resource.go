package main

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/a-h/templ"
	"github.com/bozz33/sublimeadmin/engine"
	"github.com/bozz33/sublimeadmin/form"
	"github.com/bozz33/sublimeadmin/views/generics"
)

// Comment belongs to a post (has-many). Field names / json tags match the
// relation-manager column keys so the sub-table can read them via reflection.
type Comment struct {
	ID        int    `json:"id"`
	PostID    int    `json:"post_id"`
	Author    string `json:"author"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

// CommentsRelationManager exposes a post's comments as a has-many relation,
// rendered as a sub-table on the post edit/view page with inline create/delete.
type CommentsRelationManager struct {
	*engine.BaseRelationManager
	db *sql.DB
}

// NewCommentsRelationManager wires the comments relation to the database.
func NewCommentsRelationManager(db *sql.DB) *CommentsRelationManager {
	m := &CommentsRelationManager{
		BaseRelationManager: engine.NewBaseRelationManager("comments", "Comments", "comments", engine.RelationHasMany),
		db:                  db,
	}
	m.SetIcon("comment")
	return m
}

// ListRelated returns the comments of a given post.
func (m *CommentsRelationManager) ListRelated(ctx context.Context, parentID string) ([]any, error) {
	rows, err := m.db.QueryContext(ctx, `SELECT id, post_id, author, body, created_at FROM comments WHERE post_id = ? ORDER BY id DESC`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []any
	for rows.Next() {
		c := &Comment{}
		if err := rows.Scan(&c.ID, &c.PostID, &c.Author, &c.Body, &c.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

// CreateRelated inserts a comment linked to the parent post.
func (m *CommentsRelationManager) CreateRelated(ctx context.Context, parentID string, r *http.Request) error {
	author := r.FormValue("author")
	if author == "" {
		author = "Anonymous"
	}
	_, err := m.db.ExecContext(ctx, `INSERT INTO comments (post_id, author, body) VALUES (?, ?, ?)`, parentID, author, r.FormValue("body"))
	return err
}

// DeleteRelated removes a comment from the parent post.
func (m *CommentsRelationManager) DeleteRelated(ctx context.Context, parentID, relatedID string) error {
	_, err := m.db.ExecContext(ctx, `DELETE FROM comments WHERE id = ? AND post_id = ?`, relatedID, parentID)
	return err
}

// Columns defines the sub-table columns (keys match Comment json tags).
func (m *CommentsRelationManager) Columns() []engine.Column {
	return []engine.Column{
		{Key: "author", Label: "Author"},
		{Key: "body", Label: "Body"},
		{Key: "created_at", Label: "Created"},
	}
}

// CanAttach is false: comments are created inline, not attached (has-many).
func (m *CommentsRelationManager) CanAttach(context.Context) bool { return false }

// Form renders the inline create form for a new comment.
func (m *CommentsRelationManager) Form(_ context.Context, _ string) templ.Component {
	return generics.Form(form.New().SetSchema(
		form.Text("author").Required(),
		form.Textarea("body"),
	))
}
