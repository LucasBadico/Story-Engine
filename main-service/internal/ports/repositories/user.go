package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/auth"
)

// UserRepository defines the interface for user persistence
type UserRepository interface {
	Create(ctx context.Context, u *auth.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*auth.User, error)
	GetByEmail(ctx context.Context, email string) (*auth.User, error)
	Update(ctx context.Context, u *auth.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*auth.User, error)
}

