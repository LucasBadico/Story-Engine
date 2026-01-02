package world

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEventNameRequired = errors.New("event name is required")
	ErrInvalidImportance = errors.New("importance must be between 1 and 10")
)

// Event represents an event in a world
type Event struct {
	ID          uuid.UUID  `json:"id"`
	WorldID     uuid.UUID  `json:"world_id"`
	Name        string     `json:"name"`
	Type        *string    `json:"type,omitempty"`
	Description *string    `json:"description,omitempty"`
	Timeline    *string    `json:"timeline,omitempty"`
	Importance  int        `json:"importance"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// NewEvent creates a new event
func NewEvent(worldID uuid.UUID, name string) (*Event, error) {
	if name == "" {
		return nil, ErrEventNameRequired
	}

	return &Event{
		ID:         uuid.New(),
		WorldID:    worldID,
		Name:       name,
		Importance: 5,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}

// Validate validates the event entity
func (e *Event) Validate() error {
	if e.Name == "" {
		return ErrEventNameRequired
	}
	if e.Importance < 1 || e.Importance > 10 {
		return ErrInvalidImportance
	}
	return nil
}

// UpdateName updates the event name
func (e *Event) UpdateName(name string) error {
	if name == "" {
		return ErrEventNameRequired
	}
	e.Name = name
	e.UpdatedAt = time.Now()
	return nil
}

// UpdateType updates the event type
func (e *Event) UpdateType(eventType *string) {
	e.Type = eventType
	e.UpdatedAt = time.Now()
}

// UpdateDescription updates the event description
func (e *Event) UpdateDescription(description *string) {
	e.Description = description
	e.UpdatedAt = time.Now()
}

// UpdateTimeline updates the event timeline
func (e *Event) UpdateTimeline(timeline *string) {
	e.Timeline = timeline
	e.UpdatedAt = time.Now()
}

// UpdateImportance updates the event importance
func (e *Event) UpdateImportance(importance int) error {
	if importance < 1 || importance > 10 {
		return ErrInvalidImportance
	}
	e.Importance = importance
	e.UpdatedAt = time.Now()
	return nil
}


