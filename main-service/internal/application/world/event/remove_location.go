package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// RemoveLocationFromEventUseCase handles removing a location from an event
type RemoveLocationFromEventUseCase struct {
	eventLocationRepo repositories.EventLocationRepository
	logger            logger.Logger
}

// NewRemoveLocationFromEventUseCase creates a new RemoveLocationFromEventUseCase
func NewRemoveLocationFromEventUseCase(
	eventLocationRepo repositories.EventLocationRepository,
	logger logger.Logger,
) *RemoveLocationFromEventUseCase {
	return &RemoveLocationFromEventUseCase{
		eventLocationRepo: eventLocationRepo,
		logger:           logger,
	}
}

// RemoveLocationFromEventInput represents the input for removing a location from an event
type RemoveLocationFromEventInput struct {
	EventID    uuid.UUID
	LocationID uuid.UUID
}

// Execute removes a location from an event
func (uc *RemoveLocationFromEventUseCase) Execute(ctx context.Context, input RemoveLocationFromEventInput) error {
	if err := uc.eventLocationRepo.DeleteByEventAndLocation(ctx, input.EventID, input.LocationID); err != nil {
		uc.logger.Error("failed to remove location from event", "error", err, "event_id", input.EventID, "location_id", input.LocationID)
		return err
	}

	uc.logger.Info("location removed from event", "event_id", input.EventID, "location_id", input.LocationID)
	return nil
}


