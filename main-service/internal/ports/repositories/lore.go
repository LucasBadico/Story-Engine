package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
)

// LoreRepository defines the interface for lore persistence
type LoreRepository interface {
	Create(ctx context.Context, l *world.Lore) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.Lore, error)
	ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID) ([]*world.Lore, error)
	Update(ctx context.Context, l *world.Lore) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	GetChildren(ctx context.Context, tenantID, parentID uuid.UUID) ([]*world.Lore, error)
}

