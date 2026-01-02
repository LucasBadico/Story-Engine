package story

import (
	"time"

	"github.com/google/uuid"
)

// Scene represents a scene entity
type Scene struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	StoryID        uuid.UUID  `json:"story_id"`
	ChapterID      *uuid.UUID `json:"chapter_id,omitempty"` // nullable
	OrderNum       int        `json:"order_num"`
	POVCharacterID *uuid.UUID `json:"pov_character_id,omitempty"` // nullable
	TimeRef        string     `json:"time_ref"`
	Goal           string     `json:"goal"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// NewScene creates a new scene
func NewScene(tenantID, storyID uuid.UUID, chapterID *uuid.UUID, orderNum int) (*Scene, error) {
	if orderNum < 1 {
		return nil, ErrInvalidOrderNumber
	}

	now := time.Now()
	return &Scene{
		ID:        uuid.New(),
		TenantID:  tenantID,
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

// UpdateChapter updates the chapter
func (s *Scene) UpdateChapter(chapterID *uuid.UUID) {
	s.ChapterID = chapterID
	s.UpdatedAt = time.Now()
}
