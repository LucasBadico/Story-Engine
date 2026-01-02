package artifact_stats

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetActiveArtifactStatsUseCase handles retrieving active artifact stats
type GetActiveArtifactStatsUseCase struct {
	artifactStatsRepo repositories.ArtifactRPGStatsRepository
	logger            logger.Logger
}

// NewGetActiveArtifactStatsUseCase creates a new GetActiveArtifactStatsUseCase
func NewGetActiveArtifactStatsUseCase(
	artifactStatsRepo repositories.ArtifactRPGStatsRepository,
	logger logger.Logger,
) *GetActiveArtifactStatsUseCase {
	return &GetActiveArtifactStatsUseCase{
		artifactStatsRepo: artifactStatsRepo,
		logger:            logger,
	}
}

// GetActiveArtifactStatsInput represents the input for getting active stats
type GetActiveArtifactStatsInput struct {
	ArtifactID uuid.UUID
}

// GetActiveArtifactStatsOutput represents the output of getting active stats
type GetActiveArtifactStatsOutput struct {
	Stats *rpg.ArtifactRPGStats
}

// Execute retrieves active artifact RPG stats
func (uc *GetActiveArtifactStatsUseCase) Execute(ctx context.Context, input GetActiveArtifactStatsInput) (*GetActiveArtifactStatsOutput, error) {
	stats, err := uc.artifactStatsRepo.GetActiveByArtifact(ctx, input.ArtifactID)
	if err != nil {
		uc.logger.Error("failed to get active artifact RPG stats", "error", err, "artifact_id", input.ArtifactID)
		return nil, err
	}

	return &GetActiveArtifactStatsOutput{
		Stats: stats,
	}, nil
}


