package prose_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteProseBlockReferenceUseCase handles prose block reference deletion
type DeleteProseBlockReferenceUseCase struct {
	refRepo repositories.ProseBlockReferenceRepository
	logger  logger.Logger
}

// NewDeleteProseBlockReferenceUseCase creates a new DeleteProseBlockReferenceUseCase
func NewDeleteProseBlockReferenceUseCase(
	refRepo repositories.ProseBlockReferenceRepository,
	logger logger.Logger,
) *DeleteProseBlockReferenceUseCase {
	return &DeleteProseBlockReferenceUseCase{
		refRepo: refRepo,
		logger:  logger,
	}
}

// DeleteProseBlockReferenceInput represents the input for deleting a reference
type DeleteProseBlockReferenceInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes a prose block reference
func (uc *DeleteProseBlockReferenceUseCase) Execute(ctx context.Context, input DeleteProseBlockReferenceInput) error {
	// Check if reference exists
	_, err := uc.refRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	if err := uc.refRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete prose block reference", "error", err, "reference_id", input.ID, "tenant_id", input.TenantID)
		return err
	}

	uc.logger.Info("prose block reference deleted", "reference_id", input.ID, "tenant_id", input.TenantID)

	return nil
}

