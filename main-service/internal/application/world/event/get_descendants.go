package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetDescendantsUseCase handles retrieving descendants of an event
type GetDescendantsUseCase struct {
	eventRepo repositories.EventRepository
	logger    logger.Logger
}

// NewGetDescendantsUseCase creates a new GetDescendantsUseCase
func NewGetDescendantsUseCase(
	eventRepo repositories.EventRepository,
	logger logger.Logger,
) *GetDescendantsUseCase {
	return &GetDescendantsUseCase{
		eventRepo: eventRepo,
		logger:    logger,
	}
}

// GetDescendantsInput represents the input for getting descendants
type GetDescendantsInput struct {
	TenantID uuid.UUID
	EventID  uuid.UUID
}

// GetDescendantsOutput represents the output of getting descendants
type GetDescendantsOutput struct {
	Events []*world.Event
}

// Execute retrieves all descendants (complete subtree)
func (uc *GetDescendantsUseCase) Execute(ctx context.Context, input GetDescendantsInput) (*GetDescendantsOutput, error) {
	events, err := uc.eventRepo.GetDescendants(ctx, input.TenantID, input.EventID)
	if err != nil {
		uc.logger.Error("failed to get event descendants", "error", err, "event_id", input.EventID)
		return nil, err
	}

	return &GetDescendantsOutput{
		Events: events,
	}, nil
}

