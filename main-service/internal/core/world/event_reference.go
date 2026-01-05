package world

import (
	"time"

	"github.com/google/uuid"
)

// EventReference represents the relationship between an event and another entity
type EventReference struct {
	ID               uuid.UUID  `json:"id"`
	EventID          uuid.UUID  `json:"event_id"`
	EntityType       string     `json:"entity_type"`
	EntityID         uuid.UUID  `json:"entity_id"`
	RelationshipType *string    `json:"relationship_type,omitempty"` // "role" para character/artifact, "significance" para location
	Notes            string     `json:"notes"`
	CreatedAt        time.Time  `json:"created_at"`
}

// NewEventReference creates a new event-reference relationship
func NewEventReference(eventID uuid.UUID, entityType string, entityID uuid.UUID, relationshipType *string) *EventReference {
	return &EventReference{
		ID:               uuid.New(),
		EventID:          eventID,
		EntityType:       entityType,
		EntityID:         entityID,
		RelationshipType: relationshipType,
		CreatedAt:        time.Now(),
	}
}

// UpdateRelationshipType updates the relationship type
func (er *EventReference) UpdateRelationshipType(relationshipType *string) {
	er.RelationshipType = relationshipType
}

// UpdateNotes updates the notes
func (er *EventReference) UpdateNotes(notes string) {
	er.Notes = notes
}

