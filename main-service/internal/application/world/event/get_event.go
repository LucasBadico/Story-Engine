package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetEventUseCase handles retrieving an event
type GetEventUseCase struct {
	eventRepo repositories.EventRepository
	logger    logger.Logger
}

// NewGetEventUseCase creates a new GetEventUseCase
func NewGetEventUseCase(
	eventRepo repositories.EventRepository,
	logger logger.Logger,
) *GetEventUseCase {
	return &GetEventUseCase{
		eventRepo: eventRepo,
		logger:    logger,
	}
}

// GetEventInput represents the input for getting an event
type GetEventInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetEventOutput represents the output of getting an event
type GetEventOutput struct {
	Event *world.Event
}

// Execute retrieves an event by ID
func (uc *GetEventUseCase) Execute(ctx context.Context, input GetEventInput) (*GetEventOutput, error) {
	event, err := uc.eventRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get event", "error", err, "event_id", input.ID)
		return nil, err
	}

	return &GetEventOutput{
		Event: event,
	}, nil
}


