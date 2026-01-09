package event

import (
	"context"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// RemoveReferenceUseCase handles removing a reference from an event
type RemoveReferenceUseCase struct {
	listRelationsUseCase  *relationapp.ListRelationsBySourceUseCase
	deleteRelationUseCase *relationapp.DeleteRelationUseCase
	logger                logger.Logger
}

// NewRemoveReferenceUseCase creates a new RemoveReferenceUseCase
func NewRemoveReferenceUseCase(
	listRelationsUseCase *relationapp.ListRelationsBySourceUseCase,
	deleteRelationUseCase *relationapp.DeleteRelationUseCase,
	logger logger.Logger,
) *RemoveReferenceUseCase {
	return &RemoveReferenceUseCase{
		listRelationsUseCase:  listRelationsUseCase,
		deleteRelationUseCase: deleteRelationUseCase,
		logger:                logger,
	}
}

// RemoveReferenceInput represents the input for removing a reference
type RemoveReferenceInput struct {
	TenantID   uuid.UUID
	EventID    uuid.UUID
	EntityType string
	EntityID   uuid.UUID
}

// Execute removes a reference from an event
func (uc *RemoveReferenceUseCase) Execute(ctx context.Context, input RemoveReferenceInput) error {
	// Find the relation by source (event) and target (entity)
	output, err := uc.listRelationsUseCase.Execute(ctx, relationapp.ListRelationsBySourceInput{
		TenantID:   input.TenantID,
		SourceType: "event",
		SourceID:   input.EventID,
		Options: repositories.ListOptions{
			Limit: 100,
		},
	})
	if err != nil {
		uc.logger.Error("failed to find event reference", "error", err, "event_id", input.EventID)
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
		uc.logger.Info("event reference not found", "event_id", input.EventID, "entity_type", input.EntityType, "entity_id", input.EntityID)
		return nil // Not found, but not an error
	}

	err = uc.deleteRelationUseCase.Execute(ctx, relationapp.DeleteRelationInput{
		TenantID: input.TenantID,
		ID:       *relationID,
	})
	if err != nil {
		uc.logger.Error("failed to remove event reference", "error", err, "event_id", input.EventID)
		return err
	}

	uc.logger.Info("event reference removed", "event_id", input.EventID, "entity_type", input.EntityType, "entity_id", input.EntityID)
	return nil
}
