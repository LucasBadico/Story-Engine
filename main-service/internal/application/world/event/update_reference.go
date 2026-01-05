package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateReferenceUseCase handles updating an event reference
type UpdateReferenceUseCase struct {
	eventReferenceRepo repositories.EventReferenceRepository
	logger              logger.Logger
}

// NewUpdateReferenceUseCase creates a new UpdateReferenceUseCase
func NewUpdateReferenceUseCase(
	eventReferenceRepo repositories.EventReferenceRepository,
	logger logger.Logger,
) *UpdateReferenceUseCase {
	return &UpdateReferenceUseCase{
		eventReferenceRepo: eventReferenceRepo,
		logger:            logger,
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
	// Get existing reference
	ref, err := uc.eventReferenceRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	// Update fields if provided
	if input.RelationshipType != nil {
		ref.UpdateRelationshipType(input.RelationshipType)
	}
	if input.Notes != nil {
		ref.UpdateNotes(*input.Notes)
	}

	if err := uc.eventReferenceRepo.Update(ctx, ref); err != nil {
		uc.logger.Error("failed to update event reference", "error", err, "reference_id", input.ID)
		return err
	}

	uc.logger.Info("event reference updated", "reference_id", input.ID)
	return nil
}

