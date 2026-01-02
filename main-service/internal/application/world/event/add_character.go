package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddCharacterToEventUseCase handles adding a character to an event
type AddCharacterToEventUseCase struct {
	eventRepo          repositories.EventRepository
	characterRepo      repositories.CharacterRepository
	eventCharacterRepo repositories.EventCharacterRepository
	logger             logger.Logger
}

// NewAddCharacterToEventUseCase creates a new AddCharacterToEventUseCase
func NewAddCharacterToEventUseCase(
	eventRepo repositories.EventRepository,
	characterRepo repositories.CharacterRepository,
	eventCharacterRepo repositories.EventCharacterRepository,
	logger logger.Logger,
) *AddCharacterToEventUseCase {
	return &AddCharacterToEventUseCase{
		eventRepo:          eventRepo,
		characterRepo:      characterRepo,
		eventCharacterRepo: eventCharacterRepo,
		logger:             logger,
	}
}

// AddCharacterToEventInput represents the input for adding a character to an event
type AddCharacterToEventInput struct {
	TenantID    uuid.UUID
	EventID     uuid.UUID
	CharacterID uuid.UUID
	Role        *string
}

// Execute adds a character to an event
func (uc *AddCharacterToEventUseCase) Execute(ctx context.Context, input AddCharacterToEventInput) error {
	// Validate event exists
	event, err := uc.eventRepo.GetByID(ctx, input.TenantID, input.EventID)
	if err != nil {
		return err
	}

	// Validate character exists and belongs to same world
	character, err := uc.characterRepo.GetByID(ctx, input.TenantID, input.CharacterID)
	if err != nil {
		return err
	}
	if character.WorldID != event.WorldID {
		return &platformerrors.ValidationError{
			Field:   "character_id",
			Message: "character must belong to the same world as the event",
		}
	}

	// Create relationship
	ec := world.NewEventCharacter(input.EventID, input.CharacterID, input.Role)
	if err := uc.eventCharacterRepo.Create(ctx, ec); err != nil {
		uc.logger.Error("failed to add character to event", "error", err, "event_id", input.EventID, "character_id", input.CharacterID)
		return err
	}

	uc.logger.Info("character added to event", "event_id", input.EventID, "character_id", input.CharacterID)
	return nil
}


