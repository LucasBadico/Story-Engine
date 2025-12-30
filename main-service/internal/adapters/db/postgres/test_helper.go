//go:build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/story-engine/main-service/internal/platform/config"
	"github.com/story-engine/main-service/internal/platform/database"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) (*DB, func()) {
	cfg := config.Load()
	
	// Override with test database if specified
	if testDB := os.Getenv("TEST_DB_NAME"); testDB != "" {
		cfg.Database.Database = testDB
	}

	db, err := database.New(cfg.Database)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	pgDB := NewDB(db)

	// Run migrations (in a real scenario, you'd use a migration tool)
	// For now, we assume migrations are run separately

	cleanup := func() {
		db.Close()
	}

	return pgDB, cleanup
}

// TruncateTables truncates all tables for test cleanup
func TruncateTables(ctx context.Context, db *DB) error {
	tables := []string{
		"prose_blocks",
		"beats",
		"scenes",
		"chapters",
		"stories",
		"audit_logs",
		"memberships",
		"users",
		"tenants",
	}

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		if _, err := db.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to truncate %s: %w", table, err)
		}
	}

	return nil
}

