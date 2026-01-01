package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
)

// WorldRepository defines the interface for world persistence
type WorldRepository interface {
	Create(ctx context.Context, w *world.World) error
	GetByID(ctx context.Context, id uuid.UUID) (*world.World, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*world.World, error)
	Update(ctx context.Context, w *world.World) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
}

