package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetAncestorsUseCase handles retrieving ancestors of an event
type GetAncestorsUseCase struct {
	eventRepo repositories.EventRepository
	logger    logger.Logger
}

// NewGetAncestorsUseCase creates a new GetAncestorsUseCase
func NewGetAncestorsUseCase(
	eventRepo repositories.EventRepository,
	logger logger.Logger,
) *GetAncestorsUseCase {
	return &GetAncestorsUseCase{
		eventRepo: eventRepo,
		logger:    logger,
	}
}

// GetAncestorsInput represents the input for getting ancestors
type GetAncestorsInput struct {
	TenantID uuid.UUID
	EventID  uuid.UUID
}

// GetAncestorsOutput represents the output of getting ancestors
type GetAncestorsOutput struct {
	Events []*world.Event
}

// Execute retrieves the chain of ancestors up to the root
func (uc *GetAncestorsUseCase) Execute(ctx context.Context, input GetAncestorsInput) (*GetAncestorsOutput, error) {
	events, err := uc.eventRepo.GetAncestors(ctx, input.TenantID, input.EventID)
	if err != nil {
		uc.logger.Error("failed to get event ancestors", "error", err, "event_id", input.EventID)
		return nil, err
	}

	return &GetAncestorsOutput{
		Events: events,
	}, nil
}

