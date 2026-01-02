package artifact_stats

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListArtifactStatsHistoryUseCase handles listing artifact stats history
type ListArtifactStatsHistoryUseCase struct {
	artifactStatsRepo repositories.ArtifactRPGStatsRepository
	logger            logger.Logger
}

// NewListArtifactStatsHistoryUseCase creates a new ListArtifactStatsHistoryUseCase
func NewListArtifactStatsHistoryUseCase(
	artifactStatsRepo repositories.ArtifactRPGStatsRepository,
	logger logger.Logger,
) *ListArtifactStatsHistoryUseCase {
	return &ListArtifactStatsHistoryUseCase{
		artifactStatsRepo: artifactStatsRepo,
		logger:            logger,
	}
}

// ListArtifactStatsHistoryInput represents the input for listing stats history
type ListArtifactStatsHistoryInput struct {
	TenantID   uuid.UUID
	ArtifactID uuid.UUID
}

// ListArtifactStatsHistoryOutput represents the output of listing stats history
type ListArtifactStatsHistoryOutput struct {
	Stats []*rpg.ArtifactRPGStats
}

// Execute lists all stats versions for an artifact
func (uc *ListArtifactStatsHistoryUseCase) Execute(ctx context.Context, input ListArtifactStatsHistoryInput) (*ListArtifactStatsHistoryOutput, error) {
	statsList, err := uc.artifactStatsRepo.ListByArtifact(ctx, input.TenantID, input.ArtifactID)
	if err != nil {
		uc.logger.Error("failed to list artifact RPG stats history", "error", err, "artifact_id", input.ArtifactID)
		return nil, err
	}

	return &ListArtifactStatsHistoryOutput{
		Stats: statsList,
	}, nil
}


