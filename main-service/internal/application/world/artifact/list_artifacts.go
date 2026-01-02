package artifact

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListArtifactsUseCase handles listing artifacts
type ListArtifactsUseCase struct {
	artifactRepo repositories.ArtifactRepository
	logger       logger.Logger
}

// NewListArtifactsUseCase creates a new ListArtifactsUseCase
func NewListArtifactsUseCase(
	artifactRepo repositories.ArtifactRepository,
	logger logger.Logger,
) *ListArtifactsUseCase {
	return &ListArtifactsUseCase{
		artifactRepo: artifactRepo,
		logger:       logger,
	}
}

// ListArtifactsInput represents the input for listing artifacts
type ListArtifactsInput struct {
	TenantID uuid.UUID
	WorldID  uuid.UUID
	Limit    int
	Offset   int
}

// ListArtifactsOutput represents the output of listing artifacts
type ListArtifactsOutput struct {
	Artifacts []*world.Artifact
	Total     int
}

// Execute lists artifacts for a world
func (uc *ListArtifactsUseCase) Execute(ctx context.Context, input ListArtifactsInput) (*ListArtifactsOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	artifacts, err := uc.artifactRepo.ListByWorld(ctx, input.TenantID, input.WorldID, limit, input.Offset)
	if err != nil {
		uc.logger.Error("failed to list artifacts", "error", err, "world_id", input.WorldID)
		return nil, err
	}

	total, err := uc.artifactRepo.CountByWorld(ctx, input.TenantID, input.WorldID)
	if err != nil {
		uc.logger.Warn("failed to count artifacts", "error", err)
		total = len(artifacts)
	}

	return &ListArtifactsOutput{
		Artifacts: artifacts,
		Total:     total,
	}, nil
}


