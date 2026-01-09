package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
)

// EventRepository defines the interface for event persistence
type EventRepository interface {
	Create(ctx context.Context, e *world.Event) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.Event, error)
	ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID) ([]*world.Event, error)
	Update(ctx context.Context, e *world.Event) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	// Hierarquia (causalidade)
	GetChildren(ctx context.Context, tenantID, parentID uuid.UUID) ([]*world.Event, error)
	GetAncestors(ctx context.Context, tenantID, eventID uuid.UUID) ([]*world.Event, error)
	GetDescendants(ctx context.Context, tenantID, eventID uuid.UUID) ([]*world.Event, error)
	// Timeline
	GetEpoch(ctx context.Context, tenantID, worldID uuid.UUID) (*world.Event, error)
	ListByTimeline(ctx context.Context, tenantID, worldID uuid.UUID, fromPos, toPos *float64) ([]*world.Event, error)
}

