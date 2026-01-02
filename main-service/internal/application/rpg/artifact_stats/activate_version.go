package artifact_stats

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ActivateArtifactStatsVersionUseCase handles activating a specific stats version (rollback)
type ActivateArtifactStatsVersionUseCase struct {
	artifactStatsRepo repositories.ArtifactRPGStatsRepository
	logger            logger.Logger
}

// NewActivateArtifactStatsVersionUseCase creates a new ActivateArtifactStatsVersionUseCase
func NewActivateArtifactStatsVersionUseCase(
	artifactStatsRepo repositories.ArtifactRPGStatsRepository,
	logger logger.Logger,
) *ActivateArtifactStatsVersionUseCase {
	return &ActivateArtifactStatsVersionUseCase{
		artifactStatsRepo: artifactStatsRepo,
		logger:            logger,
	}
}

// ActivateArtifactStatsVersionInput represents the input for activating a version
type ActivateArtifactStatsVersionInput struct {
	StatsID uuid.UUID
}

// ActivateArtifactStatsVersionOutput represents the output of activating a version
type ActivateArtifactStatsVersionOutput struct {
	Stats *rpg.ArtifactRPGStats
}

// Execute activates a specific stats version (rollback)
func (uc *ActivateArtifactStatsVersionUseCase) Execute(ctx context.Context, input ActivateArtifactStatsVersionInput) (*ActivateArtifactStatsVersionOutput, error) {
	// Get the stats version
	stats, err := uc.artifactStatsRepo.GetByID(ctx, input.StatsID)
	if err != nil {
		return nil, err
	}

	// Deactivate all versions for this artifact
	if err := uc.artifactStatsRepo.DeactivateAllByArtifact(ctx, stats.ArtifactID); err != nil {
		return nil, err
	}

	// Activate this version
	stats.SetActive(true)
	if err := uc.artifactStatsRepo.Update(ctx, stats); err != nil {
		uc.logger.Error("failed to activate artifact RPG stats version", "error", err, "stats_id", input.StatsID)
		return nil, err
	}

	uc.logger.Info("artifact RPG stats version activated", "stats_id", input.StatsID, "artifact_id", stats.ArtifactID, "version", stats.Version)

	return &ActivateArtifactStatsVersionOutput{
		Stats: stats,
	}, nil
}


