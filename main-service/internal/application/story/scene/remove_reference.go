package scene

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// RemoveSceneReferenceUseCase handles removing a reference from a scene
type RemoveSceneReferenceUseCase struct {
	sceneReferenceRepo repositories.SceneReferenceRepository
	logger             logger.Logger
}

// NewRemoveSceneReferenceUseCase creates a new RemoveSceneReferenceUseCase
func NewRemoveSceneReferenceUseCase(
	sceneReferenceRepo repositories.SceneReferenceRepository,
	logger logger.Logger,
) *RemoveSceneReferenceUseCase {
	return &RemoveSceneReferenceUseCase{
		sceneReferenceRepo: sceneReferenceRepo,
		logger:             logger,
	}
}

// RemoveSceneReferenceInput represents the input for removing a reference
type RemoveSceneReferenceInput struct {
	TenantID   uuid.UUID
	SceneID    uuid.UUID
	EntityType story.SceneReferenceEntityType
	EntityID   uuid.UUID
}

// Execute removes a reference from a scene
func (uc *RemoveSceneReferenceUseCase) Execute(ctx context.Context, input RemoveSceneReferenceInput) error {
	if err := uc.sceneReferenceRepo.DeleteBySceneAndEntity(ctx, input.TenantID, input.SceneID, input.EntityType, input.EntityID); err != nil {
		uc.logger.Error("failed to remove scene reference", "error", err, "scene_id", input.SceneID)
		return err
	}

	uc.logger.Info("scene reference removed", "scene_id", input.SceneID, "entity_type", input.EntityType, "entity_id", input.EntityID)

	return nil
}


