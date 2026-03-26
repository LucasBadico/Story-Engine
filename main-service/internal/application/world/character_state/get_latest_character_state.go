package character_state

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetLatestCharacterStateUseCase handles retrieving the latest character state
type GetLatestCharacterStateUseCase struct {
	characterStateRepo repositories.CharacterStateRepository
	logger             logger.Logger
}

// NewGetLatestCharacterStateUseCase creates a new GetLatestCharacterStateUseCase
func NewGetLatestCharacterStateUseCase(
	characterStateRepo repositories.CharacterStateRepository,
	logger logger.Logger,
) *GetLatestCharacterStateUseCase {
	return &GetLatestCharacterStateUseCase{
		characterStateRepo: characterStateRepo,
		logger:             logger,
	}
}

// GetLatestCharacterStateInput represents the input for getting the latest character state
type GetLatestCharacterStateInput struct {
	TenantID    uuid.UUID
	CharacterID uuid.UUID
	AsOfChapter int
}

// GetLatestCharacterStateOutput represents the output of getting the latest character state
type GetLatestCharacterStateOutput struct {
	CharacterState *world.CharacterState
}

// Execute retrieves the latest character state at or before the given chapter
func (uc *GetLatestCharacterStateUseCase) Execute(ctx context.Context, input GetLatestCharacterStateInput) (*GetLatestCharacterStateOutput, error) {
	cs, err := uc.characterStateRepo.GetLatestByCharacter(ctx, input.TenantID, input.CharacterID, input.AsOfChapter)
	if err != nil {
		uc.logger.Error("failed to get latest character state", "error", err, "character_id", input.CharacterID, "as_of_chapter", input.AsOfChapter)
		return nil, err
	}

	return &GetLatestCharacterStateOutput{
		CharacterState: cs,
	}, nil
}
