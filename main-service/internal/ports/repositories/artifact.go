package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
)

// ArtifactRepository defines the interface for artifact persistence
type ArtifactRepository interface {
	Create(ctx context.Context, a *world.Artifact) error
	GetByID(ctx context.Context, id uuid.UUID) (*world.Artifact, error)
	ListByWorld(ctx context.Context, worldID uuid.UUID, limit, offset int) ([]*world.Artifact, error)
	ListByCharacter(ctx context.Context, characterID uuid.UUID) ([]*world.Artifact, error)
	ListByLocation(ctx context.Context, locationID uuid.UUID) ([]*world.Artifact, error)
	Update(ctx context.Context, a *world.Artifact) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByWorld(ctx context.Context, worldID uuid.UUID) (int, error)
}

