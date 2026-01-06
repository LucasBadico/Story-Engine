package character

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetCharacterEventsUseCase handles getting events for a character
type GetCharacterEventsUseCase struct {
	eventReferenceRepo repositories.EventReferenceRepository
	logger             logger.Logger
}

// NewGetCharacterEventsUseCase creates a new GetCharacterEventsUseCase
func NewGetCharacterEventsUseCase(
	eventReferenceRepo repositories.EventReferenceRepository,
	logger logger.Logger,
) *GetCharacterEventsUseCase {
	return &GetCharacterEventsUseCase{
		eventReferenceRepo: eventReferenceRepo,
		logger:             logger,
	}
}

// GetCharacterEventsInput represents the input for getting character events
type GetCharacterEventsInput struct {
	TenantID    uuid.UUID
	CharacterID uuid.UUID
}

// GetCharacterEventsOutput represents the output of getting character events
type GetCharacterEventsOutput struct {
	EventReferences []*world.EventReference
}

// Execute retrieves all event references for a character
func (uc *GetCharacterEventsUseCase) Execute(ctx context.Context, input GetCharacterEventsInput) (*GetCharacterEventsOutput, error) {
	references, err := uc.eventReferenceRepo.ListByEntity(ctx, input.TenantID, "character", input.CharacterID)
	if err != nil {
		uc.logger.Error("failed to get character events", "error", err, "character_id", input.CharacterID)
		return nil, err
	}

	return &GetCharacterEventsOutput{
		EventReferences: references,
	}, nil
}

