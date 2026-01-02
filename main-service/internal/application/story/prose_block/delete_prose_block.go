package prose_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteProseBlockUseCase handles prose block deletion
type DeleteProseBlockUseCase struct {
	proseBlockRepo repositories.ProseBlockRepository
	logger         logger.Logger
}

// NewDeleteProseBlockUseCase creates a new DeleteProseBlockUseCase
func NewDeleteProseBlockUseCase(
	proseBlockRepo repositories.ProseBlockRepository,
	logger logger.Logger,
) *DeleteProseBlockUseCase {
	return &DeleteProseBlockUseCase{
		proseBlockRepo: proseBlockRepo,
		logger:         logger,
	}
}

// DeleteProseBlockInput represents the input for deleting a prose block
type DeleteProseBlockInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes a prose block
func (uc *DeleteProseBlockUseCase) Execute(ctx context.Context, input DeleteProseBlockInput) error {
	// Check if prose block exists
	_, err := uc.proseBlockRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	if err := uc.proseBlockRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete prose block", "error", err, "prose_block_id", input.ID, "tenant_id", input.TenantID)
		return err
	}

	uc.logger.Info("prose block deleted", "prose_block_id", input.ID, "tenant_id", input.TenantID)

	return nil
}

