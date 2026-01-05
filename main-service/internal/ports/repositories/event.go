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

// EventReferenceRepository defines the interface for event-reference relationships
type EventReferenceRepository interface {
	Create(ctx context.Context, er *world.EventReference) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.EventReference, error)
	ListByEvent(ctx context.Context, tenantID, eventID uuid.UUID) ([]*world.EventReference, error)
	ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) ([]*world.EventReference, error)
	Update(ctx context.Context, er *world.EventReference) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	DeleteByEventAndEntity(ctx context.Context, tenantID, eventID uuid.UUID, entityType string, entityID uuid.UUID) error
	DeleteByEvent(ctx context.Context, tenantID, eventID uuid.UUID) error
}


