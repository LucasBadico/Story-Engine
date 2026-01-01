package character_stats

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteAllCharacterStatsUseCase handles deleting all stats for a character
type DeleteAllCharacterStatsUseCase struct {
	characterStatsRepo repositories.CharacterRPGStatsRepository
	logger             logger.Logger
}

// NewDeleteAllCharacterStatsUseCase creates a new DeleteAllCharacterStatsUseCase
func NewDeleteAllCharacterStatsUseCase(
	characterStatsRepo repositories.CharacterRPGStatsRepository,
	logger logger.Logger,
) *DeleteAllCharacterStatsUseCase {
	return &DeleteAllCharacterStatsUseCase{
		characterStatsRepo: characterStatsRepo,
		logger:             logger,
	}
}

// DeleteAllCharacterStatsInput represents the input for deleting all stats
type DeleteAllCharacterStatsInput struct {
	CharacterID uuid.UUID
}

// Execute deletes all stats for a character
func (uc *DeleteAllCharacterStatsUseCase) Execute(ctx context.Context, input DeleteAllCharacterStatsInput) error {
	if err := uc.characterStatsRepo.DeleteByCharacter(ctx, input.CharacterID); err != nil {
		uc.logger.Error("failed to delete all character RPG stats", "error", err, "character_id", input.CharacterID)
		return err
	}

	uc.logger.Info("all character RPG stats deleted", "character_id", input.CharacterID)

	return nil
}

