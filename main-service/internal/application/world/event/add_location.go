package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddLocationToEventUseCase handles adding a location to an event
type AddLocationToEventUseCase struct {
	eventRepo         repositories.EventRepository
	locationRepo      repositories.LocationRepository
	eventLocationRepo repositories.EventLocationRepository
	logger            logger.Logger
}

// NewAddLocationToEventUseCase creates a new AddLocationToEventUseCase
func NewAddLocationToEventUseCase(
	eventRepo repositories.EventRepository,
	locationRepo repositories.LocationRepository,
	eventLocationRepo repositories.EventLocationRepository,
	logger logger.Logger,
) *AddLocationToEventUseCase {
	return &AddLocationToEventUseCase{
		eventRepo:         eventRepo,
		locationRepo:      locationRepo,
		eventLocationRepo: eventLocationRepo,
		logger:            logger,
	}
}

// AddLocationToEventInput represents the input for adding a location to an event
type AddLocationToEventInput struct {
	EventID      uuid.UUID
	LocationID   uuid.UUID
	Significance *string
}

// Execute adds a location to an event
func (uc *AddLocationToEventUseCase) Execute(ctx context.Context, input AddLocationToEventInput) error {
	// Validate event exists
	event, err := uc.eventRepo.GetByID(ctx, input.EventID)
	if err != nil {
		return err
	}

	// Validate location exists and belongs to same world
	location, err := uc.locationRepo.GetByID(ctx, input.LocationID)
	if err != nil {
		return err
	}
	if location.WorldID != event.WorldID {
		return &platformerrors.ValidationError{
			Field:   "location_id",
			Message: "location must belong to the same world as the event",
		}
	}

	// Create relationship
	el := world.NewEventLocation(input.EventID, input.LocationID, input.Significance)
	if err := uc.eventLocationRepo.Create(ctx, el); err != nil {
		uc.logger.Error("failed to add location to event", "error", err, "event_id", input.EventID, "location_id", input.LocationID)
		return err
	}

	uc.logger.Info("location added to event", "event_id", input.EventID, "location_id", input.LocationID)
	return nil
}


