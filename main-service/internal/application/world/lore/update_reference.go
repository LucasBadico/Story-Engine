package lore

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateReferenceUseCase handles updating a lore reference
type UpdateReferenceUseCase struct {
	loreReferenceRepo repositories.LoreReferenceRepository
	logger            logger.Logger
}

// NewUpdateReferenceUseCase creates a new UpdateReferenceUseCase
func NewUpdateReferenceUseCase(
	loreReferenceRepo repositories.LoreReferenceRepository,
	logger logger.Logger,
) *UpdateReferenceUseCase {
	return &UpdateReferenceUseCase{
		loreReferenceRepo: loreReferenceRepo,
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

// Execute updates a lore reference
func (uc *UpdateReferenceUseCase) Execute(ctx context.Context, input UpdateReferenceInput) error {
	// Get existing reference
	ref, err := uc.loreReferenceRepo.GetByID(ctx, input.TenantID, input.ID)
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

	if err := uc.loreReferenceRepo.Update(ctx, ref); err != nil {
		uc.logger.Error("failed to update lore reference", "error", err, "reference_id", input.ID)
		return err
	}

	uc.logger.Info("lore reference updated", "reference_id", input.ID)
	return nil
}
