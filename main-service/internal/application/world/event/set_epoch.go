package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// SetEpochUseCase handles setting an event as epoch
type SetEpochUseCase struct {
	eventRepo repositories.EventRepository
	logger    logger.Logger
}

// NewSetEpochUseCase creates a new SetEpochUseCase
func NewSetEpochUseCase(
	eventRepo repositories.EventRepository,
	logger logger.Logger,
) *SetEpochUseCase {
	return &SetEpochUseCase{
		eventRepo: eventRepo,
		logger:    logger,
	}
}

// SetEpochInput represents the input for setting epoch
type SetEpochInput struct {
	TenantID uuid.UUID
	EventID  uuid.UUID
}

// Execute sets an event as the epoch (time zero) of the world
func (uc *SetEpochUseCase) Execute(ctx context.Context, input SetEpochInput) error {
	// Get event
	event, err := uc.eventRepo.GetByID(ctx, input.TenantID, input.EventID)
	if err != nil {
		return err
	}

	// Check if there's already an epoch for this world
	existingEpoch, err := uc.eventRepo.GetEpoch(ctx, input.TenantID, event.WorldID)
	if err == nil && existingEpoch.ID != event.ID {
		// Remove epoch flag from existing epoch
		existingEpoch.SetAsEpoch(false)
		if err := uc.eventRepo.Update(ctx, existingEpoch); err != nil {
			uc.logger.Error("failed to remove epoch from existing event", "error", err, "event_id", existingEpoch.ID)
			return err
		}
	}

	// Set this event as epoch
	event.SetAsEpoch(true)
	if err := uc.eventRepo.Update(ctx, event); err != nil {
		uc.logger.Error("failed to set event as epoch", "error", err, "event_id", input.EventID)
		return err
	}

	uc.logger.Info("event set as epoch", "event_id", input.EventID, "world_id", event.WorldID)
	return nil
}

