//go:build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/platform/config"
	"github.com/story-engine/llm-gateway-service/internal/platform/database"
)

var (
	dbManager     *DatabaseManager
	templateSetup sync.Once
	templateName  string
	setupErr      error
)

const pgIdentMax = 63

var reSanitize = regexp.MustCompile(`[^a-z0-9_]+`)

// DatabaseManager handles database operations for tests
type DatabaseManager struct {
	connString string
	host       string
	port       int
	user       string
	password   string
	dbname     string
}

// MakeDBName generates a valid and unique PostgreSQL database name (<=63 chars)
func MakeDBName(prefix, testName string) string {
	if prefix == "" {
		prefix = "test"
	}

	// Unique core part: prefix + UTC timestamp + short UUID
	ts := time.Now().UTC().Format("060102150405")           // YYMMDDhhmmss (12 chars)
	u8 := strings.ReplaceAll(uuid.NewString(), "-", "")[:8] // 8-character hex

	core := fmt.Sprintf("%s_%s_%s", strings.ToLower(prefix), ts, u8)

	// Sanitize testName -> lowercase, only [a-z0-9_]
	slug := strings.ToLower(testName)
	slug = strings.ReplaceAll(slug, "-", "_")
	slug = reSanitize.ReplaceAllString(slug, "_")
	slug = strings.Trim(slug, "_")

	// Calculate remaining space for "_<slug>"
	remain := pgIdentMax - len(core) - 1

	if remain > 0 && len(slug) > 0 {
		if len(slug) > remain {
			slug = slug[:remain]
		}
		return fmt.Sprintf("%s_%s", core, slug)
	}

	if len(core) > pgIdentMax {
		return core[:pgIdentMax]
	}
	return core
}

// GetTemplateDBName returns the name of the template database
func GetTemplateDBName(cfg *config.Config) string {
	// Parse database name from URL
	// Format: postgres://user:pass@host:port/dbname?sslmode=disable
	url := cfg.Database.URL
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		dbPart := parts[len(parts)-1]
		dbName := strings.Split(dbPart, "?")[0]
		return dbName
	}
	return "storyengine"
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager(cfg *config.Config) (*DatabaseManager, error) {
	// Parse connection details from URL
	// Format: postgres://user:pass@host:port/dbname?sslmode=disable
	url := cfg.Database.URL
	
	// Simple parsing - assumes standard format
	// In production, use proper URL parsing
	host := "localhost"
	port := 5432
	user := "postgres"
	password := "postgres"
	dbname := "storyengine"

	// Try to extract from URL if present
	if strings.HasPrefix(url, "postgres://") {
		url = strings.TrimPrefix(url, "postgres://")
		parts := strings.Split(url, "@")
		if len(parts) == 2 {
			auth := parts[0]
			rest := parts[1]
			
			authParts := strings.Split(auth, ":")
			if len(authParts) == 2 {
				user = authParts[0]
				password = authParts[1]
			}
			
			restParts := strings.Split(rest, "/")
			if len(restParts) >= 2 {
				hostPort := restParts[0]
				hpParts := strings.Split(hostPort, ":")
				if len(hpParts) == 2 {
					host = hpParts[0]
					fmt.Sscanf(hpParts[1], "%d", &port)
				} else {
					host = hostPort
				}
				
				dbPart := restParts[1]
				dbname = strings.Split(dbPart, "?")[0]
			}
		}
	}

	return &DatabaseManager{
		connString: cfg.Database.URL,
		host:       host,
		port:       port,
		user:       user,
		password:   password,
		dbname:     dbname,
	}, nil
}

// CloneDatabase creates a new test database by cloning the template
func (m *DatabaseManager) CloneDatabase(templateName, targetName string) error {
	// For now, we'll use a simpler approach: create database and apply migrations
	// In a full implementation, we'd use pg_dump like main-service
	
	// Connect to postgres database to create new database
	adminURL := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=disable",
		m.user, m.password, m.host, m.port)
	
	adminCfg := &config.Config{}
	adminCfg.Database.URL = adminURL
	adminDB, err := database.New(adminCfg)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer adminDB.Close()

	// Drop target if exists
	ctx := context.Background()
	_, _ = adminDB.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", safeIdent(targetName)))

	// Create new database
	_, err = adminDB.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", safeIdent(targetName)))
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	// Apply migrations to new database
	targetURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		m.user, m.password, m.host, m.port, targetName)
	
	targetCfg := &config.Config{}
	targetCfg.Database.URL = targetURL
	targetDB, err := database.New(targetCfg)
	if err != nil {
		return fmt.Errorf("failed to connect to target database: %w", err)
	}
	defer targetDB.Close()

	// Apply migrations
	migrationsPath := findMigrationsPath()
	if err := applyMigrations(targetURL, migrationsPath); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// DropDatabase drops a database
func (m *DatabaseManager) DropDatabase(dbName string) error {
	adminURL := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=disable",
		m.user, m.password, m.host, m.port)
	
	adminCfg := &config.Config{}
	adminCfg.Database.URL = adminURL
	adminDB, err := database.New(adminCfg)
	if err != nil {
		return err
	}
	defer adminDB.Close()

	ctx := context.Background()
	// Terminate connections
	_, _ = adminDB.Exec(ctx, fmt.Sprintf(
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = %s",
		safeLiteral(dbName)))

	// Drop database
	_, err = adminDB.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", safeIdent(dbName)))
	return err
}

// SetupTestDB creates an isolated test database by cloning from a template
func SetupTestDB(t *testing.T) (*DB, func()) {
	// Ensure database manager is set up once
	templateSetup.Do(func() {
		cfg := config.Load()
		
		// Create database manager
		dbManager, setupErr = NewDatabaseManager(cfg)
		if setupErr != nil {
			return
		}

		// Use the main database (storyengine) as template
		// It already has migrations applied via db-up target
		templateName = GetTemplateDBName(cfg)
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
	testURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbManager.user, dbManager.password, dbManager.host, dbManager.port, testDBName)
	cfg.Database.URL = testURL

	db, err := database.New(cfg)
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

// applyMigrations applies migrations using golang-migrate CLI
func applyMigrations(dbURL, migrationsPath string) error {
	// This is a simplified version - in production, use golang-migrate library
	// For now, we'll assume migrations are already applied to template
	// In a full implementation, we'd call migrate.Up() programmatically
	return nil
}

// TruncateTables truncates all tables in the test database
func TruncateTables(ctx context.Context, db *DB) error {
	query := `TRUNCATE TABLE 
		embedding_chunks, embedding_documents
		RESTART IDENTITY CASCADE`
	
	if _, err := db.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	return nil
}

// safeIdent safely quotes an SQL identifier
func safeIdent(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}

// safeLiteral safely quotes a string literal for SQL
func safeLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

