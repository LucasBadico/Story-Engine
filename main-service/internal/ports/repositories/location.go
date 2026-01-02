package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
)

// LocationRepository defines the interface for location persistence
type LocationRepository interface {
	Create(ctx context.Context, l *world.Location) error
	GetByID(ctx context.Context, id uuid.UUID) (*world.Location, error)
	ListByWorld(ctx context.Context, worldID uuid.UUID, limit, offset int) ([]*world.Location, error)
	ListByWorldTree(ctx context.Context, worldID uuid.UUID) ([]*world.Location, error)
	GetChildren(ctx context.Context, locationID uuid.UUID) ([]*world.Location, error)
	GetAncestors(ctx context.Context, locationID uuid.UUID) ([]*world.Location, error)
	GetDescendants(ctx context.Context, locationID uuid.UUID) ([]*world.Location, error)
	Update(ctx context.Context, l *world.Location) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByWorld(ctx context.Context, worldID uuid.UUID) (int, error)
}


