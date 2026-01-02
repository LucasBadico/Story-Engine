package character_skill

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteCharacterSkillUseCase handles character skill deletion (unlearn)
type DeleteCharacterSkillUseCase struct {
	characterSkillRepo repositories.CharacterSkillRepository
	logger             logger.Logger
}

// NewDeleteCharacterSkillUseCase creates a new DeleteCharacterSkillUseCase
func NewDeleteCharacterSkillUseCase(
	characterSkillRepo repositories.CharacterSkillRepository,
	logger logger.Logger,
) *DeleteCharacterSkillUseCase {
	return &DeleteCharacterSkillUseCase{
		characterSkillRepo: characterSkillRepo,
		logger:             logger,
	}
}

// DeleteCharacterSkillInput represents the input for deleting a character skill
type DeleteCharacterSkillInput struct {
	ID uuid.UUID
}

// Execute deletes a character skill (unlearn)
func (uc *DeleteCharacterSkillUseCase) Execute(ctx context.Context, input DeleteCharacterSkillInput) error {
	if err := uc.characterSkillRepo.Delete(ctx, input.ID); err != nil {
		uc.logger.Error("failed to delete character skill", "error", err, "id", input.ID)
		return err
	}

	uc.logger.Info("character skill deleted", "id", input.ID)

	return nil
}


