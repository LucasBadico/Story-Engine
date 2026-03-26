package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
)

// TimelineEventRepository defines the interface for timeline event persistence
type TimelineEventRepository interface {
	Create(ctx context.Context, te *story.TimelineEvent) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.TimelineEvent, error)
	ListByStory(ctx context.Context, tenantID, storyID uuid.UUID, asOfChapter int, limit int) ([]*story.TimelineEvent, error)
	ListByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) ([]*story.TimelineEvent, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	// Entity links
	AddEntity(ctx context.Context, entity *story.TimelineEventEntity) error
	ListEntities(ctx context.Context, timelineEventID uuid.UUID) ([]*story.TimelineEventEntity, error)
}
