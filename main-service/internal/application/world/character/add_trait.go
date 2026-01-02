package character

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddTraitToCharacterUseCase handles adding a trait to a character
type AddTraitToCharacterUseCase struct {
	characterRepo      repositories.CharacterRepository
	traitRepo          repositories.TraitRepository
	characterTraitRepo repositories.CharacterTraitRepository
	logger             logger.Logger
}

// NewAddTraitToCharacterUseCase creates a new AddTraitToCharacterUseCase
func NewAddTraitToCharacterUseCase(
	characterRepo repositories.CharacterRepository,
	traitRepo repositories.TraitRepository,
	characterTraitRepo repositories.CharacterTraitRepository,
	logger logger.Logger,
) *AddTraitToCharacterUseCase {
	return &AddTraitToCharacterUseCase{
		characterRepo:      characterRepo,
		traitRepo:          traitRepo,
		characterTraitRepo: characterTraitRepo,
		logger:             logger,
	}
}

// AddTraitToCharacterInput represents the input for adding a trait to a character
type AddTraitToCharacterInput struct {
	CharacterID uuid.UUID
	TraitID     uuid.UUID
	Value       string
	Notes       string
}

// Execute adds a trait to a character (creates a copy/snapshot)
func (uc *AddTraitToCharacterUseCase) Execute(ctx context.Context, input AddTraitToCharacterInput) error {
	// Validate character exists
	_, err := uc.characterRepo.GetByID(ctx, input.CharacterID)
	if err != nil {
		return err
	}

	// Get trait template to copy its data
	trait, err := uc.traitRepo.GetByID(ctx, input.TraitID)
	if err != nil {
		return err
	}

	// Check if trait already exists for this character
	_, err = uc.characterTraitRepo.GetByCharacterAndTrait(ctx, input.CharacterID, input.TraitID)
	if err == nil {
		return &platformerrors.ValidationError{
			Field:   "trait_id",
			Message: "trait already assigned to character",
		}
	}

	// Create character trait with copied trait data
	characterTrait := world.NewCharacterTrait(input.CharacterID, input.TraitID, trait, input.Value)
	if input.Notes != "" {
		characterTrait.UpdateNotes(input.Notes)
	}

	if err := uc.characterTraitRepo.Create(ctx, characterTrait); err != nil {
		uc.logger.Error("failed to add trait to character", "error", err, "character_id", input.CharacterID, "trait_id", input.TraitID)
		return err
	}

	uc.logger.Info("trait added to character", "character_id", input.CharacterID, "trait_id", input.TraitID)

	return nil
}


