package character_stats

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ActivateCharacterStatsVersionUseCase handles activating a specific stats version (rollback)
type ActivateCharacterStatsVersionUseCase struct {
	characterStatsRepo repositories.CharacterRPGStatsRepository
	logger             logger.Logger
}

// NewActivateCharacterStatsVersionUseCase creates a new ActivateCharacterStatsVersionUseCase
func NewActivateCharacterStatsVersionUseCase(
	characterStatsRepo repositories.CharacterRPGStatsRepository,
	logger logger.Logger,
) *ActivateCharacterStatsVersionUseCase {
	return &ActivateCharacterStatsVersionUseCase{
		characterStatsRepo: characterStatsRepo,
		logger:             logger,
	}
}

// ActivateCharacterStatsVersionInput represents the input for activating a version
type ActivateCharacterStatsVersionInput struct {
	StatsID uuid.UUID
}

// ActivateCharacterStatsVersionOutput represents the output of activating a version
type ActivateCharacterStatsVersionOutput struct {
	Stats *rpg.CharacterRPGStats
}

// Execute activates a specific stats version (rollback)
func (uc *ActivateCharacterStatsVersionUseCase) Execute(ctx context.Context, input ActivateCharacterStatsVersionInput) (*ActivateCharacterStatsVersionOutput, error) {
	// Get the stats version
	stats, err := uc.characterStatsRepo.GetByID(ctx, input.StatsID)
	if err != nil {
		return nil, err
	}

	// Deactivate all versions for this character
	if err := uc.characterStatsRepo.DeactivateAllByCharacter(ctx, stats.CharacterID); err != nil {
		return nil, err
	}

	// Activate this version
	stats.SetActive(true)
	if err := uc.characterStatsRepo.Update(ctx, stats); err != nil {
		uc.logger.Error("failed to activate character RPG stats version", "error", err, "stats_id", input.StatsID)
		return nil, err
	}

	uc.logger.Info("character RPG stats version activated", "stats_id", input.StatsID, "character_id", stats.CharacterID, "version", stats.Version)

	return &ActivateCharacterStatsVersionOutput{
		Stats: stats,
	}, nil
}


