package faction

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateReferenceUseCase handles updating a faction reference
type UpdateReferenceUseCase struct {
	factionReferenceRepo repositories.FactionReferenceRepository
	logger               logger.Logger
}

// NewUpdateReferenceUseCase creates a new UpdateReferenceUseCase
func NewUpdateReferenceUseCase(
	factionReferenceRepo repositories.FactionReferenceRepository,
	logger logger.Logger,
) *UpdateReferenceUseCase {
	return &UpdateReferenceUseCase{
		factionReferenceRepo: factionReferenceRepo,
		logger:               logger,
	}
}

// UpdateReferenceInput represents the input for updating a reference
type UpdateReferenceInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
	Role     *string
	Notes    *string
}

// Execute updates a faction reference
func (uc *UpdateReferenceUseCase) Execute(ctx context.Context, input UpdateReferenceInput) error {
	// Get existing reference
	ref, err := uc.factionReferenceRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	// Update fields if provided
	if input.Role != nil {
		ref.UpdateRole(input.Role)
	}
	if input.Notes != nil {
		ref.UpdateNotes(*input.Notes)
	}

	if err := uc.factionReferenceRepo.Update(ctx, ref); err != nil {
		uc.logger.Error("failed to update faction reference", "error", err, "reference_id", input.ID)
		return err
	}

	uc.logger.Info("faction reference updated", "reference_id", input.ID)
	return nil
}
