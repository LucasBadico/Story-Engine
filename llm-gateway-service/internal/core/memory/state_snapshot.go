package memory

import (
	"time"

	"github.com/google/uuid"
)

// StateSnapshot represents a per-chapter entity state (hot layer)
type StateSnapshot struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	StoryID            uuid.UUID
	EntityID           uuid.UUID
	EntityType         string
	ChapterNumber      int
	TimelineDay        *int
	TimelineConfidence *string
	StateData          map[string]interface{}
	SceneID            *uuid.UUID
	CreatedAt          time.Time
}

// NewStateSnapshot creates a new state snapshot
func NewStateSnapshot(tenantID, storyID, entityID uuid.UUID, entityType string, chapterNumber int, stateData map[string]interface{}) *StateSnapshot {
	return &StateSnapshot{
		ID:            uuid.New(),
		TenantID:      tenantID,
		StoryID:       storyID,
		EntityID:      entityID,
		EntityType:    entityType,
		ChapterNumber: chapterNumber,
		StateData:     stateData,
		CreatedAt:     time.Now(),
	}
}
