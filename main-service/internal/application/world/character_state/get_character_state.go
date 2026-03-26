package character_state

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetCharacterStateUseCase handles character state retrieval
type GetCharacterStateUseCase struct {
	characterStateRepo repositories.CharacterStateRepository
	logger             logger.Logger
}

// NewGetCharacterStateUseCase creates a new GetCharacterStateUseCase
func NewGetCharacterStateUseCase(
	characterStateRepo repositories.CharacterStateRepository,
	logger logger.Logger,
) *GetCharacterStateUseCase {
	return &GetCharacterStateUseCase{
		characterStateRepo: characterStateRepo,
		logger:             logger,
	}
}

// GetCharacterStateInput represents the input for getting a character state
type GetCharacterStateInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetCharacterStateOutput represents the output of getting a character state
type GetCharacterStateOutput struct {
	CharacterState *world.CharacterState
}

// Execute retrieves a character state by ID
func (uc *GetCharacterStateUseCase) Execute(ctx context.Context, input GetCharacterStateInput) (*GetCharacterStateOutput, error) {
	cs, err := uc.characterStateRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get character state", "error", err, "character_state_id", input.ID)
		return nil, err
	}

	return &GetCharacterStateOutput{
		CharacterState: cs,
	}, nil
}
