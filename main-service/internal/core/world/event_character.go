package world

import (
	"time"

	"github.com/google/uuid"
)

// EventCharacter represents the relationship between an event and a character
type EventCharacter struct {
	ID          uuid.UUID  `json:"id"`
	EventID     uuid.UUID  `json:"event_id"`
	CharacterID uuid.UUID  `json:"character_id"`
	Role        *string    `json:"role,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// NewEventCharacter creates a new event-character relationship
func NewEventCharacter(eventID, characterID uuid.UUID, role *string) *EventCharacter {
	return &EventCharacter{
		ID:          uuid.New(),
		EventID:     eventID,
		CharacterID: characterID,
		Role:        role,
		CreatedAt:   time.Now(),
	}
}


