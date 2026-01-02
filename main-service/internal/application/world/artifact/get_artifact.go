package artifact

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetArtifactUseCase handles artifact retrieval
type GetArtifactUseCase struct {
	artifactRepo repositories.ArtifactRepository
	logger       logger.Logger
}

// NewGetArtifactUseCase creates a new GetArtifactUseCase
func NewGetArtifactUseCase(
	artifactRepo repositories.ArtifactRepository,
	logger logger.Logger,
) *GetArtifactUseCase {
	return &GetArtifactUseCase{
		artifactRepo: artifactRepo,
		logger:       logger,
	}
}

// GetArtifactInput represents the input for getting an artifact
type GetArtifactInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetArtifactOutput represents the output of getting an artifact
type GetArtifactOutput struct {
	Artifact *world.Artifact
}

// Execute retrieves an artifact by ID
func (uc *GetArtifactUseCase) Execute(ctx context.Context, input GetArtifactInput) (*GetArtifactOutput, error) {
	a, err := uc.artifactRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get artifact", "error", err, "artifact_id", input.ID)
		return nil, err
	}

	return &GetArtifactOutput{
		Artifact: a,
	}, nil
}


