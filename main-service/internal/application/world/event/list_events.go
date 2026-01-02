package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListEventsUseCase handles listing events for a world
type ListEventsUseCase struct {
	eventRepo repositories.EventRepository
	logger    logger.Logger
}

// NewListEventsUseCase creates a new ListEventsUseCase
func NewListEventsUseCase(
	eventRepo repositories.EventRepository,
	logger logger.Logger,
) *ListEventsUseCase {
	return &ListEventsUseCase{
		eventRepo: eventRepo,
		logger:    logger,
	}
}

// ListEventsInput represents the input for listing events
type ListEventsInput struct {
	WorldID uuid.UUID
}

// ListEventsOutput represents the output of listing events
type ListEventsOutput struct {
	Events []*world.Event
}

// Execute lists events for a world
func (uc *ListEventsUseCase) Execute(ctx context.Context, input ListEventsInput) (*ListEventsOutput, error) {
	events, err := uc.eventRepo.ListByWorld(ctx, input.WorldID)
	if err != nil {
		uc.logger.Error("failed to list events", "error", err, "world_id", input.WorldID)
		return nil, err
	}

	return &ListEventsOutput{
		Events: events,
	}, nil
}


