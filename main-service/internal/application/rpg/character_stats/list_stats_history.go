package character_stats

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListCharacterStatsHistoryUseCase handles listing character stats history
type ListCharacterStatsHistoryUseCase struct {
	characterStatsRepo repositories.CharacterRPGStatsRepository
	logger             logger.Logger
}

// NewListCharacterStatsHistoryUseCase creates a new ListCharacterStatsHistoryUseCase
func NewListCharacterStatsHistoryUseCase(
	characterStatsRepo repositories.CharacterRPGStatsRepository,
	logger logger.Logger,
) *ListCharacterStatsHistoryUseCase {
	return &ListCharacterStatsHistoryUseCase{
		characterStatsRepo: characterStatsRepo,
		logger:             logger,
	}
}

// ListCharacterStatsHistoryInput represents the input for listing stats history
type ListCharacterStatsHistoryInput struct {
	CharacterID uuid.UUID
}

// ListCharacterStatsHistoryOutput represents the output of listing stats history
type ListCharacterStatsHistoryOutput struct {
	Stats []*rpg.CharacterRPGStats
}

// Execute lists all stats versions for a character
func (uc *ListCharacterStatsHistoryUseCase) Execute(ctx context.Context, input ListCharacterStatsHistoryInput) (*ListCharacterStatsHistoryOutput, error) {
	statsList, err := uc.characterStatsRepo.ListByCharacter(ctx, input.CharacterID)
	if err != nil {
		uc.logger.Error("failed to list character RPG stats history", "error", err, "character_id", input.CharacterID)
		return nil, err
	}

	return &ListCharacterStatsHistoryOutput{
		Stats: statsList,
	}, nil
}


