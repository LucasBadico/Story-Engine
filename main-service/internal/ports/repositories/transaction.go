package repositories

import (
	"context"
)

// Tx represents a database transaction (generic, not database-specific)
type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// TransactionRepository defines the interface for transaction management
type TransactionRepository interface {
	BeginTx(ctx context.Context) (Tx, error)
	WithTx(ctx context.Context, fn func(Tx) error) error
}

