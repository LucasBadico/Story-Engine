package scene

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// RemoveSceneReferenceUseCase handles removing a reference from a scene
type RemoveSceneReferenceUseCase struct {
	listRelationsUseCase  *relationapp.ListRelationsBySourceUseCase
	deleteRelationUseCase *relationapp.DeleteRelationUseCase
	logger                logger.Logger
}

// NewRemoveSceneReferenceUseCase creates a new RemoveSceneReferenceUseCase
func NewRemoveSceneReferenceUseCase(
	listRelationsUseCase *relationapp.ListRelationsBySourceUseCase,
	deleteRelationUseCase *relationapp.DeleteRelationUseCase,
	logger logger.Logger,
) *RemoveSceneReferenceUseCase {
	return &RemoveSceneReferenceUseCase{
		listRelationsUseCase:  listRelationsUseCase,
		deleteRelationUseCase: deleteRelationUseCase,
		logger:                logger,
	}
}

// RemoveSceneReferenceInput represents the input for removing a reference
type RemoveSceneReferenceInput struct {
	TenantID   uuid.UUID
	SceneID    uuid.UUID
	EntityType string // "character", "location", or "artifact"
	EntityID   uuid.UUID
}

// Execute removes a reference from a scene
func (uc *RemoveSceneReferenceUseCase) Execute(ctx context.Context, input RemoveSceneReferenceInput) error {
	// Find the relation by source (scene) and target (entity)
	output, err := uc.listRelationsUseCase.Execute(ctx, relationapp.ListRelationsBySourceInput{
		TenantID:   input.TenantID,
		SourceType: "scene",
		SourceID:   input.SceneID,
		Options: repositories.ListOptions{
			Limit: 100,
		},
	})
	if err != nil {
		uc.logger.Error("failed to find scene reference", "error", err, "scene_id", input.SceneID)
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
		uc.logger.Info("scene reference not found", "scene_id", input.SceneID, "entity_type", input.EntityType, "entity_id", input.EntityID)
		return &platformerrors.NotFoundError{
			Resource: "scene_reference",
			ID:       fmt.Sprintf("scene:%s-%s:%s", input.SceneID, input.EntityType, input.EntityID),
		}
	}

	err = uc.deleteRelationUseCase.Execute(ctx, relationapp.DeleteRelationInput{
		TenantID: input.TenantID,
		ID:       *relationID,
	})
	if err != nil {
		uc.logger.Error("failed to remove scene reference", "error", err, "scene_id", input.SceneID)
		return err
	}

	uc.logger.Info("scene reference removed", "scene_id", input.SceneID, "entity_type", input.EntityType, "entity_id", input.EntityID)

	return nil
}
