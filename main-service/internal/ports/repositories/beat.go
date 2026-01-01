package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
)

// BeatRepository defines the interface for beat persistence
type BeatRepository interface {
	Create(ctx context.Context, b *story.Beat) error
	GetByID(ctx context.Context, id uuid.UUID) (*story.Beat, error)
	ListByScene(ctx context.Context, sceneID uuid.UUID) ([]*story.Beat, error)
	ListBySceneOrdered(ctx context.Context, sceneID uuid.UUID) ([]*story.Beat, error)
	ListByStory(ctx context.Context, storyID uuid.UUID) ([]*story.Beat, error)
	Update(ctx context.Context, b *story.Beat) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByScene(ctx context.Context, sceneID uuid.UUID) error
}

