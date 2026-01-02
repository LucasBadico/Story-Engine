package scene

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateSceneUseCase handles scene updates
type UpdateSceneUseCase struct {
	sceneRepo repositories.SceneRepository
	logger    logger.Logger
}

// NewUpdateSceneUseCase creates a new UpdateSceneUseCase
func NewUpdateSceneUseCase(
	sceneRepo repositories.SceneRepository,
	logger logger.Logger,
) *UpdateSceneUseCase {
	return &UpdateSceneUseCase{
		sceneRepo: sceneRepo,
		logger:    logger,
	}
}

// UpdateSceneInput represents the input for updating a scene
type UpdateSceneInput struct {
	TenantID       uuid.UUID
	ID             uuid.UUID
	OrderNum       *int
	POVCharacterID *uuid.UUID
	TimeRef        *string
	Goal           *string
}

// UpdateSceneOutput represents the output of updating a scene
type UpdateSceneOutput struct {
	Scene *story.Scene
}

// Execute updates a scene
func (uc *UpdateSceneUseCase) Execute(ctx context.Context, input UpdateSceneInput) (*UpdateSceneOutput, error) {
	// Get existing scene
	scene, err := uc.sceneRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.OrderNum != nil {
		if *input.OrderNum < 1 {
			return nil, &platformerrors.ValidationError{
				Field:   "order_num",
				Message: "must be greater than 0",
			}
		}
		scene.OrderNum = *input.OrderNum
	}

	if input.Goal != nil {
		scene.UpdateGoal(*input.Goal)
	}

	if input.TimeRef != nil {
		scene.TimeRef = *input.TimeRef
	}

	if input.POVCharacterID != nil {
		scene.UpdatePOV(input.POVCharacterID)
	}

	if err := scene.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "scene",
			Message: err.Error(),
		}
	}

	if err := uc.sceneRepo.Update(ctx, scene); err != nil {
		uc.logger.Error("failed to update scene", "error", err, "scene_id", input.ID, "tenant_id", input.TenantID)
		return nil, err
	}

	uc.logger.Info("scene updated", "scene_id", input.ID, "tenant_id", input.TenantID)

	return &UpdateSceneOutput{
		Scene: scene,
	}, nil
}

