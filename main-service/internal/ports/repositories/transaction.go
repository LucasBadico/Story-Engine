package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// TransactionRepository defines the interface for transaction management
type TransactionRepository interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)
	WithTx(ctx context.Context, fn func(pgx.Tx) error) error
}

