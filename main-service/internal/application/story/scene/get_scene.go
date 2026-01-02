package scene

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetSceneUseCase handles retrieving a scene
type GetSceneUseCase struct {
	sceneRepo repositories.SceneRepository
	logger    logger.Logger
}

// NewGetSceneUseCase creates a new GetSceneUseCase
func NewGetSceneUseCase(
	sceneRepo repositories.SceneRepository,
	logger logger.Logger,
) *GetSceneUseCase {
	return &GetSceneUseCase{
		sceneRepo: sceneRepo,
		logger:    logger,
	}
}

// GetSceneInput represents the input for getting a scene
type GetSceneInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetSceneOutput represents the output of getting a scene
type GetSceneOutput struct {
	Scene *story.Scene
}

// Execute retrieves a scene by ID
func (uc *GetSceneUseCase) Execute(ctx context.Context, input GetSceneInput) (*GetSceneOutput, error) {
	scene, err := uc.sceneRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get scene", "error", err, "scene_id", input.ID, "tenant_id", input.TenantID)
		return nil, err
	}

	return &GetSceneOutput{
		Scene: scene,
	}, nil
}

