package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
)

// FactionRepository defines the interface for faction persistence
type FactionRepository interface {
	Create(ctx context.Context, f *world.Faction) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.Faction, error)
	ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID) ([]*world.Faction, error)
	Update(ctx context.Context, f *world.Faction) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	GetChildren(ctx context.Context, tenantID, parentID uuid.UUID) ([]*world.Faction, error)
}

