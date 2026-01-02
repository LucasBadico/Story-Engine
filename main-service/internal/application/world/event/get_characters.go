package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetEventCharactersUseCase handles retrieving characters for an event
type GetEventCharactersUseCase struct {
	eventCharacterRepo repositories.EventCharacterRepository
	logger             logger.Logger
}

// NewGetEventCharactersUseCase creates a new GetEventCharactersUseCase
func NewGetEventCharactersUseCase(
	eventCharacterRepo repositories.EventCharacterRepository,
	logger logger.Logger,
) *GetEventCharactersUseCase {
	return &GetEventCharactersUseCase{
		eventCharacterRepo: eventCharacterRepo,
		logger:             logger,
	}
}

// GetEventCharactersInput represents the input for getting characters
type GetEventCharactersInput struct {
	EventID uuid.UUID
}

// GetEventCharactersOutput represents the output of getting characters
type GetEventCharactersOutput struct {
	Characters []*world.EventCharacter
}

// Execute retrieves characters for an event
func (uc *GetEventCharactersUseCase) Execute(ctx context.Context, input GetEventCharactersInput) (*GetEventCharactersOutput, error) {
	characters, err := uc.eventCharacterRepo.ListByEvent(ctx, input.EventID)
	if err != nil {
		uc.logger.Error("failed to get event characters", "error", err, "event_id", input.EventID)
		return nil, err
	}

	return &GetEventCharactersOutput{
		Characters: characters,
	}, nil
}


