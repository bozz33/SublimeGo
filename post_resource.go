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
	"github.com/bozz33/sublimeadmin/notifications"
	"github.com/bozz33/sublimeadmin/schema"
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
	db       *sql.DB
	qbFilter *table.QueryBuilderFilter
}

// NewPostResource wires the resource to the database.
func NewPostResource(db *sql.DB) *PostResource {
	r := &PostResource{BaseResource: engine.NewAutoResource("Post"), db: db}
	r.SetIcon("article")

	r.SetTableColumns(
		table.Text("id").Using(func(it any) string { return strconv.Itoa(it.(*Post).ID) }).
			Align(table.AlignEnd).Width("w-20"),
		table.Text("title").Using(func(it any) string { return it.(*Post).Title }).Sortable().Searchable().
			Tooltip("The post title"),
		table.Text("status").Using(func(it any) string { return it.(*Post).Status }).Badge().
			Align(table.AlignCenter).
			WithColorFunc(func(value string, _ any) string {
				if value == "published" {
					return "green"
				}
				return "gray"
			}),
		table.Text("created_at").Using(func(it any) string { return it.(*Post).CreatedAt }).
			Align(table.AlignEnd),
	)

	// Grouped two-tier header (Filament-style ColumnGroup).
	r.SetTableColumnGroups(
		table.NewColumnGroup("Details", "id", "title"),
		table.NewColumnGroup("Publication", "status", "created_at"),
	)

	// Inline SelectAction in each row to change a post's status.
	r.SetTableRowActions(
		actions.SelectAction("status").
			WithSelectParam("status").
			SetUrl(func(it any) string { return "/posts/" + strconv.Itoa(it.(*Post).ID) }).
			WithOptions(
				actions.ActionSelectOption{Value: "draft", Label: "Set draft"},
				actions.ActionSelectOption{Value: "published", Label: "Set published"},
			),
	)

	// Status filter + visual AND/OR query builder (both appear in the toolbar).
	r.qbFilter = table.QueryBuilder("conditions").WithLabel("Advanced").
		TextField("title", "Title").
		SelectField("status", "Status", []table.FilterOption{
			{Value: "draft", Label: "Draft"},
			{Value: "published", Label: "Published"},
		}).
		NumberField("id", "ID").
		DateField("created_at", "Created")
	r.SetTypedFilters(
		table.Select("status").WithLabel("Status").WithOptions([]table.FilterOption{
			{Value: "draft", Label: "Draft"},
			{Value: "published", Label: "Published"},
		}),
		r.qbFilter,
	)

	// Multi-row selection + bulk delete, built from the actions package
	// convenience constructor (URL, icon, color and confirmation preconfigured).
	r.EnableSelection()
	r.SetTableBulkActionsFromActions(actions.BulkDeleteAction("/posts"))

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
	// Visual query builder: parse the submitted JSON conditions into SQL,
	// restricted to the fields the filter actually exposes (ParseValue) so a
	// crafted payload cannot probe other columns.
	if b, ok := r.qbFilter.ParseValue(q.Filters["conditions"]); ok {
		if frag, qbArgs := b.ToSQL(); frag != "" {
			where += " AND (" + frag + ")"
			args = append(args, qbArgs...)
		}
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

// View renders the read-only detail view (Infolist) for a post. Implementing
// ResourceViewable enables GET /posts/{id} and the row View action.
func (r *PostResource) View(_ context.Context, item any) templ.Component {
	p, ok := item.(*Post)
	if !ok || p == nil {
		return templ.NopComponent
	}
	statusColor := "gray"
	if p.Status == "published" {
		statusColor = "green"
	}
	il := infolist.New().
		AddSection(
			infolist.NewSection("Details").WithColumns(2).Add(
				infolist.TextEntry("id", "ID", p.ID).Align(infolist.AlignEnd),
				infolist.BadgeEntry("status", "Status", p.Status, statusColor),
				infolist.TextEntry("title", "Title", p.Title).WithCopy(),
				infolist.DateEntry("created_at", "Created", p.CreatedAt, "2006-01-02 15:04"),
			),
		).
		AddSection(
			infolist.NewSection("Content").WithColumns(1).Add(
				infolist.TextEntry("body", "Body", p.Body).WithPlaceholder("No content"),
			),
		)
	return generics.Infolist(il)
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

		// Display-only field (form.Placeholder) showing the creation date.
		created := form.Placeholder("created_at").Label("Created").Content(p.CreatedAt)
		return generics.Form(form.New().SetSchema(title, body, status, created))
	}

	// schema primes are static content placed between fields (no input).
	heading := schema.Text("Compose your post").WithWeight("semibold")
	tips := schema.UnorderedList(
		"Keep titles short and descriptive",
		"Drafts are only visible to editors",
	)
	// form.View embeds an arbitrary component into the form (custom-UI escape
	// hatch, equivalent to Filament's ViewField).
	tip := form.View("tip", templ.Raw(
		`<div class="rounded-lg bg-primary-50 dark:bg-primary-900/20 border border-primary-200 dark:border-primary-800 p-3 text-sm text-primary-800 dark:text-primary-200">Tip: published posts appear immediately in the public list.</div>`,
	)).Label("")
	// MorphToSelect: a polymorphic relationship picker (type + record), with
	// the "post" options resolved live from the database.
	related := form.MorphToSelect("related").Label("Related to").
		Type("post", "Post", func(ctx context.Context) ([]form.SelectOption, error) {
			rows, err := r.db.QueryContext(ctx, "SELECT id, title FROM posts ORDER BY id DESC LIMIT 5")
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			var opts []form.SelectOption
			for rows.Next() {
				var id int
				var t string
				if err := rows.Scan(&id, &t); err != nil {
					return nil, err
				}
				opts = append(opts, form.SelectOption{Value: strconv.Itoa(id), Label: t})
			}
			return opts, rows.Err()
		}).
		Type("category", "Category", func(context.Context) ([]form.SelectOption, error) {
			return []form.SelectOption{{Value: "news", Label: "News"}, {Value: "tech", Label: "Tech"}}, nil
		})
	return generics.Form(form.New().SetSchema(heading, tips, title, body, status, related, tip))
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

func (r *PostResource) create(ctx context.Context, req *http.Request) error {
	_, err := r.db.Exec(
		`INSERT INTO posts (title, body, status) VALUES (?, ?, ?)`,
		req.FormValue("title"), req.FormValue("body"), statusOrDefault(req.FormValue("status")),
	)
	if err == nil {
		// Send(ctx) resolves the current authenticated user automatically.
		notifications.Success("Post created").Send(ctx)
	}
	return err
}

func (r *PostResource) update(ctx context.Context, id string, req *http.Request) error {
	// Preserve fields the request did not submit, so a partial update (such as
	// the inline status SelectAction) does not clobber title/body.
	cur, err := r.get(ctx, id)
	if err != nil || cur == nil {
		return err
	}
	p := cur.(*Post)
	title, body, status := p.Title, p.Body, p.Status
	if req.Form.Has("title") {
		title = req.FormValue("title")
	}
	if req.Form.Has("body") {
		body = req.FormValue("body")
	}
	if req.Form.Has("status") {
		status = statusOrDefault(req.FormValue("status"))
	}
	_, err = r.db.Exec(`UPDATE posts SET title = ?, body = ?, status = ? WHERE id = ?`, title, body, status, id)
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
