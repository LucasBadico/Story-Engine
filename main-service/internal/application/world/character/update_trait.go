package character

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateCharacterTraitUseCase handles updating a character trait
type UpdateCharacterTraitUseCase struct {
	characterTraitRepo repositories.CharacterTraitRepository
	traitRepo          repositories.TraitRepository
	logger             logger.Logger
}

// NewUpdateCharacterTraitUseCase creates a new UpdateCharacterTraitUseCase
func NewUpdateCharacterTraitUseCase(
	characterTraitRepo repositories.CharacterTraitRepository,
	traitRepo repositories.TraitRepository,
	logger logger.Logger,
) *UpdateCharacterTraitUseCase {
	return &UpdateCharacterTraitUseCase{
		characterTraitRepo: characterTraitRepo,
		traitRepo:          traitRepo,
		logger:             logger,
	}
}

// UpdateCharacterTraitInput represents the input for updating a character trait
type UpdateCharacterTraitInput struct {
	TenantID    uuid.UUID
	CharacterID uuid.UUID
	TraitID     uuid.UUID
	Value       *string
	Notes       *string
}

// UpdateCharacterTraitOutput represents the output of updating a character trait
type UpdateCharacterTraitOutput struct {
	CharacterTrait *world.CharacterTrait
}

// Execute updates a character trait
func (uc *UpdateCharacterTraitUseCase) Execute(ctx context.Context, input UpdateCharacterTraitInput) (*UpdateCharacterTraitOutput, error) {
	ct, err := uc.characterTraitRepo.GetByCharacterAndTrait(ctx, input.TenantID, input.CharacterID, input.TraitID)
	if err != nil {
		return nil, err
	}

	if input.Value != nil {
		ct.UpdateValue(*input.Value)
	}
	if input.Notes != nil {
		ct.UpdateNotes(*input.Notes)
	}

	if err := uc.characterTraitRepo.Update(ctx, ct); err != nil {
		uc.logger.Error("failed to update character trait", "error", err, "character_id", input.CharacterID, "trait_id", input.TraitID)
		return nil, err
	}

	uc.logger.Info("character trait updated", "character_id", input.CharacterID, "trait_id", input.TraitID)

	return &UpdateCharacterTraitOutput{
		CharacterTrait: ct,
	}, nil
}

