package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type HotSearch struct {
	ID        int64
	Title     string
	URL       string
	Platform  string
	Rank      int
	Heat      int64
	Category  string
	CreatedAt time.Time
}

type Storage struct {
	db *sql.DB
}

func Open(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	schema := `
CREATE TABLE IF NOT EXISTS hot_searches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    url TEXT,
    platform TEXT NOT NULL,
    rank INTEGER NOT NULL,
    heat INTEGER DEFAULT 0,
    category TEXT,
    created_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_platform ON hot_searches(platform);
CREATE INDEX IF NOT EXISTS idx_created_at ON hot_searches(created_at);
`
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Storage) Save(items []HotSearch) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO hot_searches (title, url, platform, rank, heat, category, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err := stmt.Exec(
			item.Title,
			item.URL,
			item.Platform,
			item.Rank,
			item.Heat,
			item.Category,
			item.CreatedAt,
		); err != nil {
			return fmt.Errorf("insert item %q: %w", item.Title, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (s *Storage) ListAll() ([]HotSearch, error) {
	rows, err := s.db.Query(`
		SELECT id, title, url, platform, rank, heat, category, created_at
		FROM hot_searches
		ORDER BY rank ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query all: %w", err)
	}
	defer rows.Close()

	return scanHotSearches(rows)
}

func (s *Storage) ListByPlatform(platform string) ([]HotSearch, error) {
	rows, err := s.db.Query(`
		SELECT id, title, url, platform, rank, heat, category, created_at
		FROM hot_searches
		WHERE platform = ?
		ORDER BY rank ASC
	`, platform)
	if err != nil {
		return nil, fmt.Errorf("query by platform: %w", err)
	}
	defer rows.Close()

	return scanHotSearches(rows)
}

func (s *Storage) DeleteBefore(t time.Time) error {
	_, err := s.db.Exec(`
		DELETE FROM hot_searches WHERE created_at < ?
	`, t)
	if err != nil {
		return fmt.Errorf("delete before: %w", err)
	}
	return nil
}

func scanHotSearches(rows *sql.Rows) ([]HotSearch, error) {
	var items []HotSearch
	for rows.Next() {
		var item HotSearch
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.URL,
			&item.Platform,
			&item.Rank,
			&item.Heat,
			&item.Category,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return items, nil
}
