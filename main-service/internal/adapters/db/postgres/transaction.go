package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.TransactionRepository = (*TransactionRepository)(nil)

// TransactionRepository implements transaction management
type TransactionRepository struct {
	db *DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// BeginTx starts a new transaction
func (r *TransactionRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}

// WithTx executes a function within a transaction
func (r *TransactionRepository) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	err = fn(tx)
	return err
}

