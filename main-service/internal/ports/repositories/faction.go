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

// FactionReferenceRepository defines the interface for faction-reference relationships
type FactionReferenceRepository interface {
	Create(ctx context.Context, fr *world.FactionReference) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.FactionReference, error)
	ListByFaction(ctx context.Context, tenantID, factionID uuid.UUID) ([]*world.FactionReference, error)
	ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) ([]*world.FactionReference, error)
	Update(ctx context.Context, fr *world.FactionReference) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	DeleteByFactionAndEntity(ctx context.Context, tenantID, factionID uuid.UUID, entityType string, entityID uuid.UUID) error
	DeleteByFaction(ctx context.Context, tenantID, factionID uuid.UUID) error
}

