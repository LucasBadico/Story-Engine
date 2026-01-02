package beat

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetBeatUseCase handles retrieving a beat
type GetBeatUseCase struct {
	beatRepo repositories.BeatRepository
	logger   logger.Logger
}

// NewGetBeatUseCase creates a new GetBeatUseCase
func NewGetBeatUseCase(
	beatRepo repositories.BeatRepository,
	logger logger.Logger,
) *GetBeatUseCase {
	return &GetBeatUseCase{
		beatRepo: beatRepo,
		logger:   logger,
	}
}

// GetBeatInput represents the input for getting a beat
type GetBeatInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetBeatOutput represents the output of getting a beat
type GetBeatOutput struct {
	Beat *story.Beat
}

// Execute retrieves a beat by ID
func (uc *GetBeatUseCase) Execute(ctx context.Context, input GetBeatInput) (*GetBeatOutput, error) {
	beat, err := uc.beatRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get beat", "error", err, "beat_id", input.ID, "tenant_id", input.TenantID)
		return nil, err
	}

	return &GetBeatOutput{
		Beat: beat,
	}, nil
}

