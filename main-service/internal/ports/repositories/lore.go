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

// LoreReferenceRepository defines the interface for lore-reference relationships
type LoreReferenceRepository interface {
	Create(ctx context.Context, lr *world.LoreReference) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.LoreReference, error)
	ListByLore(ctx context.Context, tenantID, loreID uuid.UUID) ([]*world.LoreReference, error)
	ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) ([]*world.LoreReference, error)
	Update(ctx context.Context, lr *world.LoreReference) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	DeleteByLoreAndEntity(ctx context.Context, tenantID, loreID uuid.UUID, entityType string, entityID uuid.UUID) error
	DeleteByLore(ctx context.Context, tenantID, loreID uuid.UUID) error
}

