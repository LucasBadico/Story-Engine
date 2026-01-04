package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/story-engine/main-service/internal/platform/config"
)

// SQLiteDB wraps sql.DB for SQLite database operations
type SQLiteDB struct {
	db *sql.DB
}

// NewSQLite creates a new SQLite database connection
func NewSQLite(cfg config.DatabaseConfig) (*SQLiteDB, error) {
	// Ensure the directory exists
	path := cfg.Path
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	// Open SQLite database with foreign keys enabled
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=1")
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.ExecContext(context.Background(), "PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Test connection
	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	return &SQLiteDB{db: db}, nil
}

// Close closes the database connection
func (db *SQLiteDB) Close() error {
	return db.db.Close()
}

// DB returns the underlying sql.DB
func (db *SQLiteDB) DB() *sql.DB {
	return db.db
}

// Begin starts a new transaction
func (db *SQLiteDB) Begin(ctx context.Context) (*sql.Tx, error) {
	return db.db.BeginTx(ctx, nil)
}

