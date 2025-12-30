package story

import (
	"time"

	"github.com/google/uuid"
)

// Scene represents a scene entity
type Scene struct {
	ID              uuid.UUID
	StoryID         uuid.UUID
	ChapterID       uuid.UUID
	OrderNum        int
	POVCharacterID  *uuid.UUID // nullable
	LocationID      *uuid.UUID // nullable
	TimeRef         string
	Goal            string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewScene creates a new scene
func NewScene(storyID, chapterID uuid.UUID, orderNum int) (*Scene, error) {
	if orderNum < 1 {
		return nil, ErrInvalidOrderNumber
	}

	now := time.Now()
	return &Scene{
		ID:        uuid.New(),
		StoryID:   storyID,
		ChapterID: chapterID,
		OrderNum:  orderNum,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Validate validates the scene entity
func (s *Scene) Validate() error {
	if s.OrderNum < 1 {
		return ErrInvalidOrderNumber
	}
	return nil
}

// UpdateGoal updates the scene goal
func (s *Scene) UpdateGoal(goal string) {
	s.Goal = goal
	s.UpdatedAt = time.Now()
}

// UpdatePOV updates the POV character
func (s *Scene) UpdatePOV(characterID *uuid.UUID) {
	s.POVCharacterID = characterID
	s.UpdatedAt = time.Now()
}

// UpdateLocation updates the location
func (s *Scene) UpdateLocation(locationID *uuid.UUID) {
	s.LocationID = locationID
	s.UpdatedAt = time.Now()
}

