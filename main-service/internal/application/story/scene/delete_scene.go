package scene

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteSceneUseCase handles scene deletion
type DeleteSceneUseCase struct {
	sceneRepo    repositories.SceneRepository
	relationRepo repositories.EntityRelationRepository
	logger       logger.Logger
}

// NewDeleteSceneUseCase creates a new DeleteSceneUseCase
func NewDeleteSceneUseCase(
	sceneRepo repositories.SceneRepository,
	relationRepo repositories.EntityRelationRepository,
	logger logger.Logger,
) *DeleteSceneUseCase {
	return &DeleteSceneUseCase{
		sceneRepo:    sceneRepo,
		relationRepo: relationRepo,
		logger:       logger,
	}
}

// DeleteSceneInput represents the input for deleting a scene
type DeleteSceneInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes a scene
func (uc *DeleteSceneUseCase) Execute(ctx context.Context, input DeleteSceneInput) error {
	// Check if scene exists
	_, err := uc.sceneRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	// Delete relations where scene is source or target
	if err := uc.relationRepo.DeleteByEntity(ctx, input.TenantID, "scene", input.ID); err != nil {
		uc.logger.Error("failed to delete scene relations", "error", err)
		// Continue anyway
	}

	if err := uc.sceneRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete scene", "error", err, "scene_id", input.ID, "tenant_id", input.TenantID)
		return err
	}

	uc.logger.Info("scene deleted", "scene_id", input.ID, "tenant_id", input.TenantID)

	return nil
}
