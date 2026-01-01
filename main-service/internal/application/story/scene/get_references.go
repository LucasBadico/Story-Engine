package scene

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetSceneReferencesUseCase handles getting references for a scene
type GetSceneReferencesUseCase struct {
	sceneReferenceRepo repositories.SceneReferenceRepository
	logger             logger.Logger
}

// NewGetSceneReferencesUseCase creates a new GetSceneReferencesUseCase
func NewGetSceneReferencesUseCase(
	sceneReferenceRepo repositories.SceneReferenceRepository,
	logger logger.Logger,
) *GetSceneReferencesUseCase {
	return &GetSceneReferencesUseCase{
		sceneReferenceRepo: sceneReferenceRepo,
		logger:             logger,
	}
}

// GetSceneReferencesInput represents the input for getting references
type GetSceneReferencesInput struct {
	SceneID uuid.UUID
}

// GetSceneReferencesOutput represents the output of getting references
type GetSceneReferencesOutput struct {
	References []*story.SceneReference
}

// Execute retrieves all references for a scene
func (uc *GetSceneReferencesUseCase) Execute(ctx context.Context, input GetSceneReferencesInput) (*GetSceneReferencesOutput, error) {
	references, err := uc.sceneReferenceRepo.ListByScene(ctx, input.SceneID)
	if err != nil {
		uc.logger.Error("failed to get scene references", "error", err, "scene_id", input.SceneID)
		return nil, err
	}

	return &GetSceneReferencesOutput{
		References: references,
	}, nil
}

