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
}

// EventCharacterRepository defines the interface for event-character relationships
type EventCharacterRepository interface {
	Create(ctx context.Context, ec *world.EventCharacter) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.EventCharacter, error)
	ListByEvent(ctx context.Context, tenantID, eventID uuid.UUID) ([]*world.EventCharacter, error)
	ListByCharacter(ctx context.Context, tenantID, characterID uuid.UUID) ([]*world.EventCharacter, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	DeleteByEventAndCharacter(ctx context.Context, tenantID, eventID, characterID uuid.UUID) error
	DeleteByEvent(ctx context.Context, tenantID, eventID uuid.UUID) error
}

// EventLocationRepository defines the interface for event-location relationships
type EventLocationRepository interface {
	Create(ctx context.Context, el *world.EventLocation) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.EventLocation, error)
	ListByEvent(ctx context.Context, tenantID, eventID uuid.UUID) ([]*world.EventLocation, error)
	ListByLocation(ctx context.Context, tenantID, locationID uuid.UUID) ([]*world.EventLocation, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	DeleteByEventAndLocation(ctx context.Context, tenantID, eventID, locationID uuid.UUID) error
	DeleteByEvent(ctx context.Context, tenantID, eventID uuid.UUID) error
}

// EventArtifactRepository defines the interface for event-artifact relationships
type EventArtifactRepository interface {
	Create(ctx context.Context, ea *world.EventArtifact) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.EventArtifact, error)
	ListByEvent(ctx context.Context, tenantID, eventID uuid.UUID) ([]*world.EventArtifact, error)
	ListByArtifact(ctx context.Context, tenantID, artifactID uuid.UUID) ([]*world.EventArtifact, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	DeleteByEventAndArtifact(ctx context.Context, tenantID, eventID, artifactID uuid.UUID) error
	DeleteByEvent(ctx context.Context, tenantID, eventID uuid.UUID) error
}


