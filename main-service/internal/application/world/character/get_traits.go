package character

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetCharacterTraitsUseCase handles getting traits for a character
type GetCharacterTraitsUseCase struct {
	characterTraitRepo repositories.CharacterTraitRepository
	logger             logger.Logger
}

// NewGetCharacterTraitsUseCase creates a new GetCharacterTraitsUseCase
func NewGetCharacterTraitsUseCase(
	characterTraitRepo repositories.CharacterTraitRepository,
	logger logger.Logger,
) *GetCharacterTraitsUseCase {
	return &GetCharacterTraitsUseCase{
		characterTraitRepo: characterTraitRepo,
		logger:             logger,
	}
}

// GetCharacterTraitsInput represents the input for getting character traits
type GetCharacterTraitsInput struct {
	TenantID    uuid.UUID
	CharacterID uuid.UUID
}

// GetCharacterTraitsOutput represents the output of getting character traits
type GetCharacterTraitsOutput struct {
	Traits []*world.CharacterTrait
}

// Execute retrieves all traits for a character
func (uc *GetCharacterTraitsUseCase) Execute(ctx context.Context, input GetCharacterTraitsInput) (*GetCharacterTraitsOutput, error) {
	traits, err := uc.characterTraitRepo.GetByCharacter(ctx, input.TenantID, input.CharacterID)
	if err != nil {
		uc.logger.Error("failed to get character traits", "error", err, "character_id", input.CharacterID)
		return nil, err
	}

	return &GetCharacterTraitsOutput{
		Traits: traits,
	}, nil
}


