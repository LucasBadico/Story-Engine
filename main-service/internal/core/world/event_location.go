package world

import (
	"time"

	"github.com/google/uuid"
)

// EventLocation represents the relationship between an event and a location
type EventLocation struct {
	ID           uuid.UUID  `json:"id"`
	EventID       uuid.UUID  `json:"event_id"`
	LocationID    uuid.UUID  `json:"location_id"`
	Significance *string    `json:"significance,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// NewEventLocation creates a new event-location relationship
func NewEventLocation(eventID, locationID uuid.UUID, significance *string) *EventLocation {
	return &EventLocation{
		ID:           uuid.New(),
		EventID:       eventID,
		LocationID:    locationID,
		Significance:  significance,
		CreatedAt:     time.Now(),
	}
}


