package world

import (
	"time"

	"github.com/google/uuid"
)

// EventArtifact represents the relationship between an event and an artifact
type EventArtifact struct {
	ID         uuid.UUID  `json:"id"`
	EventID    uuid.UUID  `json:"event_id"`
	ArtifactID uuid.UUID  `json:"artifact_id"`
	Role       *string    `json:"role,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// NewEventArtifact creates a new event-artifact relationship
func NewEventArtifact(eventID, artifactID uuid.UUID, role *string) *EventArtifact {
	return &EventArtifact{
		ID:         uuid.New(),
		EventID:     eventID,
		ArtifactID:  artifactID,
		Role:        role,
		CreatedAt:   time.Now(),
	}
}

