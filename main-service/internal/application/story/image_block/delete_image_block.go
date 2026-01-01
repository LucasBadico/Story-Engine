package image_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteImageBlockUseCase handles image block deletion
type DeleteImageBlockUseCase struct {
	imageBlockRepo         repositories.ImageBlockRepository
	imageBlockReferenceRepo repositories.ImageBlockReferenceRepository
	logger                 logger.Logger
}

// NewDeleteImageBlockUseCase creates a new DeleteImageBlockUseCase
func NewDeleteImageBlockUseCase(
	imageBlockRepo repositories.ImageBlockRepository,
	imageBlockReferenceRepo repositories.ImageBlockReferenceRepository,
	logger logger.Logger,
) *DeleteImageBlockUseCase {
	return &DeleteImageBlockUseCase{
		imageBlockRepo:         imageBlockRepo,
		imageBlockReferenceRepo: imageBlockReferenceRepo,
		logger:                 logger,
	}
}

// DeleteImageBlockInput represents the input for deleting an image block
type DeleteImageBlockInput struct {
	ID uuid.UUID
}

// Execute deletes an image block
func (uc *DeleteImageBlockUseCase) Execute(ctx context.Context, input DeleteImageBlockInput) error {
	// Delete references (will be handled by CASCADE, but explicit for clarity)
	if err := uc.imageBlockReferenceRepo.DeleteByImageBlock(ctx, input.ID); err != nil {
		uc.logger.Error("failed to delete image block references", "error", err)
		// Continue anyway
	}

	// Delete image block
	if err := uc.imageBlockRepo.Delete(ctx, input.ID); err != nil {
		uc.logger.Error("failed to delete image block", "error", err, "image_block_id", input.ID)
		return err
	}

	uc.logger.Info("image block deleted", "image_block_id", input.ID)

	return nil
}

