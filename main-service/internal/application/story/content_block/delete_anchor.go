package content_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteContentAnchorUseCase handles content anchor deletion
type DeleteContentAnchorUseCase struct {
	anchorRepo repositories.ContentAnchorRepository
	logger     logger.Logger
}

// NewDeleteContentAnchorUseCase creates a new DeleteContentAnchorUseCase
func NewDeleteContentAnchorUseCase(
	anchorRepo repositories.ContentAnchorRepository,
	logger logger.Logger,
) *DeleteContentAnchorUseCase {
	return &DeleteContentAnchorUseCase{
		anchorRepo: anchorRepo,
		logger:     logger,
	}
}

// DeleteContentAnchorInput represents the input for deleting an anchor
type DeleteContentAnchorInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes a content anchor
func (uc *DeleteContentAnchorUseCase) Execute(ctx context.Context, input DeleteContentAnchorInput) error {
	_, err := uc.anchorRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	if err := uc.anchorRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete content anchor", "error", err, "anchor_id", input.ID, "tenant_id", input.TenantID)
		return err
	}

	uc.logger.Info("content anchor deleted", "anchor_id", input.ID, "tenant_id", input.TenantID)

	return nil
}


