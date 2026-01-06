package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// isIgnorableError checks if the error is something we can safely ignore
// (e.g., duplicate column, index already exists, table already exists)
func isIgnorableError(err error) bool {
	if err == nil {
		return true
	}
	errMsg := strings.ToLower(err.Error())
	ignorableErrors := []string{
		"duplicate column name",
		"already exists",
		"table already exists",
		"index already exists",
	}
	for _, ie := range ignorableErrors {
		if strings.Contains(errMsg, ie) {
			return true
		}
	}
	return false
}

// ApplyMigrations applies all SQLite migrations to the database
// This function can be used at runtime to ensure the database schema is up to date
func ApplyMigrations(db *sql.DB) error {
	migrationsPath := findMigrationsPath()
	migrationsDir := filepath.Join(migrationsPath, "migrations")

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort migration files by name (they should be numbered)
	var migrationFiles []string
	for _, entry := range files {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Apply each migration
	for _, filename := range migrationFiles {
		filepath := filepath.Join(migrationsDir, filename)
		content, err := os.ReadFile(filepath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Execute migration statements individually to handle idempotent operations
		statements := splitSQLStatements(string(content))
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.ExecContext(context.Background(), stmt); err != nil {
				// Ignore errors for duplicate columns, existing indexes, etc.
				if !isIgnorableError(err) {
					return fmt.Errorf("failed to execute migration %s: %w", filename, err)
				}
			}
		}
	}

	return nil
}

// splitSQLStatements splits SQL content into individual statements
func splitSQLStatements(content string) []string {
	var statements []string
	var current strings.Builder
	
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		
		current.WriteString(line)
		current.WriteString("\n")
		
		// Check if this line ends a statement
		if strings.HasSuffix(trimmed, ";") {
			statements = append(statements, current.String())
			current.Reset()
		}
	}
	
	// Add any remaining content
	if current.Len() > 0 {
		statements = append(statements, current.String())
	}
	
	return statements
}

// findMigrationsPath attempts to locate the SQLite migrations directory
// It tries various paths relative to the current working directory
// Note: SQLite migrations are in internal/adapters/db/sqlite/migrations, NOT in main-service/migrations
func findMigrationsPath() string {
	// Try SQLite-specific paths first (NOT "." which would find PostgreSQL migrations)
	paths := []string{
		"internal/adapters/db/sqlite",
		"../internal/adapters/db/sqlite",
		"../../internal/adapters/db/sqlite",
		"../../../internal/adapters/db/sqlite",
		"../../../../internal/adapters/db/sqlite",
		"../../../../../internal/adapters/db/sqlite",
	}

	currentDir, _ := os.Getwd()

	for _, p := range paths {
		fullPath := filepath.Join(currentDir, p)
		migrationsPath := filepath.Join(fullPath, "migrations")
		if _, err := os.Stat(migrationsPath); err == nil {
			return fullPath
		}
	}

	// Fallback - try to find from executable location
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		paths := []string{
			filepath.Join(exeDir, "internal/adapters/db/sqlite"),
			filepath.Join(exeDir, "../internal/adapters/db/sqlite"),
			filepath.Join(exeDir, "../../internal/adapters/db/sqlite"),
		}
		for _, p := range paths {
			migrationsPath := filepath.Join(p, "migrations")
			if _, err := os.Stat(migrationsPath); err == nil {
				return p
			}
		}
	}

	// Final fallback
	return "internal/adapters/db/sqlite"
}
