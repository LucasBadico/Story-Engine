package content_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteContentBlockReferenceUseCase handles content block reference deletion
type DeleteContentBlockReferenceUseCase struct {
	refRepo repositories.ContentBlockReferenceRepository
	logger  logger.Logger
}

// NewDeleteContentBlockReferenceUseCase creates a new DeleteContentBlockReferenceUseCase
func NewDeleteContentBlockReferenceUseCase(
	refRepo repositories.ContentBlockReferenceRepository,
	logger logger.Logger,
) *DeleteContentBlockReferenceUseCase {
	return &DeleteContentBlockReferenceUseCase{
		refRepo: refRepo,
		logger:  logger,
	}
}

// DeleteContentBlockReferenceInput represents the input for deleting a reference
type DeleteContentBlockReferenceInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes a content block reference
func (uc *DeleteContentBlockReferenceUseCase) Execute(ctx context.Context, input DeleteContentBlockReferenceInput) error {
	// Check if reference exists
	_, err := uc.refRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	if err := uc.refRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete content block reference", "error", err, "reference_id", input.ID, "tenant_id", input.TenantID)
		return err
	}

	uc.logger.Info("content block reference deleted", "reference_id", input.ID, "tenant_id", input.TenantID)

	return nil
}

