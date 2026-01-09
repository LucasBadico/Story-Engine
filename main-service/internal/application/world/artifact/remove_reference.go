package artifact

import (
	"context"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// RemoveArtifactReferenceUseCase handles removing a reference from an artifact
type RemoveArtifactReferenceUseCase struct {
	listRelationsUseCase  *relationapp.ListRelationsBySourceUseCase
	deleteRelationUseCase *relationapp.DeleteRelationUseCase
	logger                logger.Logger
}

// NewRemoveArtifactReferenceUseCase creates a new RemoveArtifactReferenceUseCase
func NewRemoveArtifactReferenceUseCase(
	listRelationsUseCase *relationapp.ListRelationsBySourceUseCase,
	deleteRelationUseCase *relationapp.DeleteRelationUseCase,
	logger logger.Logger,
) *RemoveArtifactReferenceUseCase {
	return &RemoveArtifactReferenceUseCase{
		listRelationsUseCase:  listRelationsUseCase,
		deleteRelationUseCase: deleteRelationUseCase,
		logger:                logger,
	}
}

// RemoveArtifactReferenceInput represents the input for removing a reference
type RemoveArtifactReferenceInput struct {
	TenantID   uuid.UUID
	ArtifactID uuid.UUID
	EntityType string // "character" or "location"
	EntityID   uuid.UUID
}

// Execute removes a reference from an artifact
func (uc *RemoveArtifactReferenceUseCase) Execute(ctx context.Context, input RemoveArtifactReferenceInput) error {
	// Find the relation by source (artifact) and target (entity)
	output, err := uc.listRelationsUseCase.Execute(ctx, relationapp.ListRelationsBySourceInput{
		TenantID:   input.TenantID,
		SourceType: "artifact",
		SourceID:   input.ArtifactID,
		Options: repositories.ListOptions{
			Limit: 100,
		},
	})
	if err != nil {
		uc.logger.Error("failed to find artifact reference", "error", err, "artifact_id", input.ArtifactID)
		return err
	}

	var relationID *uuid.UUID
	for _, rel := range output.Relations.Items {
		if rel.TargetType == input.EntityType && rel.TargetID == input.EntityID {
			id := rel.ID
			relationID = &id
			break
		}
	}

	if relationID == nil {
		uc.logger.Info("artifact reference not found", "artifact_id", input.ArtifactID, "entity_type", input.EntityType, "entity_id", input.EntityID)
		return nil // Not found, but not an error
	}

	err = uc.deleteRelationUseCase.Execute(ctx, relationapp.DeleteRelationInput{
		TenantID: input.TenantID,
		ID:       *relationID,
	})
	if err != nil {
		uc.logger.Error("failed to remove artifact reference", "error", err, "artifact_id", input.ArtifactID)
		return err
	}

	uc.logger.Info("artifact reference removed", "artifact_id", input.ArtifactID, "entity_type", input.EntityType, "entity_id", input.EntityID)

	return nil
}
