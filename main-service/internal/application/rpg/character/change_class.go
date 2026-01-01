package character

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ChangeCharacterClassUseCase handles changing a character's class
type ChangeCharacterClassUseCase struct {
	characterRepo repositories.CharacterRepository
	classRepo     repositories.RPGClassRepository
	logger        logger.Logger
}

// NewChangeCharacterClassUseCase creates a new ChangeCharacterClassUseCase
func NewChangeCharacterClassUseCase(
	characterRepo repositories.CharacterRepository,
	classRepo repositories.RPGClassRepository,
	logger logger.Logger,
) *ChangeCharacterClassUseCase {
	return &ChangeCharacterClassUseCase{
		characterRepo: characterRepo,
		classRepo:     classRepo,
		logger:        logger,
	}
}

// ChangeCharacterClassInput represents the input for changing a character's class
type ChangeCharacterClassInput struct {
	CharacterID uuid.UUID
	ClassID     *uuid.UUID // nil = remove class
	ClassLevel  *int       // optional: set level (defaults to 1)
}

// ChangeCharacterClassOutput represents the output of changing a character's class
type ChangeCharacterClassOutput struct {
	Character *world.Character
}

// Execute changes a character's class
func (uc *ChangeCharacterClassUseCase) Execute(ctx context.Context, input ChangeCharacterClassInput) (*ChangeCharacterClassOutput, error) {
	// Get character
	character, err := uc.characterRepo.GetByID(ctx, input.CharacterID)
	if err != nil {
		return nil, err
	}

	// Validate class exists if provided
	if input.ClassID != nil {
		_, err := uc.classRepo.GetByID(ctx, *input.ClassID)
		if err != nil {
			return nil, err
		}
	}

	// Set class
	character.SetClass(input.ClassID)

	// Set level
	if input.ClassLevel != nil {
		if err := character.SetClassLevel(*input.ClassLevel); err != nil {
			return nil, err
		}
	} else if input.ClassID != nil && character.CurrentClassID == nil {
		// New class, set level to 1
		character.SetClassLevel(1)
	}

	if err := character.Validate(); err != nil {
		return nil, err
	}

	if err := uc.characterRepo.Update(ctx, character); err != nil {
		uc.logger.Error("failed to change character class", "error", err, "character_id", input.CharacterID)
		return nil, err
	}

	uc.logger.Info("character class changed", "character_id", input.CharacterID, "class_id", input.ClassID)

	return &ChangeCharacterClassOutput{
		Character: character,
	}, nil
}

