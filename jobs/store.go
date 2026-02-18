package jobs

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Store provides SQLite-backed persistence for jobs.
// Jobs are persisted across restarts; pending jobs are re-queued on startup.
type Store struct {
	db *sql.DB
}

// NewStore opens (or creates) the SQLite database at the given path.
func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("jobs: open store: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("jobs: ping store: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("jobs: migrate store: %w", err)
	}

	return s, nil
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// migrate creates the jobs table if it does not exist.
func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS jobs (
			id           TEXT PRIMARY KEY,
			name         TEXT NOT NULL,
			status       TEXT NOT NULL DEFAULT 'pending',
			progress     INTEGER NOT NULL DEFAULT 0,
			result       TEXT,
			error        TEXT,
			created_at   DATETIME NOT NULL,
			started_at   DATETIME,
			completed_at DATETIME
		)
	`)
	return err
}

// Save inserts or updates a job record.
func (s *Store) Save(job *Job) error {
	var resultJSON []byte
	if job.Result != nil {
		var err error
		resultJSON, err = json.Marshal(job.Result)
		if err != nil {
			resultJSON = nil
		}
	}

	var errStr *string
	if job.Error != nil {
		s := job.Error.Error()
		errStr = &s
	}

	_, err := s.db.Exec(`
		INSERT INTO jobs (id, name, status, progress, result, error, created_at, started_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			status       = excluded.status,
			progress     = excluded.progress,
			result       = excluded.result,
			error        = excluded.error,
			started_at   = excluded.started_at,
			completed_at = excluded.completed_at
	`,
		job.ID,
		job.Name,
		string(job.Status),
		job.Progress,
		nullableBytes(resultJSON),
		errStr,
		job.CreatedAt,
		job.StartedAt,
		job.CompletedAt,
	)
	return err
}

// LoadPending returns all jobs with status "pending" (to re-queue after restart).
func (s *Store) LoadPending() ([]*Job, error) {
	rows, err := s.db.Query(`
		SELECT id, name, status, progress, result, error, created_at, started_at, completed_at
		FROM jobs
		WHERE status = 'pending'
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanJobs(rows)
}

// LoadAll returns all jobs ordered by creation date descending.
func (s *Store) LoadAll() ([]*Job, error) {
	rows, err := s.db.Query(`
		SELECT id, name, status, progress, result, error, created_at, started_at, completed_at
		FROM jobs
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanJobs(rows)
}

// DeleteOlderThan removes completed/failed/cancelled jobs older than the given duration.
func (s *Store) DeleteOlderThan(d time.Duration) (int64, error) {
	threshold := time.Now().Add(-d)
	result, err := s.db.Exec(`
		DELETE FROM jobs
		WHERE status IN ('completed', 'failed', 'cancelled')
		AND completed_at < ?
	`, threshold)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// scanJobs scans SQL rows into Job slices.
func (s *Store) scanJobs(rows *sql.Rows) ([]*Job, error) {
	var jobs []*Job

	for rows.Next() {
		var (
			id          string
			name        string
			status      string
			progress    int
			resultJSON  sql.NullString
			errStr      sql.NullString
			createdAt   time.Time
			startedAt   sql.NullTime
			completedAt sql.NullTime
		)

		if err := rows.Scan(&id, &name, &status, &progress, &resultJSON, &errStr, &createdAt, &startedAt, &completedAt); err != nil {
			return nil, err
		}

		job := &Job{
			ID:        id,
			Name:      name,
			Status:    Status(status),
			Progress:  progress,
			CreatedAt: createdAt,
		}

		if startedAt.Valid {
			t := startedAt.Time
			job.StartedAt = &t
		}
		if completedAt.Valid {
			t := completedAt.Time
			job.CompletedAt = &t
		}
		if errStr.Valid {
			job.Error = fmt.Errorf("%s", errStr.String)
		}

		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

// nullableBytes returns nil if b is empty, otherwise returns b.
func nullableBytes(b []byte) interface{} {
	if len(b) == 0 {
		return nil
	}
	return string(b)
}
