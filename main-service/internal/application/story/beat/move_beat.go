package beat

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// MoveBeatUseCase handles moving a beat to a different scene
type MoveBeatUseCase struct {
	beatRepo  repositories.BeatRepository
	sceneRepo repositories.SceneRepository
	logger    logger.Logger
}

// NewMoveBeatUseCase creates a new MoveBeatUseCase
func NewMoveBeatUseCase(
	beatRepo repositories.BeatRepository,
	sceneRepo repositories.SceneRepository,
	logger logger.Logger,
) *MoveBeatUseCase {
	return &MoveBeatUseCase{
		beatRepo:  beatRepo,
		sceneRepo: sceneRepo,
		logger:    logger,
	}
}

// MoveBeatInput represents the input for moving a beat
type MoveBeatInput struct {
	TenantID  uuid.UUID
	BeatID    uuid.UUID
	NewSceneID uuid.UUID
}

// MoveBeatOutput represents the output of moving a beat
type MoveBeatOutput struct {
	Beat *story.Beat
}

// Execute moves a beat to a different scene
func (uc *MoveBeatUseCase) Execute(ctx context.Context, input MoveBeatInput) (*MoveBeatOutput, error) {
	// Get existing beat
	beat, err := uc.beatRepo.GetByID(ctx, input.TenantID, input.BeatID)
	if err != nil {
		return nil, err
	}

	// Validate new scene exists
	_, err = uc.sceneRepo.GetByID(ctx, input.TenantID, input.NewSceneID)
	if err != nil {
		return nil, err
	}

	// Update scene
	beat.SceneID = input.NewSceneID

	if err := beat.Validate(); err != nil {
		return nil, err
	}

	if err := uc.beatRepo.Update(ctx, beat); err != nil {
		uc.logger.Error("failed to move beat", "error", err, "beat_id", input.BeatID, "tenant_id", input.TenantID)
		return nil, err
	}

	uc.logger.Info("beat moved", "beat_id", input.BeatID, "new_scene_id", input.NewSceneID, "tenant_id", input.TenantID)

	return &MoveBeatOutput{
		Beat: beat,
	}, nil
}

