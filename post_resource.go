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
	ID           int
	Title        string
	Body         string
	Status       string
	CategoryID   int    `export:"-"`
	CategoryName string `export:"category"`
	TagIDs       []int  `export:"-"` // current tag ids (loaded for the edit form)
	TagsStr      string `export:"tags"` // comma-joined tag names (for the list column)
	CreatedAt    string
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
		table.Text("category").Using(func(it any) string { return it.(*Post).CategoryName }),
		table.Text("tags").Using(func(it any) string { return it.(*Post).TagsStr }),
		table.Text("created_at").Using(func(it any) string { return it.(*Post).CreatedAt }).
			Align(table.AlignEnd),
	)

	// Grouped two-tier header (Filament-style ColumnGroup).
	r.SetTableColumnGroups(
		table.NewColumnGroup("Details", "id", "title"),
		table.NewColumnGroup("Publication", "status", "category", "tags", "created_at"),
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
		table.Select("category").WithLabel("Category").WithOptions(r.categoryFilterOptions()),
		r.qbFilter,
	)

	// Multi-row selection + bulk delete, built from the actions package
	// convenience constructor (URL, icon, color and confirmation preconfigured).
	r.EnableSelection()
	r.SetTableBulkActionsFromActions(actions.BulkDeleteAction("/posts"))
	r.SetExportURL("/posts/export") // CSV/Excel export (route auto-mounted by the panel)

	// Has-many relation: a post's comments, managed inline on the edit page.
	r.SetRelationManagers(NewCommentsRelationManager(db))

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
	if cat := q.Filters["category"]; cat != "" {
		where += " AND category_id = ?"
		args = append(args, cat)
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

	order := "p.id DESC"
	switch q.SortKey {
	case "title":
		order = "p.title " + sortDir(q.SortDir)
	case "status":
		order = "p.status " + sortDir(q.SortDir)
	case "id":
		order = "p.id " + sortDir(q.SortDir)
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
		SELECT p.id, p.title, p.body, p.status, COALESCE(p.category_id, 0), COALESCE(c.name, ''),
			COALESCE((SELECT GROUP_CONCAT(tg.name, ', ') FROM post_tags pt JOIN tags tg ON tg.id = pt.tag_id WHERE pt.post_id = p.id), ''),
			p.created_at
		FROM posts p LEFT JOIN categories c ON c.id = p.category_id
		`+where+" ORDER BY "+order+" LIMIT ? OFFSET ?", qArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []any
	for rows.Next() {
		p := &Post{}
		if err := rows.Scan(&p.ID, &p.Title, &p.Body, &p.Status, &p.CategoryID, &p.CategoryName, &p.TagsStr, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	return items, total, rows.Err()
}

// categoryFilterOptions loads categories as table filter options.
func (r *PostResource) categoryFilterOptions() []table.FilterOption {
	rows, err := r.db.Query(`SELECT id, name FROM categories ORDER BY name`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var opts []table.FilterOption
	for rows.Next() {
		var id int
		var name string
		if rows.Scan(&id, &name) == nil {
			opts = append(opts, table.FilterOption{Value: strconv.Itoa(id), Label: name})
		}
	}
	return opts
}

// categorySelectOptions loads categories as form select options (resolved live).
func (r *PostResource) categorySelectOptions(ctx context.Context) ([]form.SelectOption, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	opts := []form.SelectOption{{Value: "", Label: "— None —"}}
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		opts = append(opts, form.SelectOption{Value: strconv.Itoa(id), Label: name})
	}
	return opts, rows.Err()
}

// tagSelectOptions loads tags as form select options.
func (r *PostResource) tagSelectOptions(ctx context.Context) ([]form.SelectOption, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name FROM tags ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var opts []form.SelectOption
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		opts = append(opts, form.SelectOption{Value: strconv.Itoa(id), Label: name})
	}
	return opts, rows.Err()
}

// syncPostTags replaces the tag links for a post with the given tag ids.
func (r *PostResource) syncPostTags(postID int64, tagIDs []string) {
	_, _ = r.db.Exec(`DELETE FROM post_tags WHERE post_id = ?`, postID)
	for _, t := range tagIDs {
		if id, err := strconv.Atoi(strings.TrimSpace(t)); err == nil && id > 0 {
			_, _ = r.db.Exec(`INSERT OR IGNORE INTO post_tags (post_id, tag_id) VALUES (?, ?)`, postID, id)
		}
	}
}

// joinTagIDs joins tag ids into a comma-separated string for select pre-selection.
func joinTagIDs(ids []int) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.Itoa(id))
	}
	return strings.Join(parts, ",")
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
				infolist.TextEntry("category", "Category", p.CategoryName).WithPlaceholder("Uncategorized"),
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
func (r *PostResource) Form(ctx context.Context, item any) templ.Component {
	title := form.Text("title").Required()
	body := form.Textarea("body")
	status := form.Select("status").Options(map[string]string{
		"draft":     "Draft",
		"published": "Published",
	})
	// Category: a relation-backed select whose options are loaded from the DB.
	category := form.RelationSelect("category_id").Label("Category").OptionsFrom(r.categorySelectOptions)
	// Tags: a multi-select backed by the many-to-many post_tags relation.
	tagOpts, _ := r.tagSelectOptions(ctx)
	tags := form.Select("tags").Label("Tags").Multiple().OptionsOrdered(tagOpts)

	if p, ok := item.(*Post); ok && p != nil {
		if p.CategoryID > 0 {
			category.SetValue(strconv.Itoa(p.CategoryID))
		}
		tags.SetValue(joinTagIDs(p.TagIDs))
		title = form.Text("title").Default(p.Title).Required()
		body = form.Textarea("body")
		body.SetValue(p.Body)
		status = form.Select("status").Options(map[string]string{
			"draft":     "Draft",
			"published": "Published",
		}).Default(p.Status)

		// Display-only field (form.Placeholder) showing the creation date.
		created := form.Placeholder("created_at").Label("Created").Content(p.CreatedAt)
		return generics.Form(form.New().SetSchema(title, body, status, category, tags, created))
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
	// TableSelect: pick a related post from a searchable modal table.
	relatedPost := form.TableSelect("related_post").Label("Related post (table picker)").
		WithColumns(
			form.TableSelectColumn{Key: "id", Label: "ID"},
			form.TableSelectColumn{Key: "title", Label: "Title"},
		).
		Searchable().
		Options(func(ctx context.Context) ([]form.TableSelectRow, error) {
			rows, err := r.db.QueryContext(ctx, "SELECT id, title FROM posts ORDER BY id DESC LIMIT 50")
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			var out []form.TableSelectRow
			for rows.Next() {
				var id int
				var t string
				if err := rows.Scan(&id, &t); err != nil {
					return nil, err
				}
				sid := strconv.Itoa(id)
				out = append(out, form.TableSelectRow{
					Value: sid, Label: t,
					Cells: map[string]string{"id": sid, "title": t},
				})
			}
			return out, rows.Err()
		})
	// RelationshipRepeater: add initial comments inline while creating the post
	// (on edit, comments are managed by the relation-manager sub-table).
	comments := form.RelationshipRepeater("comments").Label("Initial comments").
		AddButtonLabel("Add comment").
		Fields(
			form.RepeaterColumn{Key: "author", Label: "Author"},
			form.RepeaterColumn{Key: "body", Label: "Comment", Type: "textarea"},
		)
	return generics.Form(form.New().SetSchema(heading, tips, title, body, status, category, tags, related, relatedPost, comments, tip))
}

// --- CRUD operations (SQLite) ---------------------------------------------

func (r *PostResource) get(_ context.Context, id string) (any, error) {
	p := &Post{}
	err := r.db.QueryRow(`
		SELECT p.id, p.title, p.body, p.status, COALESCE(p.category_id, 0), COALESCE(c.name, ''), p.created_at
		FROM posts p LEFT JOIN categories c ON c.id = p.category_id WHERE p.id = ?`, id).
		Scan(&p.ID, &p.Title, &p.Body, &p.Status, &p.CategoryID, &p.CategoryName, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if tagRows, terr := r.db.Query(`SELECT tag_id FROM post_tags WHERE post_id = ?`, p.ID); terr == nil {
		defer tagRows.Close()
		for tagRows.Next() {
			var tid int
			if tagRows.Scan(&tid) == nil {
				p.TagIDs = append(p.TagIDs, tid)
			}
		}
	}
	return p, nil
}

// nullableID converts a form value to a nullable category id (empty -> NULL).
func nullableID(s string) any {
	if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && n > 0 {
		return n
	}
	return nil
}

func (r *PostResource) create(ctx context.Context, req *http.Request) error {
	res, err := r.db.Exec(
		`INSERT INTO posts (title, body, status, category_id) VALUES (?, ?, ?, ?)`,
		req.FormValue("title"), req.FormValue("body"), statusOrDefault(req.FormValue("status")), nullableID(req.FormValue("category_id")),
	)
	if err == nil {
		if id, e := res.LastInsertId(); e == nil {
			r.syncPostTags(id, req.Form["tags"])
			// Persist the initial comments from the RelationshipRepeater.
			for _, row := range form.ParseRepeater(req, "comments") {
				author := row["author"]
				if author == "" {
					author = "Anonymous"
				}
				_, _ = r.db.Exec(`INSERT INTO comments (post_id, author, body) VALUES (?, ?, ?)`, id, author, row["body"])
			}
		}
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
	var categoryID any
	if p.CategoryID > 0 {
		categoryID = p.CategoryID
	}
	if req.Form.Has("title") {
		title = req.FormValue("title")
	}
	if req.Form.Has("body") {
		body = req.FormValue("body")
	}
	if req.Form.Has("status") {
		status = statusOrDefault(req.FormValue("status"))
	}
	if req.Form.Has("category_id") {
		categoryID = nullableID(req.FormValue("category_id"))
	}
	_, err = r.db.Exec(`UPDATE posts SET title = ?, body = ?, status = ?, category_id = ? WHERE id = ?`, title, body, status, categoryID, id)
	if err != nil {
		return err
	}
	// Only the full edit form (which includes title) syncs tags, so the inline
	// status SelectAction does not clear them.
	if req.Form.Has("title") {
		if pid, perr := strconv.ParseInt(id, 10, 64); perr == nil {
			r.syncPostTags(pid, req.Form["tags"])
		}
	}
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
