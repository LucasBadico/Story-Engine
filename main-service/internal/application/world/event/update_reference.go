package event

import (
	"context"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	"github.com/story-engine/main-service/internal/platform/logger"
)

// UpdateReferenceUseCase handles updating an event reference
type UpdateReferenceUseCase struct {
	getRelationUseCase    *relationapp.GetRelationUseCase
	updateRelationUseCase *relationapp.UpdateRelationUseCase
	logger                logger.Logger
}

// NewUpdateReferenceUseCase creates a new UpdateReferenceUseCase
func NewUpdateReferenceUseCase(
	getRelationUseCase *relationapp.GetRelationUseCase,
	updateRelationUseCase *relationapp.UpdateRelationUseCase,
	logger logger.Logger,
) *UpdateReferenceUseCase {
	return &UpdateReferenceUseCase{
		getRelationUseCase:    getRelationUseCase,
		updateRelationUseCase: updateRelationUseCase,
		logger:                logger,
	}
}

// UpdateReferenceInput represents the input for updating a reference
type UpdateReferenceInput struct {
	TenantID         uuid.UUID
	ID               uuid.UUID
	RelationshipType *string
	Notes            *string
}

// Execute updates an event reference
func (uc *UpdateReferenceUseCase) Execute(ctx context.Context, input UpdateReferenceInput) error {
	// Get existing relation
	output, err := uc.getRelationUseCase.Execute(ctx, relationapp.GetRelationInput{
		TenantID: input.TenantID,
		ID:       input.ID,
	})
	if err != nil {
		return err
	}

	rel := output.Relation

	// Update fields if provided
	attributes := make(map[string]interface{})
	if rel.Attributes != nil {
		for k, v := range rel.Attributes {
			attributes[k] = v
		}
	}

	if input.Notes != nil {
		attributes["notes"] = *input.Notes
	}

	relationType := rel.RelationType
	if input.RelationshipType != nil {
		relationType = *input.RelationshipType
	}

	_, err = uc.updateRelationUseCase.Execute(ctx, relationapp.UpdateRelationInput{
		TenantID:     input.TenantID,
		ID:           input.ID,
		RelationType: &relationType,
		Attributes:   &attributes,
	})
	if err != nil {
		uc.logger.Error("failed to update event reference", "error", err, "reference_id", input.ID)
		return err
	}

	uc.logger.Info("event reference updated", "reference_id", input.ID)
	return nil
}
