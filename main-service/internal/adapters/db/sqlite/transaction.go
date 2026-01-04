package sqlite

import (
	"context"
	"database/sql"

	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.TransactionRepository = (*TransactionRepository)(nil)
var _ repositories.Tx = (*txWrapper)(nil)

// txWrapper wraps sql.Tx to implement the generic Tx interface
type txWrapper struct {
	tx *sql.Tx
}

// Commit commits the transaction
func (w *txWrapper) Commit(ctx context.Context) error {
	return w.tx.Commit()
}

// Rollback rolls back the transaction
func (w *txWrapper) Rollback(ctx context.Context) error {
	return w.tx.Rollback()
}

// TransactionRepository implements transaction management for SQLite
type TransactionRepository struct {
	db *DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// BeginTx starts a new transaction
func (r *TransactionRepository) BeginTx(ctx context.Context) (repositories.Tx, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &txWrapper{tx: tx}, nil
}

// WithTx executes a function within a transaction
func (r *TransactionRepository) WithTx(ctx context.Context, fn func(repositories.Tx) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	wrapper := &txWrapper{tx: tx}

	defer func() {
		if p := recover(); p != nil {
			_ = wrapper.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = wrapper.Rollback(ctx)
		} else {
			err = wrapper.Commit(ctx)
		}
	}()

	err = fn(wrapper)
	return err
}

