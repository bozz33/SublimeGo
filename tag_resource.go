package main

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/bozz33/sublimeadmin/actions"
	"github.com/bozz33/sublimeadmin/engine"
	"github.com/bozz33/sublimeadmin/form"
	"github.com/bozz33/sublimeadmin/table"
	"github.com/bozz33/sublimeadmin/views/generics"
)

// Tag is a blog tag. Posts and tags form a many-to-many relation via post_tags.
type Tag struct {
	ID        int
	Name      string
	PostCount int
}

// TagResource is a SQLite-backed CRUD resource for blog tags.
type TagResource struct {
	*engine.BaseResource
	db *sql.DB
}

// NewTagResource wires the resource to the database.
func NewTagResource(db *sql.DB) *TagResource {
	r := &TagResource{BaseResource: engine.NewAutoResource("Tag"), db: db}
	r.SetIcon("label")

	r.SetTableColumns(
		table.Text("id").Using(func(it any) string { return strconv.Itoa(it.(*Tag).ID) }).Align(table.AlignEnd).Width("w-20"),
		table.Text("name").Using(func(it any) string { return it.(*Tag).Name }).Sortable().Searchable(),
		table.Text("posts").Using(func(it any) string { return strconv.Itoa(it.(*Tag).PostCount) }).Align(table.AlignEnd),
	)

	r.EnableSelection()
	r.SetTableBulkActionsFromActions(actions.BulkDeleteAction("/tags"))

	r.SetGetOperation(r.get)
	r.SetCreateOperation(r.create)
	r.SetUpdateOperation(r.update)
	r.SetDeleteOperation(r.delete)
	r.SetBulkDeleteOperation(r.bulkDelete)
	return r
}

// ListQuery applies search, sorting and pagination, counting posts per tag.
func (r *TagResource) ListQuery(_ context.Context, q engine.ListQuery) ([]any, int, error) {
	where := "WHERE 1=1"
	args := []any{}
	if s := strings.TrimSpace(q.Search); s != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+s+"%")
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM tags "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	order := "id DESC"
	if q.SortKey == "name" {
		order = "name " + sortDir(q.SortDir)
	}

	perPage := q.PerPage
	if perPage <= 0 {
		perPage = 15
	}
	page := q.Page
	if page <= 0 {
		page = 1
	}
	qArgs := append(append([]any{}, args...), perPage, (page-1)*perPage)
	rows, err := r.db.Query(`
		SELECT t.id, t.name, (SELECT COUNT(*) FROM post_tags pt WHERE pt.tag_id = t.id)
		FROM tags t `+where+" ORDER BY "+order+" LIMIT ? OFFSET ?", qArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []any
	for rows.Next() {
		t := &Tag{}
		if err := rows.Scan(&t.ID, &t.Name, &t.PostCount); err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, rows.Err()
}

func (r *TagResource) Table(ctx context.Context) templ.Component {
	state, _ := r.BuildTableStateFor(ctx, r, r.CanCreate(ctx), r.CanDelete(ctx))
	return generics.List(state)
}

func (r *TagResource) Form(_ context.Context, item any) templ.Component {
	name := form.Text("name").Required()
	if t, ok := item.(*Tag); ok && t != nil {
		name = form.Text("name").Default(t.Name).Required()
	}
	return generics.Form(form.New().SetSchema(name))
}

func (r *TagResource) get(_ context.Context, id string) (any, error) {
	t := &Tag{}
	err := r.db.QueryRow(`SELECT id, name FROM tags WHERE id = ?`, id).Scan(&t.ID, &t.Name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func (r *TagResource) create(_ context.Context, req *http.Request) error {
	_, err := r.db.Exec(`INSERT INTO tags (name) VALUES (?)`, req.FormValue("name"))
	return err
}

func (r *TagResource) update(_ context.Context, id string, req *http.Request) error {
	_, err := r.db.Exec(`UPDATE tags SET name = ? WHERE id = ?`, req.FormValue("name"), id)
	return err
}

func (r *TagResource) delete(_ context.Context, id string) error {
	_, _ = r.db.Exec(`DELETE FROM post_tags WHERE tag_id = ?`, id)
	_, err := r.db.Exec(`DELETE FROM tags WHERE id = ?`, id)
	return err
}

func (r *TagResource) bulkDelete(_ context.Context, ids []string) error {
	for _, id := range ids {
		if err := r.delete(context.Background(), id); err != nil {
			return err
		}
	}
	return nil
}
