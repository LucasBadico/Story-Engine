package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetEpochUseCase handles retrieving the epoch event for a world
type GetEpochUseCase struct {
	eventRepo repositories.EventRepository
	logger    logger.Logger
}

// NewGetEpochUseCase creates a new GetEpochUseCase
func NewGetEpochUseCase(
	eventRepo repositories.EventRepository,
	logger logger.Logger,
) *GetEpochUseCase {
	return &GetEpochUseCase{
		eventRepo: eventRepo,
		logger:    logger,
	}
}

// GetEpochInput represents the input for getting epoch
type GetEpochInput struct {
	TenantID uuid.UUID
	WorldID  uuid.UUID
}

// GetEpochOutput represents the output of getting epoch
type GetEpochOutput struct {
	Event *world.Event
}

// Execute retrieves the epoch event (time zero) for a world
func (uc *GetEpochUseCase) Execute(ctx context.Context, input GetEpochInput) (*GetEpochOutput, error) {
	event, err := uc.eventRepo.GetEpoch(ctx, input.TenantID, input.WorldID)
	if err != nil {
		uc.logger.Error("failed to get epoch event", "error", err, "world_id", input.WorldID)
		return nil, err
	}

	return &GetEpochOutput{
		Event: event,
	}, nil
}

