package beat

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateBeatUseCase handles beat updates
type UpdateBeatUseCase struct {
	beatRepo repositories.BeatRepository
	logger   logger.Logger
}

// NewUpdateBeatUseCase creates a new UpdateBeatUseCase
func NewUpdateBeatUseCase(
	beatRepo repositories.BeatRepository,
	logger logger.Logger,
) *UpdateBeatUseCase {
	return &UpdateBeatUseCase{
		beatRepo: beatRepo,
		logger:   logger,
	}
}

// UpdateBeatInput represents the input for updating a beat
type UpdateBeatInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
	OrderNum *int
	Type     *story.BeatType
	Intent   *string
	Outcome  *string
}

// UpdateBeatOutput represents the output of updating a beat
type UpdateBeatOutput struct {
	Beat *story.Beat
}

// Execute updates a beat
func (uc *UpdateBeatUseCase) Execute(ctx context.Context, input UpdateBeatInput) (*UpdateBeatOutput, error) {
	// Get existing beat
	beat, err := uc.beatRepo.GetByID(ctx, input.TenantID, input.ID)
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
		beat.OrderNum = *input.OrderNum
	}

	if input.Type != nil {
		beat.Type = *input.Type
	}

	if input.Intent != nil {
		beat.UpdateIntent(*input.Intent)
	}

	if input.Outcome != nil {
		beat.UpdateOutcome(*input.Outcome)
	}

	if err := beat.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "beat",
			Message: err.Error(),
		}
	}

	if err := uc.beatRepo.Update(ctx, beat); err != nil {
		uc.logger.Error("failed to update beat", "error", err, "beat_id", input.ID, "tenant_id", input.TenantID)
		return nil, err
	}

	uc.logger.Info("beat updated", "beat_id", input.ID, "tenant_id", input.TenantID)

	return &UpdateBeatOutput{
		Beat: beat,
	}, nil
}

