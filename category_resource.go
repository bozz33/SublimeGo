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
	"github.com/bozz33/sublimeadmin/infolist"
	"github.com/bozz33/sublimeadmin/support"
	"github.com/bozz33/sublimeadmin/table"
	"github.com/bozz33/sublimeadmin/views/generics"
)

// Category is a blog category. Posts belong to a category (1-n).
type Category struct {
	ID        int
	Name      string
	Slug      string
	PostCount int
}

// CategoryResource is a SQLite-backed CRUD resource for blog categories.
type CategoryResource struct {
	*engine.BaseResource
	db *sql.DB
}

// NewCategoryResource wires the resource to the database.
func NewCategoryResource(db *sql.DB) *CategoryResource {
	r := &CategoryResource{BaseResource: engine.NewAutoResource("Category"), db: db}
	r.SetIcon("folder")

	r.SetTableColumns(
		table.Text("id").Using(func(it any) string { return strconv.Itoa(it.(*Category).ID) }).Align(table.AlignEnd).Width("w-20"),
		table.Text("name").Using(func(it any) string { return it.(*Category).Name }).Sortable().Searchable(),
		table.Text("slug").Using(func(it any) string { return it.(*Category).Slug }),
		table.Text("posts").Using(func(it any) string { return strconv.Itoa(it.(*Category).PostCount) }).Align(table.AlignEnd),
	)

	r.EnableSelection()
	r.SetTableBulkActionsFromActions(actions.BulkDeleteAction("/categories"))
	r.SetExportURL("/categories/export")

	r.SetGetOperation(r.get)
	r.SetCreateOperation(r.create)
	r.SetUpdateOperation(r.update)
	r.SetDeleteOperation(r.delete)
	r.SetBulkDeleteOperation(r.bulkDelete)
	return r
}

// ListQuery applies search, sorting and pagination, and counts posts per category.
func (r *CategoryResource) ListQuery(_ context.Context, q engine.ListQuery) ([]any, int, error) {
	where := "WHERE 1=1"
	args := []any{}
	if s := strings.TrimSpace(q.Search); s != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+s+"%")
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM categories "+where, args...).Scan(&total); err != nil {
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
		SELECT c.id, c.name, c.slug, (SELECT COUNT(*) FROM posts p WHERE p.category_id = c.id)
		FROM categories c `+where+" ORDER BY "+order+" LIMIT ? OFFSET ?", qArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []any
	for rows.Next() {
		c := &Category{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.PostCount); err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, rows.Err()
}

func (r *CategoryResource) Table(ctx context.Context) templ.Component {
	state, _ := r.BuildTableStateFor(ctx, r, r.CanCreate(ctx), r.CanDelete(ctx))
	return generics.List(state)
}

func (r *CategoryResource) View(_ context.Context, item any) templ.Component {
	c, ok := item.(*Category)
	if !ok || c == nil {
		return templ.NopComponent
	}
	il := infolist.New().AddSection(
		infolist.NewSection("Category").WithColumns(2).Add(
			infolist.TextEntry("id", "ID", c.ID).Align(infolist.AlignEnd),
			infolist.TextEntry("name", "Name", c.Name).WithCopy(),
			infolist.TextEntry("slug", "Slug", c.Slug),
			infolist.TextEntry("posts", "Posts", c.PostCount).Align(infolist.AlignEnd),
		),
	)
	return generics.Infolist(il)
}

func (r *CategoryResource) Form(_ context.Context, item any) templ.Component {
	name := form.Text("name").Required()
	slug := form.Text("slug").HelperText("URL-friendly identifier; left blank it is derived from the name.")
	if c, ok := item.(*Category); ok && c != nil {
		name = form.Text("name").Default(c.Name).Required()
		slug = form.Text("slug").Default(c.Slug)
	}
	return generics.Form(form.New().SetSchema(name, slug))
}

func (r *CategoryResource) get(_ context.Context, id string) (any, error) {
	c := &Category{}
	err := r.db.QueryRow(`SELECT id, name, slug FROM categories WHERE id = ?`, id).Scan(&c.ID, &c.Name, &c.Slug)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *CategoryResource) create(_ context.Context, req *http.Request) error {
	name := req.FormValue("name")
	slug := slugOrDerive(req.FormValue("slug"), name)
	_, err := r.db.Exec(`INSERT INTO categories (name, slug) VALUES (?, ?)`, name, slug)
	return err
}

func (r *CategoryResource) update(_ context.Context, id string, req *http.Request) error {
	name := req.FormValue("name")
	slug := slugOrDerive(req.FormValue("slug"), name)
	_, err := r.db.Exec(`UPDATE categories SET name = ?, slug = ? WHERE id = ?`, name, slug, id)
	return err
}

func (r *CategoryResource) delete(_ context.Context, id string) error {
	// Detach posts from the deleted category to avoid dangling references.
	_, _ = r.db.Exec(`UPDATE posts SET category_id = NULL WHERE category_id = ?`, id)
	_, err := r.db.Exec(`DELETE FROM categories WHERE id = ?`, id)
	return err
}

func (r *CategoryResource) bulkDelete(_ context.Context, ids []string) error {
	for _, id := range ids {
		if err := r.delete(context.Background(), id); err != nil {
			return err
		}
	}
	return nil
}

func slugOrDerive(slug, name string) string {
	if strings.TrimSpace(slug) != "" {
		return slug
	}
	return support.Slug(name)
}
