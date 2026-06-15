package main

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

// openDB opens (and migrates/seeds) the SQLite database used by the starter.
// modernc.org/sqlite is a pure-Go driver (no CGO), registered as "sqlite".
//
// The schema models a small blog so the demo can exercise real relations:
//
//	categories 1───n posts n───n tags        (via post_tags)
//	                  posts 1───n comments
func openDB(path string) *sql.DB {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	// SQLite allows many readers but only one writer. Keeping one pooled
	// connection and using WAL avoids stale rollback journals after restarts.
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		log.Fatalf("connect db: %v", err)
	}
	for _, pragma := range []string{
		"PRAGMA busy_timeout = 5000",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA foreign_keys = ON",
	} {
		if _, err := db.Exec(pragma); err != nil {
			log.Fatalf("configure db (%s): %v", pragma, err)
		}
	}

	schema := []string{
		`CREATE TABLE IF NOT EXISTS categories (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL DEFAULT ''
		);`,
		`CREATE TABLE IF NOT EXISTS posts (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			title       TEXT NOT NULL,
			body        TEXT NOT NULL DEFAULT '',
			status      TEXT NOT NULL DEFAULT 'draft',
			category_id INTEGER,
			created_at  TEXT NOT NULL DEFAULT (datetime('now'))
		);`,
		`CREATE TABLE IF NOT EXISTS tags (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS post_tags (
			post_id INTEGER NOT NULL,
			tag_id  INTEGER NOT NULL,
			PRIMARY KEY (post_id, tag_id)
		);`,
		`CREATE TABLE IF NOT EXISTS comments (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			post_id    INTEGER NOT NULL,
			author     TEXT NOT NULL DEFAULT '',
			body       TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		);`,
	}
	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			log.Fatalf("migrate: %v", err)
		}
	}

	// Older databases created before category_id existed: add the column.
	if !columnExists(db, "posts", "category_id") {
		if _, err := db.Exec(`ALTER TABLE posts ADD COLUMN category_id INTEGER`); err != nil {
			log.Printf("migrate category_id: %v", err)
		}
	}

	seedBlog(db)
	return db
}

// columnExists reports whether a table has a given column.
func columnExists(db *sql.DB, table, column string) bool {
	rows, err := db.Query(`SELECT name FROM pragma_table_info(?)`, table)
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if rows.Scan(&name) == nil && name == column {
			return true
		}
	}
	return false
}

// seedBlog inserts sample categories, tags, posts (with a category), post-tag
// links and comments when the database is empty.
func seedBlog(db *sql.DB) {
	var posts int
	if err := db.QueryRow(`SELECT COUNT(*) FROM posts`).Scan(&posts); err != nil || posts > 0 {
		return
	}

	cats := []struct{ name, slug string }{
		{"Announcements", "announcements"},
		{"Engineering", "engineering"},
		{"Tutorials", "tutorials"},
	}
	catIDs := make([]int64, 0, len(cats))
	for _, c := range cats {
		res, err := db.Exec(`INSERT INTO categories (name, slug) VALUES (?, ?)`, c.name, c.slug)
		if err != nil {
			log.Printf("seed category: %v", err)
			continue
		}
		id, _ := res.LastInsertId()
		catIDs = append(catIDs, id)
	}

	tags := []string{"go", "templ", "sqlite", "release", "guide"}
	tagIDs := make([]int64, 0, len(tags))
	for _, t := range tags {
		res, err := db.Exec(`INSERT INTO tags (name) VALUES (?)`, t)
		if err != nil {
			log.Printf("seed tag: %v", err)
			continue
		}
		id, _ := res.LastInsertId()
		tagIDs = append(tagIDs, id)
	}

	type seedPost struct {
		title, body, status string
		cat                 int // index into catIDs
		tags                []int
		comments            []string
	}
	seeds := []seedPost{
		{"Welcome to SublimeGo", "Your first post, served from SQLite.", "published", 0, []int{0, 3}, []string{"Great start!", "Looking forward to more."}},
		{"How the table builder works", "A deep dive into declarative tables.", "published", 1, []int{0, 1}, []string{"Very helpful."}},
		{"Draft idea", "Something to finish later.", "draft", 2, []int{4}, nil},
		{"Release notes v1", "What changed in this version.", "published", 0, []int{3}, []string{"Thanks for the update."}},
		{"Building forms", "Fields, layouts and validation.", "draft", 2, []int{1, 4}, nil},
	}
	for _, s := range seeds {
		var catID any
		if s.cat < len(catIDs) {
			catID = catIDs[s.cat]
		}
		res, err := db.Exec(`INSERT INTO posts (title, body, status, category_id) VALUES (?, ?, ?, ?)`, s.title, s.body, s.status, catID)
		if err != nil {
			log.Printf("seed post: %v", err)
			continue
		}
		postID, _ := res.LastInsertId()
		for _, ti := range s.tags {
			if ti < len(tagIDs) {
				_, _ = db.Exec(`INSERT OR IGNORE INTO post_tags (post_id, tag_id) VALUES (?, ?)`, postID, tagIDs[ti])
			}
		}
		for _, c := range s.comments {
			_, _ = db.Exec(`INSERT INTO comments (post_id, author, body) VALUES (?, ?, ?)`, postID, "Reader", c)
		}
	}
}
