package content_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteContentBlockUseCase handles content block deletion
type DeleteContentBlockUseCase struct {
	contentBlockRepo repositories.ContentBlockRepository
	logger           logger.Logger
}

// NewDeleteContentBlockUseCase creates a new DeleteContentBlockUseCase
func NewDeleteContentBlockUseCase(
	contentBlockRepo repositories.ContentBlockRepository,
	logger logger.Logger,
) *DeleteContentBlockUseCase {
	return &DeleteContentBlockUseCase{
		contentBlockRepo: contentBlockRepo,
		logger:           logger,
	}
}

// DeleteContentBlockInput represents the input for deleting a content block
type DeleteContentBlockInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes a content block
func (uc *DeleteContentBlockUseCase) Execute(ctx context.Context, input DeleteContentBlockInput) error {
	// Check if content block exists
	_, err := uc.contentBlockRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	if err := uc.contentBlockRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete content block", "error", err, "content_block_id", input.ID, "tenant_id", input.TenantID)
		return err
	}

	uc.logger.Info("content block deleted", "content_block_id", input.ID, "tenant_id", input.TenantID)

	return nil
}

