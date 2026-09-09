// Package database opens the SQLite connection and keeps its schema current.
package database

import (
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"time"

	_ "modernc.org/sqlite"

	"velostats/internal/support"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Open connects to the SQLite database at the given path, creating the file and
// its directory when they do not exist yet.
//
// Writes are serialised onto a single connection: SQLite allows only one writer
// at a time, and the API, the scheduler and every worker share one file.
func Open(path string) (*sql.DB, error) {
	if directory := filepath.Dir(path); directory != "." {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	dsn := fmt.Sprintf(
		"file:%s?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)",
		url.PathEscape(path),
	)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(0)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	return db, nil
}

// Migrate applies every migration that has not run yet.
func Migrate(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
		migration VARCHAR NOT NULL,
		ran_at DATETIME NOT NULL
	)`); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		applied, err := hasRun(db, name)
		if err != nil {
			return err
		}

		if applied {
			continue
		}

		statements, err := migrations.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		if _, err := db.Exec(string(statements)); err != nil {
			return fmt.Errorf("run migration %s: %w", name, err)
		}

		if _, err := db.Exec(
			`INSERT INTO migrations (migration, ran_at) VALUES (?, ?)`,
			name, support.DatabaseDateTime(time.Now()),
		); err != nil {
			return fmt.Errorf("record migration %s: %w", name, err)
		}
	}

	return nil
}

func hasRun(db *sql.DB, name string) (bool, error) {
	var count int

	if err := db.QueryRow(`SELECT COUNT(*) FROM migrations WHERE migration = ?`, name).Scan(&count); err != nil {
		return false, fmt.Errorf("check migration %s: %w", name, err)
	}

	return count > 0, nil
}
