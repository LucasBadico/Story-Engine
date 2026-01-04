package sqlite

import (
	"context"
	"database/sql"

	"github.com/story-engine/main-service/internal/platform/database"
)

// DB wraps the SQLite database connection
type DB struct {
	db *sql.DB
}

// NewDB creates a new DB instance from a database.SQLiteDB
func NewDB(sqliteDB *database.SQLiteDB) *DB {
	return &DB{db: sqliteDB.DB()}
}

// Begin starts a transaction
func (db *DB) Begin(ctx context.Context) (*sql.Tx, error) {
	return db.db.BeginTx(ctx, nil)
}

// Query executes a query that returns rows
func (db *DB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (db *DB) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return db.db.QueryRowContext(ctx, query, args...)
}

// Exec executes a query that doesn't return rows
func (db *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}

