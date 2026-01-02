package character_stats

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetActiveCharacterStatsUseCase handles retrieving active character stats
type GetActiveCharacterStatsUseCase struct {
	characterStatsRepo repositories.CharacterRPGStatsRepository
	logger             logger.Logger
}

// NewGetActiveCharacterStatsUseCase creates a new GetActiveCharacterStatsUseCase
func NewGetActiveCharacterStatsUseCase(
	characterStatsRepo repositories.CharacterRPGStatsRepository,
	logger logger.Logger,
) *GetActiveCharacterStatsUseCase {
	return &GetActiveCharacterStatsUseCase{
		characterStatsRepo: characterStatsRepo,
		logger:             logger,
	}
}

// GetActiveCharacterStatsInput represents the input for getting active stats
type GetActiveCharacterStatsInput struct {
	TenantID    uuid.UUID
	CharacterID uuid.UUID
}

// GetActiveCharacterStatsOutput represents the output of getting active stats
type GetActiveCharacterStatsOutput struct {
	Stats *rpg.CharacterRPGStats
}

// Execute retrieves active character RPG stats
func (uc *GetActiveCharacterStatsUseCase) Execute(ctx context.Context, input GetActiveCharacterStatsInput) (*GetActiveCharacterStatsOutput, error) {
	stats, err := uc.characterStatsRepo.GetActiveByCharacter(ctx, input.TenantID, input.CharacterID)
	if err != nil {
		uc.logger.Error("failed to get active character RPG stats", "error", err, "character_id", input.CharacterID)
		return nil, err
	}

	return &GetActiveCharacterStatsOutput{
		Stats: stats,
	}, nil
}


