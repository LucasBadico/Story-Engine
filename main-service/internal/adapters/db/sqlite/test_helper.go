//go:build integration

package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// SetupTestSQLiteDB creates an isolated in-memory SQLite database for testing
// Returns the DB wrapper and a cleanup function
func SetupTestSQLiteDB(t *testing.T) (*DB, func()) {
	// Use in-memory database for tests (fast and isolated)
	// For file-based tests, we could use a temporary file instead
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=1")
	if err != nil {
		t.Fatalf("failed to open SQLite database: %v", err)
	}

	// Enable foreign keys
	if _, err := db.ExecContext(context.Background(), "PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	// Apply migrations
	if err := ApplyMigrations(db); err != nil {
		db.Close()
		t.Fatalf("failed to apply migrations: %v", err)
	}

	// Create DB wrapper directly
	wrapper := NewDBFromSQL(db)

	cleanup := func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close database: %v", err)
		}
	}

	return wrapper, cleanup
}

// SetupTestDBFile creates a temporary file-based SQLite database for testing
// Useful for tests that need to persist data across connections
func SetupTestDBFile(t *testing.T) (*DB, func(), string) {
	// Create temporary file
	tmpfile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpfile.Close()

	dbPath := tmpfile.Name()

	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=1")
	if err != nil {
		os.Remove(dbPath)
		t.Fatalf("failed to open SQLite database: %v", err)
	}

	// Enable foreign keys
	if _, err := db.ExecContext(context.Background(), "PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		os.Remove(dbPath)
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	// Apply migrations
	if err := ApplyMigrations(db); err != nil {
		db.Close()
		os.Remove(dbPath)
		t.Fatalf("failed to apply migrations: %v", err)
	}

	wrapper := NewDBFromSQL(db)

	cleanup := func() {
		db.Close()
		if err := os.Remove(dbPath); err != nil {
			t.Logf("warning: failed to remove temp file: %v", err)
		}
	}

	return wrapper, cleanup, dbPath
}

// NewDBFromSQL creates a DB wrapper from a raw sql.DB
// This is a helper function for tests
func NewDBFromSQL(db *sql.DB) *DB {
	return &DB{db: db}
}

// TruncateTables truncates all tables in the test database
// For SQLite, we use DELETE instead of TRUNCATE (SQLite doesn't have TRUNCATE)
func TruncateTables(ctx context.Context, db *DB) error {
	// Get all table names
	tables := []string{
		"scene_references",
		"content_block_references",
		"beats",
		"scenes",
		"chapters",
		"stories",
		"lore_references",
		"lores",
		"faction_references",
		"factions",
		"archetype_traits",
		"archetypes",
		"traits",
		"event_artifacts",
		"event_locations",
		"event_characters",
		"events",
		"artifact_references",
		"artifacts",
		"character_traits",
		"characters",
		"locations",
		"worlds",
		"tenants",
	}

	// Disable foreign keys temporarily for truncation
	if _, err := db.Exec(ctx, "PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("failed to disable foreign keys: %w", err)
	}

	// Delete all data from tables (in reverse order of dependencies)
	for _, table := range tables {
		if _, err := db.Exec(ctx, fmt.Sprintf("DELETE FROM %s", table)); err != nil {
			// Re-enable foreign keys even if there's an error
			_, _ = db.Exec(ctx, "PRAGMA foreign_keys = ON")
			return fmt.Errorf("failed to delete from %s: %w", table, err)
		}
	}

	// Re-enable foreign keys
	if _, err := db.Exec(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return nil
}
