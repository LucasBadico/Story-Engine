package character

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// RemoveTraitFromCharacterUseCase handles removing a trait from a character
type RemoveTraitFromCharacterUseCase struct {
	characterTraitRepo repositories.CharacterTraitRepository
	logger             logger.Logger
}

// NewRemoveTraitFromCharacterUseCase creates a new RemoveTraitFromCharacterUseCase
func NewRemoveTraitFromCharacterUseCase(
	characterTraitRepo repositories.CharacterTraitRepository,
	logger logger.Logger,
) *RemoveTraitFromCharacterUseCase {
	return &RemoveTraitFromCharacterUseCase{
		characterTraitRepo: characterTraitRepo,
		logger:             logger,
	}
}

// RemoveTraitFromCharacterInput represents the input for removing a trait from a character
type RemoveTraitFromCharacterInput struct {
	TenantID    uuid.UUID
	CharacterID uuid.UUID
	TraitID     uuid.UUID
}

// Execute removes a trait from a character
func (uc *RemoveTraitFromCharacterUseCase) Execute(ctx context.Context, input RemoveTraitFromCharacterInput) error {
	ct, err := uc.characterTraitRepo.GetByCharacterAndTrait(ctx, input.TenantID, input.CharacterID, input.TraitID)
	if err != nil {
		return err
	}

	if err := uc.characterTraitRepo.Delete(ctx, input.TenantID, ct.ID); err != nil {
		uc.logger.Error("failed to remove trait from character", "error", err, "character_id", input.CharacterID, "trait_id", input.TraitID)
		return err
	}

	uc.logger.Info("trait removed from character", "character_id", input.CharacterID, "trait_id", input.TraitID)

	return nil
}

