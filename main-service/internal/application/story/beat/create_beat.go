package beat

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateBeatUseCase handles beat creation
type CreateBeatUseCase struct {
	beatRepo  repositories.BeatRepository
	sceneRepo repositories.SceneRepository
	logger    logger.Logger
}

// NewCreateBeatUseCase creates a new CreateBeatUseCase
func NewCreateBeatUseCase(
	beatRepo repositories.BeatRepository,
	sceneRepo repositories.SceneRepository,
	logger logger.Logger,
) *CreateBeatUseCase {
	return &CreateBeatUseCase{
		beatRepo:  beatRepo,
		sceneRepo: sceneRepo,
		logger:    logger,
	}
}

// CreateBeatInput represents the input for creating a beat
type CreateBeatInput struct {
	TenantID uuid.UUID
	SceneID  uuid.UUID
	OrderNum int
	Type     story.BeatType
	Intent   string
	Outcome  string
}

// CreateBeatOutput represents the output of creating a beat
type CreateBeatOutput struct {
	Beat *story.Beat
}

// Execute creates a new beat
func (uc *CreateBeatUseCase) Execute(ctx context.Context, input CreateBeatInput) (*CreateBeatOutput, error) {
	// Validate scene exists
	_, err := uc.sceneRepo.GetByID(ctx, input.TenantID, input.SceneID)
	if err != nil {
		return nil, err
	}

	if input.OrderNum < 1 {
		return nil, &platformerrors.ValidationError{
			Field:   "order_num",
			Message: "must be greater than 0",
		}
	}

	if input.Type == "" {
		return nil, &platformerrors.ValidationError{
			Field:   "type",
			Message: "type is required",
		}
	}

	beat, err := story.NewBeat(input.TenantID, input.SceneID, input.OrderNum, input.Type)
	if err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "beat",
			Message: err.Error(),
		}
	}

	if input.Intent != "" {
		beat.UpdateIntent(input.Intent)
	}
	if input.Outcome != "" {
		beat.UpdateOutcome(input.Outcome)
	}

	if err := beat.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "beat",
			Message: err.Error(),
		}
	}

	if err := uc.beatRepo.Create(ctx, beat); err != nil {
		uc.logger.Error("failed to create beat", "error", err, "scene_id", input.SceneID, "tenant_id", input.TenantID)
		return nil, err
	}

	uc.logger.Info("beat created", "beat_id", beat.ID, "scene_id", input.SceneID, "tenant_id", input.TenantID)

	return &CreateBeatOutput{
		Beat: beat,
	}, nil
}

