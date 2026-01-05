package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetTimelineUseCase handles retrieving events ordered by timeline position
type GetTimelineUseCase struct {
	eventRepo repositories.EventRepository
	logger    logger.Logger
}

// NewGetTimelineUseCase creates a new GetTimelineUseCase
func NewGetTimelineUseCase(
	eventRepo repositories.EventRepository,
	logger logger.Logger,
) *GetTimelineUseCase {
	return &GetTimelineUseCase{
		eventRepo: eventRepo,
		logger:    logger,
	}
}

// GetTimelineInput represents the input for getting timeline
type GetTimelineInput struct {
	TenantID uuid.UUID
	WorldID  uuid.UUID
	FromPos  *float64
	ToPos    *float64
}

// GetTimelineOutput represents the output of getting timeline
type GetTimelineOutput struct {
	Events []*world.Event
}

// Execute retrieves events ordered by timeline position
func (uc *GetTimelineUseCase) Execute(ctx context.Context, input GetTimelineInput) (*GetTimelineOutput, error) {
	events, err := uc.eventRepo.ListByTimeline(ctx, input.TenantID, input.WorldID, input.FromPos, input.ToPos)
	if err != nil {
		uc.logger.Error("failed to get timeline", "error", err, "world_id", input.WorldID)
		return nil, err
	}

	return &GetTimelineOutput{
		Events: events,
	}, nil
}

