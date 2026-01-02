package artifact

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetArtifactReferencesUseCase handles getting references for an artifact
type GetArtifactReferencesUseCase struct {
	artifactReferenceRepo repositories.ArtifactReferenceRepository
	logger                 logger.Logger
}

// NewGetArtifactReferencesUseCase creates a new GetArtifactReferencesUseCase
func NewGetArtifactReferencesUseCase(
	artifactReferenceRepo repositories.ArtifactReferenceRepository,
	logger logger.Logger,
) *GetArtifactReferencesUseCase {
	return &GetArtifactReferencesUseCase{
		artifactReferenceRepo: artifactReferenceRepo,
		logger:                 logger,
	}
}

// GetArtifactReferencesInput represents the input for getting references
type GetArtifactReferencesInput struct {
	TenantID   uuid.UUID
	ArtifactID uuid.UUID
}

// GetArtifactReferencesOutput represents the output of getting references
type GetArtifactReferencesOutput struct {
	References []*world.ArtifactReference
}

// Execute retrieves all references for an artifact
func (uc *GetArtifactReferencesUseCase) Execute(ctx context.Context, input GetArtifactReferencesInput) (*GetArtifactReferencesOutput, error) {
	references, err := uc.artifactReferenceRepo.ListByArtifact(ctx, input.TenantID, input.ArtifactID)
	if err != nil {
		uc.logger.Error("failed to get artifact references", "error", err, "artifact_id", input.ArtifactID)
		return nil, err
	}

	return &GetArtifactReferencesOutput{
		References: references,
	}, nil
}


