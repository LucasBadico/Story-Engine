package story

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTimelineEventDescriptionRequired   = errors.New("description is required for timeline event")
	ErrTimelineEventChapterNumberRequired = errors.New("chapter_number is required for timeline event")
)

// TimelineEventEntity represents an entity linked to a timeline event
type TimelineEventEntity struct {
	ID              uuid.UUID
	TimelineEventID uuid.UUID
	Role            string
	EntityType      string
	EntityID        uuid.UUID
	CreatedAt       time.Time
}

// TimelineEvent represents a narrative event anchored to a chapter's timeline
type TimelineEvent struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	StoryID            uuid.UUID
	ChapterID          uuid.UUID
	ChapterNumber      int
	TimelineDay        *int
	TimelineConfidence *string
	Description        string
	Entities           []TimelineEventEntity
	CreatedAt          time.Time
}

// NewTimelineEvent creates a new timeline event
func NewTimelineEvent(tenantID, storyID uuid.UUID, chapterNumber int, description string) (*TimelineEvent, error) {
	if chapterNumber < 0 {
		return nil, ErrTimelineEventChapterNumberRequired
	}
	if description == "" {
		return nil, ErrTimelineEventDescriptionRequired
	}

	return &TimelineEvent{
		ID:            uuid.New(),
		TenantID:      tenantID,
		StoryID:       storyID,
		ChapterNumber: chapterNumber,
		Description:   description,
		Entities:      []TimelineEventEntity{},
		CreatedAt:     time.Now(),
	}, nil
}

// Validate validates the timeline event entity
func (te *TimelineEvent) Validate() error {
	if te.Description == "" {
		return ErrTimelineEventDescriptionRequired
	}
	if te.ChapterNumber < 0 {
		return ErrTimelineEventChapterNumberRequired
	}
	return nil
}

// SetTimeline sets the timeline day and confidence
func (te *TimelineEvent) SetTimeline(day int, confidence string) {
	te.TimelineDay = &day
	te.TimelineConfidence = &confidence
}

// AddEntity adds an entity link to this timeline event
func (te *TimelineEvent) AddEntity(role, entityType string, entityID uuid.UUID) {
	te.Entities = append(te.Entities, TimelineEventEntity{
		ID:              uuid.New(),
		TimelineEventID: te.ID,
		Role:            role,
		EntityType:      entityType,
		EntityID:        entityID,
		CreatedAt:       time.Now(),
	})
}

// SetChapterID sets the chapter ID
func (te *TimelineEvent) SetChapterID(chapterID uuid.UUID) {
	te.ChapterID = chapterID
}
