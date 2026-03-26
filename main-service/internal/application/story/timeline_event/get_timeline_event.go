package timeline_event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetTimelineEventUseCase handles timeline event retrieval
type GetTimelineEventUseCase struct {
	timelineEventRepo repositories.TimelineEventRepository
	logger            logger.Logger
}

// NewGetTimelineEventUseCase creates a new GetTimelineEventUseCase
func NewGetTimelineEventUseCase(
	timelineEventRepo repositories.TimelineEventRepository,
	logger logger.Logger,
) *GetTimelineEventUseCase {
	return &GetTimelineEventUseCase{
		timelineEventRepo: timelineEventRepo,
		logger:            logger,
	}
}

// GetTimelineEventInput represents the input for getting a timeline event
type GetTimelineEventInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetTimelineEventOutput represents the output of getting a timeline event
type GetTimelineEventOutput struct {
	TimelineEvent *story.TimelineEvent
}

// Execute retrieves a timeline event by ID
func (uc *GetTimelineEventUseCase) Execute(ctx context.Context, input GetTimelineEventInput) (*GetTimelineEventOutput, error) {
	te, err := uc.timelineEventRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get timeline event", "error", err, "timeline_event_id", input.ID)
		return nil, err
	}

	return &GetTimelineEventOutput{
		TimelineEvent: te,
	}, nil
}
