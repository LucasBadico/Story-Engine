package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
)

// TraitRepository defines the interface for trait persistence
type TraitRepository interface {
	Create(ctx context.Context, t *world.Trait) error
	GetByID(ctx context.Context, id uuid.UUID) (*world.Trait, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*world.Trait, error)
	Update(ctx context.Context, t *world.Trait) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
}


