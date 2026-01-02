package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
)

// ChapterRepository defines the interface for chapter persistence
type ChapterRepository interface {
	Create(ctx context.Context, c *story.Chapter) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.Chapter, error)
	ListByStory(ctx context.Context, tenantID, storyID uuid.UUID) ([]*story.Chapter, error)
	ListByStoryOrdered(ctx context.Context, tenantID, storyID uuid.UUID) ([]*story.Chapter, error)
	Update(ctx context.Context, c *story.Chapter) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	DeleteByStory(ctx context.Context, tenantID, storyID uuid.UUID) error
}

