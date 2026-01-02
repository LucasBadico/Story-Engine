package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
)

// SceneReferenceRepository defines the interface for scene reference persistence
type SceneReferenceRepository interface {
	Create(ctx context.Context, ref *story.SceneReference) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.SceneReference, error)
	ListByScene(ctx context.Context, tenantID, sceneID uuid.UUID) ([]*story.SceneReference, error)
	ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType story.SceneReferenceEntityType, entityID uuid.UUID) ([]*story.SceneReference, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	DeleteByScene(ctx context.Context, tenantID, sceneID uuid.UUID) error
	DeleteBySceneAndEntity(ctx context.Context, tenantID, sceneID uuid.UUID, entityType story.SceneReferenceEntityType, entityID uuid.UUID) error
}


