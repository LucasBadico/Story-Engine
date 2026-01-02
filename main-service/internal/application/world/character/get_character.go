package character

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetCharacterUseCase handles character retrieval
type GetCharacterUseCase struct {
	characterRepo repositories.CharacterRepository
	logger        logger.Logger
}

// NewGetCharacterUseCase creates a new GetCharacterUseCase
func NewGetCharacterUseCase(
	characterRepo repositories.CharacterRepository,
	logger logger.Logger,
) *GetCharacterUseCase {
	return &GetCharacterUseCase{
		characterRepo: characterRepo,
		logger:        logger,
	}
}

// GetCharacterInput represents the input for getting a character
type GetCharacterInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetCharacterOutput represents the output of getting a character
type GetCharacterOutput struct {
	Character *world.Character
}

// Execute retrieves a character by ID
func (uc *GetCharacterUseCase) Execute(ctx context.Context, input GetCharacterInput) (*GetCharacterOutput, error) {
	c, err := uc.characterRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get character", "error", err, "character_id", input.ID)
		return nil, err
	}

	return &GetCharacterOutput{
		Character: c,
	}, nil
}


