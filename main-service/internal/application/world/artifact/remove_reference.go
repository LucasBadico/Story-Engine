package artifact

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// RemoveArtifactReferenceUseCase handles removing a reference from an artifact
type RemoveArtifactReferenceUseCase struct {
	artifactReferenceRepo repositories.ArtifactReferenceRepository
	logger                 logger.Logger
}

// NewRemoveArtifactReferenceUseCase creates a new RemoveArtifactReferenceUseCase
func NewRemoveArtifactReferenceUseCase(
	artifactReferenceRepo repositories.ArtifactReferenceRepository,
	logger logger.Logger,
) *RemoveArtifactReferenceUseCase {
	return &RemoveArtifactReferenceUseCase{
		artifactReferenceRepo: artifactReferenceRepo,
		logger:                 logger,
	}
}

// RemoveArtifactReferenceInput represents the input for removing a reference
type RemoveArtifactReferenceInput struct {
	TenantID   uuid.UUID
	ArtifactID uuid.UUID
	EntityType world.ArtifactReferenceEntityType
	EntityID   uuid.UUID
}

// Execute removes a reference from an artifact
func (uc *RemoveArtifactReferenceUseCase) Execute(ctx context.Context, input RemoveArtifactReferenceInput) error {
	if err := uc.artifactReferenceRepo.DeleteByArtifactAndEntity(ctx, input.TenantID, input.ArtifactID, input.EntityType, input.EntityID); err != nil {
		uc.logger.Error("failed to remove artifact reference", "error", err, "artifact_id", input.ArtifactID)
		return err
	}

	uc.logger.Info("artifact reference removed", "artifact_id", input.ArtifactID, "entity_type", input.EntityType, "entity_id", input.EntityID)

	return nil
}

