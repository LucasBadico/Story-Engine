package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// RemoveCharacterFromEventUseCase handles removing a character from an event
type RemoveCharacterFromEventUseCase struct {
	eventCharacterRepo repositories.EventCharacterRepository
	logger             logger.Logger
}

// NewRemoveCharacterFromEventUseCase creates a new RemoveCharacterFromEventUseCase
func NewRemoveCharacterFromEventUseCase(
	eventCharacterRepo repositories.EventCharacterRepository,
	logger logger.Logger,
) *RemoveCharacterFromEventUseCase {
	return &RemoveCharacterFromEventUseCase{
		eventCharacterRepo: eventCharacterRepo,
		logger:             logger,
	}
}

// RemoveCharacterFromEventInput represents the input for removing a character from an event
type RemoveCharacterFromEventInput struct {
	EventID     uuid.UUID
	CharacterID uuid.UUID
}

// Execute removes a character from an event
func (uc *RemoveCharacterFromEventUseCase) Execute(ctx context.Context, input RemoveCharacterFromEventInput) error {
	if err := uc.eventCharacterRepo.DeleteByEventAndCharacter(ctx, input.EventID, input.CharacterID); err != nil {
		uc.logger.Error("failed to remove character from event", "error", err, "event_id", input.EventID, "character_id", input.CharacterID)
		return err
	}

	uc.logger.Info("character removed from event", "event_id", input.EventID, "character_id", input.CharacterID)
	return nil
}


