//go:build !cgo
// +build !cgo

package commands

import (
	"database/sql"

	"modernc.org/sqlite"
)

func init() {
	// Enregistrer modernc.org/sqlite avec le nom "sqlite3" pour compatibilité avec Ent
	sql.Register("sqlite3", &sqlite.Driver{})
}
