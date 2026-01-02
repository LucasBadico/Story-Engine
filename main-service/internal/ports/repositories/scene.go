package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
)

// SceneRepository defines the interface for scene persistence
type SceneRepository interface {
	Create(ctx context.Context, s *story.Scene) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.Scene, error)
	ListByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) ([]*story.Scene, error)
	ListByChapterOrdered(ctx context.Context, tenantID, chapterID uuid.UUID) ([]*story.Scene, error)
	ListByStory(ctx context.Context, tenantID, storyID uuid.UUID) ([]*story.Scene, error)
	Update(ctx context.Context, s *story.Scene) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	DeleteByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) error
	DeleteByStory(ctx context.Context, tenantID, storyID uuid.UUID) error
}

