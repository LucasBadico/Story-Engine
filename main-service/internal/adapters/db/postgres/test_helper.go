//go:build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/story-engine/main-service/internal/platform/config"
	"github.com/story-engine/main-service/internal/platform/database"
)

var (
	dbManager     *DatabaseManager
	templateSetup sync.Once
	templateName  string
	setupErr      error
)

// SetupTestDB creates an isolated test database by cloning from a template
func SetupTestDB(t *testing.T) (*DB, func()) {
	// Ensure database manager is set up once
	templateSetup.Do(func() {
		cfg := config.Load()
		
		// Create database manager
		dbManager, setupErr = NewDatabaseManager(cfg.Database)
		if setupErr != nil {
			return
		}

		// Use the main database (storyengine) as template
		// It already has migrations applied via db-up target
		templateName = GetTemplateDBName(cfg.Database)
	})

	if setupErr != nil {
		t.Fatalf("failed to setup template database: %v", setupErr)
	}

	// Generate unique database name for this test
	testDBName := MakeDBName("test", t.Name())

	// Clone from template
	if err := dbManager.CloneDatabase(templateName, testDBName); err != nil {
		t.Fatalf("failed to clone test database: %v", err)
	}

	// Connect to the test database
	cfg := config.Load()
	cfg.Database.Host = dbManager.host
	cfg.Database.Port = dbManager.port
	cfg.Database.User = dbManager.user
	cfg.Database.Password = dbManager.password
	cfg.Database.Database = testDBName

	db, err := database.New(cfg.Database)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	pgDB := NewDB(db)

	cleanup := func() {
		db.Close()
		// Drop the test database after test completes
		if err := dbManager.DropDatabase(testDBName); err != nil {
			t.Logf("warning: failed to drop test database %s: %v", testDBName, err)
		}
	}

	return pgDB, cleanup
}

// findMigrationsPath attempts to locate the migrations directory
func findMigrationsPath() string {
	// Try common paths
	paths := []string{
		"migrations",
		"../migrations",
		"../../migrations",
		"../../../migrations",
		"../../../../migrations",
	}

	currentDir, _ := os.Getwd()
	
	for _, p := range paths {
		fullPath := filepath.Join(currentDir, p)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
	}

	// Fallback to relative path
	return "migrations"
}

// TruncateTables truncates all tables in the test database
// NOTE: With database cloning, this is typically not needed as each test gets a fresh database
func TruncateTables(ctx context.Context, db *DB) error {
	query := `TRUNCATE TABLE 
		content_blocks, beats, scenes, chapters, stories, 
		audit_logs, memberships, users, tenants 
		RESTART IDENTITY CASCADE`
	
	if _, err := db.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	return nil
}

// CleanupTemplateDB cleans up the template database (call this in TestMain if needed)
func CleanupTemplateDB() {
	if dbManager != nil {
		if templateName != "" {
			_ = dbManager.DropDatabase(templateName)
		}
		_ = dbManager.Close()
	}
}
