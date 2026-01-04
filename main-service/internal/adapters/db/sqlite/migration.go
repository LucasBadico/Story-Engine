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

		// Execute migration
		if _, err := db.ExecContext(context.Background(), string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}
	}

	return nil
}

// findMigrationsPath attempts to locate the migrations directory
// It tries various paths relative to the current working directory
func findMigrationsPath() string {
	// Try common paths relative to execution
	paths := []string{
		".",
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
