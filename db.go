package main

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

// openDB opens (and migrates/seeds) the SQLite database used by the starter.
// modernc.org/sqlite is a pure-Go driver (no CGO), registered as "sqlite".
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
	} {
		if _, err := db.Exec(pragma); err != nil {
			log.Fatalf("configure db (%s): %v", pragma, err)
		}
	}
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			title      TEXT NOT NULL,
			body       TEXT NOT NULL DEFAULT '',
			status     TEXT NOT NULL DEFAULT 'draft',
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		);
	`); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM posts`).Scan(&count); err == nil && count == 0 {
		seed := []struct{ title, body, status string }{
			{"Welcome to SublimeGo", "Your first post, served from SQLite.", "published"},
			{"Draft idea", "Something to finish later.", "draft"},
			{"Release notes", "What changed in this version.", "published"},
		}
		for _, p := range seed {
			if _, err := db.Exec(`INSERT INTO posts (title, body, status) VALUES (?, ?, ?)`, p.title, p.body, p.status); err != nil {
				log.Printf("seed: %v", err)
			}
		}
	}
	return db
}
