package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/yourusername/the-engine/internal/domain"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path+"?_busy_timeout=5000&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	s := &Store{db: db}
	if err := s.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) init() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS resources (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			provider TEXT NOT NULL,
			region TEXT NOT NULL,
			size TEXT NOT NULL,
			composition TEXT NOT NULL,
			labels TEXT NOT NULL,
			status TEXT NOT NULL,
			status_message TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			last_activity_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_resources_provider ON resources(provider);`,
		`CREATE TABLE IF NOT EXISTS cleanup_settings (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			mode TEXT NOT NULL
		);`,
		`INSERT OR IGNORE INTO cleanup_settings (id, mode) VALUES (1, 'disabled');`,
		`CREATE TABLE IF NOT EXISTS cost_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			resource_id TEXT NOT NULL,
			provider TEXT NOT NULL,
			amount_usd REAL NOT NULL,
			recorded_at TEXT NOT NULL,
			tags TEXT NOT NULL
		);`,
	}
	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("initialize database: %w", err)
		}
	}
	return nil
}

func (s *Store) SaveResource(r domain.Resource) error {
	labels, err := json.Marshal(r.Labels)
	if err != nil {
		return fmt.Errorf("encode resource labels: %w", err)
	}
	_, err = s.db.Exec(`INSERT INTO resources (
		id, name, type, provider, region, size, composition, labels, status, status_message,
		created_at, updated_at, last_activity_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		name = excluded.name,
		type = excluded.type,
		provider = excluded.provider,
		region = excluded.region,
		size = excluded.size,
		composition = excluded.composition,
		labels = excluded.labels,
		status = excluded.status,
		status_message = excluded.status_message,
		updated_at = excluded.updated_at,
		last_activity_at = excluded.last_activity_at`,
		r.ID, r.Name, r.Type, r.Provider, r.Region, r.Size, r.Composition, string(labels),
		string(r.Status), r.StatusMessage, formatTime(r.CreatedAt), formatTime(r.UpdatedAt), formatTime(r.LastActivityAt),
	)
	if err != nil {
		return fmt.Errorf("save resource %q: %w", r.ID, err)
	}
	return nil
}

func (s *Store) GetResource(id string) (domain.Resource, error) {
	row := s.db.QueryRow(`SELECT id, name, type, provider, region, size, composition, labels, status,
		status_message, created_at, updated_at, last_activity_at FROM resources WHERE id = ?`, id)
	return scanResource(row)
}

func (s *Store) ListResources(filter map[string]string) ([]domain.Resource, error) {
	query := `SELECT id, name, type, provider, region, size, composition, labels, status,
		status_message, created_at, updated_at, last_activity_at FROM resources`
	args := []any{}
	if provider := filter["provider"]; provider != "" {
		query += " WHERE provider = ?"
		args = append(args, provider)
	}
	query += " ORDER BY created_at DESC"
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list resources: %w", err)
	}
	defer rows.Close()
	var out []domain.Resource
	for rows.Next() {
		r, err := scanResource(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate resources: %w", err)
	}
	return out, nil
}

func (s *Store) DeleteResource(id string) error {
	if _, err := s.db.Exec(`DELETE FROM resources WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete resource %q: %w", id, err)
	}
	return nil
}

func (s *Store) CleanupMode() (domain.CleanupMode, error) {
	var mode string
	if err := s.db.QueryRow(`SELECT mode FROM cleanup_settings WHERE id = 1`).Scan(&mode); err != nil {
		return "", fmt.Errorf("read cleanup mode: %w", err)
	}
	return domain.CleanupMode(mode), nil
}

func (s *Store) SetCleanupMode(mode domain.CleanupMode) error {
	if mode != domain.CleanupEnabled && mode != domain.CleanupDisabled {
		return fmt.Errorf("invalid cleanup mode %q", mode)
	}
	if _, err := s.db.Exec(`UPDATE cleanup_settings SET mode = ? WHERE id = 1`, string(mode)); err != nil {
		return fmt.Errorf("set cleanup mode: %w", err)
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanResource(row rowScanner) (domain.Resource, error) {
	var r domain.Resource
	var labels, status, created, updated, lastActivity string
	if err := row.Scan(&r.ID, &r.Name, &r.Type, &r.Provider, &r.Region, &r.Size, &r.Composition,
		&labels, &status, &r.StatusMessage, &created, &updated, &lastActivity); err != nil {
		return domain.Resource{}, fmt.Errorf("scan resource: %w", err)
	}
	if err := json.Unmarshal([]byte(labels), &r.Labels); err != nil {
		return domain.Resource{}, fmt.Errorf("decode resource labels: %w", err)
	}
	r.Status = domain.ResourceStatus(status)
	var err error
	if r.CreatedAt, err = parseTime(created); err != nil {
		return domain.Resource{}, err
	}
	if r.UpdatedAt, err = parseTime(updated); err != nil {
		return domain.Resource{}, err
	}
	if r.LastActivityAt, err = parseTime(lastActivity); err != nil {
		return domain.Resource{}, err
	}
	return r, nil
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse timestamp: %w", err)
	}
	return t, nil
}
