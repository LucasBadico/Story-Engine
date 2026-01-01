package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetEventLocationsUseCase handles retrieving locations for an event
type GetEventLocationsUseCase struct {
	eventLocationRepo repositories.EventLocationRepository
	logger            logger.Logger
}

// NewGetEventLocationsUseCase creates a new GetEventLocationsUseCase
func NewGetEventLocationsUseCase(
	eventLocationRepo repositories.EventLocationRepository,
	logger logger.Logger,
) *GetEventLocationsUseCase {
	return &GetEventLocationsUseCase{
		eventLocationRepo: eventLocationRepo,
		logger:           logger,
	}
}

// GetEventLocationsInput represents the input for getting locations
type GetEventLocationsInput struct {
	EventID uuid.UUID
}

// GetEventLocationsOutput represents the output of getting locations
type GetEventLocationsOutput struct {
	Locations []*world.EventLocation
}

// Execute retrieves locations for an event
func (uc *GetEventLocationsUseCase) Execute(ctx context.Context, input GetEventLocationsInput) (*GetEventLocationsOutput, error) {
	locations, err := uc.eventLocationRepo.ListByEvent(ctx, input.EventID)
	if err != nil {
		uc.logger.Error("failed to get event locations", "error", err, "event_id", input.EventID)
		return nil, err
	}

	return &GetEventLocationsOutput{
		Locations: locations,
	}, nil
}

