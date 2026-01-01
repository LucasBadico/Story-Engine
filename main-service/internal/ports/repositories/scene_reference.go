package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
)

// SceneReferenceRepository defines the interface for scene reference persistence
type SceneReferenceRepository interface {
	Create(ctx context.Context, ref *story.SceneReference) error
	GetByID(ctx context.Context, id uuid.UUID) (*story.SceneReference, error)
	ListByScene(ctx context.Context, sceneID uuid.UUID) ([]*story.SceneReference, error)
	ListByEntity(ctx context.Context, entityType story.SceneReferenceEntityType, entityID uuid.UUID) ([]*story.SceneReference, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByScene(ctx context.Context, sceneID uuid.UUID) error
	DeleteBySceneAndEntity(ctx context.Context, sceneID uuid.UUID, entityType story.SceneReferenceEntityType, entityID uuid.UUID) error
}

