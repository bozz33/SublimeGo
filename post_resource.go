package main

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/bozz33/sublimeadmin/engine"
	"github.com/bozz33/sublimeadmin/form"
	"github.com/bozz33/sublimeadmin/table"
	"github.com/bozz33/sublimeadmin/views/generics"
)

// Post is the domain model backed by the posts table.
type Post struct {
	ID        int
	Title     string
	Body      string
	Status    string
	CreatedAt string
}

// PostResource is a real, SQLite-backed resource: list, view, create, edit,
// delete, with a declarative table and form.
type PostResource struct {
	*engine.BaseResource
	db *sql.DB
}

// NewPostResource wires the resource to the database.
func NewPostResource(db *sql.DB) *PostResource {
	r := &PostResource{BaseResource: engine.NewAutoResource("Post"), db: db}
	r.SetIcon("article")

	r.SetTableColumns(
		table.Text("id").Using(func(it any) string { return strconv.Itoa(it.(*Post).ID) }),
		table.Text("title").Using(func(it any) string { return it.(*Post).Title }).Sortable().Searchable(),
		table.Text("status").Using(func(it any) string { return it.(*Post).Status }).Badge().
			WithColorFunc(func(value string, _ any) string {
				if value == "published" {
					return "green"
				}
				return "gray"
			}),
		table.Text("created_at").Using(func(it any) string { return it.(*Post).CreatedAt }),
	)

	// Status filter (appears in the list toolbar).
	r.SetTypedFilters(
		table.Select("status").WithLabel("Status").WithOptions([]table.FilterOption{
			{Value: "draft", Label: "Draft"},
			{Value: "published", Label: "Published"},
		}),
	)

	// Multi-row selection + bulk delete.
	r.EnableSelection()
	r.SetTableBulkActions(engine.BulkActionDef{
		Key:         "delete",
		Label:       "Delete selected",
		Icon:        "delete",
		Color:       "danger",
		URL:         "/posts/bulk-delete",
		Method:      "POST",
		RequireConf: true,
		ConfTitle:   "Delete posts",
		ConfDesc:    "This will permanently delete the selected posts.",
	})

	r.SetGetOperation(r.get)
	r.SetCreateOperation(r.create)
	r.SetUpdateOperation(r.update)
	r.SetDeleteOperation(r.delete)
	r.SetBulkDeleteOperation(r.bulkDelete)
	return r
}

// ListQuery applies search, the status filter, sorting and pagination in SQL.
// Implementing ResourceQueryable makes the toolbar controls actually filter data.
func (r *PostResource) ListQuery(_ context.Context, q engine.ListQuery) ([]any, int, error) {
	where := "WHERE 1=1"
	args := []any{}
	if s := strings.TrimSpace(q.Search); s != "" {
		where += " AND title LIKE ?"
		args = append(args, "%"+s+"%")
	}
	if st := q.Filters["status"]; st != "" {
		where += " AND status = ?"
		args = append(args, st)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM posts "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	order := "id DESC"
	switch q.SortKey {
	case "title":
		order = "title " + sortDir(q.SortDir)
	case "status":
		order = "status " + sortDir(q.SortDir)
	case "id":
		order = "id " + sortDir(q.SortDir)
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
	rows, err := r.db.Query("SELECT id, title, body, status, created_at FROM posts "+where+" ORDER BY "+order+" LIMIT ? OFFSET ?", qArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []any
	for rows.Next() {
		p := &Post{}
		if err := rows.Scan(&p.ID, &p.Title, &p.Body, &p.Status, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	return items, total, rows.Err()
}

func sortDir(d string) string {
	if strings.EqualFold(d, "asc") {
		return "ASC"
	}
	return "DESC"
}

// Table renders the resource list via the official generic table render.
func (r *PostResource) Table(ctx context.Context) templ.Component {
	state, _ := r.BuildTableStateFor(ctx, r, r.CanCreate(ctx), r.CanDelete(ctx))
	return generics.List(state)
}

// Form renders the create/edit form, pre-filled when editing an existing post.
func (r *PostResource) Form(_ context.Context, item any) templ.Component {
	title := form.Text("title").Required()
	body := form.Textarea("body")
	status := form.Select("status").Options(map[string]string{
		"draft":     "Draft",
		"published": "Published",
	})

	if p, ok := item.(*Post); ok && p != nil {
		title = form.Text("title").Default(p.Title).Required()
		body = form.Textarea("body")
		body.SetValue(p.Body)
		status = form.Select("status").Options(map[string]string{
			"draft":     "Draft",
			"published": "Published",
		}).Default(p.Status)
	}

	return generics.Form(form.New().SetSchema(title, body, status))
}

// --- CRUD operations (SQLite) ---------------------------------------------

func (r *PostResource) get(_ context.Context, id string) (any, error) {
	p := &Post{}
	err := r.db.QueryRow(`SELECT id, title, body, status, created_at FROM posts WHERE id = ?`, id).
		Scan(&p.ID, &p.Title, &p.Body, &p.Status, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PostResource) create(_ context.Context, req *http.Request) error {
	_, err := r.db.Exec(
		`INSERT INTO posts (title, body, status) VALUES (?, ?, ?)`,
		req.FormValue("title"), req.FormValue("body"), statusOrDefault(req.FormValue("status")),
	)
	return err
}

func (r *PostResource) update(_ context.Context, id string, req *http.Request) error {
	_, err := r.db.Exec(
		`UPDATE posts SET title = ?, body = ?, status = ? WHERE id = ?`,
		req.FormValue("title"), req.FormValue("body"), statusOrDefault(req.FormValue("status")), id,
	)
	return err
}

func (r *PostResource) delete(_ context.Context, id string) error {
	_, err := r.db.Exec(`DELETE FROM posts WHERE id = ?`, id)
	return err
}

func (r *PostResource) bulkDelete(_ context.Context, ids []string) error {
	for _, id := range ids {
		if _, err := r.db.Exec(`DELETE FROM posts WHERE id = ?`, id); err != nil {
			return err
		}
	}
	return nil
}

func statusOrDefault(s string) string {
	if s == "published" {
		return "published"
	}
	return "draft"
}
