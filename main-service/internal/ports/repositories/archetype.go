package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
)

// ArchetypeRepository defines the interface for archetype persistence
type ArchetypeRepository interface {
	Create(ctx context.Context, a *world.Archetype) error
	GetByID(ctx context.Context, id uuid.UUID) (*world.Archetype, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*world.Archetype, error)
	Update(ctx context.Context, a *world.Archetype) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
}

// ArchetypeTraitRepository defines the interface for archetype-trait junction persistence
type ArchetypeTraitRepository interface {
	Create(ctx context.Context, at *world.ArchetypeTrait) error
	GetByArchetype(ctx context.Context, archetypeID uuid.UUID) ([]*world.ArchetypeTrait, error)
	GetByTrait(ctx context.Context, traitID uuid.UUID) ([]*world.ArchetypeTrait, error)
	Delete(ctx context.Context, archetypeID, traitID uuid.UUID) error
	DeleteByArchetype(ctx context.Context, archetypeID uuid.UUID) error
}

