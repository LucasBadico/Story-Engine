package memory

import (
	"time"

	"github.com/google/uuid"
)

// Event represents an atomic narrative action in the timeline
type Event struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	StoryID            uuid.UUID
	ChapterNumber      int
	TimelineDay        *int
	TimelineConfidence *string
	SceneID            *uuid.UUID
	Description        string
	EntitiesInvolved   []string
	CreatedAt          time.Time
}

// NewEvent creates a new event
func NewEvent(tenantID, storyID uuid.UUID, chapterNumber int, description string) *Event {
	return &Event{
		ID:               uuid.New(),
		TenantID:         tenantID,
		StoryID:          storyID,
		ChapterNumber:    chapterNumber,
		Description:      description,
		EntitiesInvolved: []string{},
		CreatedAt:        time.Now(),
	}
}
