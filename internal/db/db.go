package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Open opens a SQLite database, sets pragmas, and runs migrations.
func Open(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Set a single connection — SQLite doesn't benefit from a pool
	db.SetMaxOpenConns(1)

	if err := setPragmas(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("set pragmas: %w", err)
	}

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return db, nil
}

func setPragmas(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA foreign_keys = ON",
		"PRAGMA cache_size = -64000",
		"PRAGMA temp_store = MEMORY",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
	}
	return nil
}

func runMigrations(db *sql.DB) error {
	var currentVersion int
	err := db.QueryRow("PRAGMA user_version").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("get user_version: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	// Filter to .sql files only
	var sqlEntries []fs.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			sqlEntries = append(sqlEntries, entry)
		}
	}

	for i, entry := range sqlEntries {
		version := i + 1
		if version <= currentVersion {
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		upSQL := extractUpSection(string(content))
		if upSQL == "" {
			return fmt.Errorf("migration %s: no -- up section found", entry.Name())
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin transaction for migration %d: %w", version, err)
		}

		if _, err := tx.Exec(upSQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("execute migration %s: %w", entry.Name(), err)
		}

		if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", version)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("set user_version to %d: %w", version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", version, err)
		}
	}

	return nil
}

func extractUpSection(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inUp := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "-- up" {
			inUp = true
			continue
		}
		if trimmed == "-- down" {
			break
		}
		if inUp {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
