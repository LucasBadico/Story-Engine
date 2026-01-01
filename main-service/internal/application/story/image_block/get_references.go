package image_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetImageBlockReferencesUseCase handles getting references for an image block
type GetImageBlockReferencesUseCase struct {
	imageBlockReferenceRepo repositories.ImageBlockReferenceRepository
	logger                  logger.Logger
}

// NewGetImageBlockReferencesUseCase creates a new GetImageBlockReferencesUseCase
func NewGetImageBlockReferencesUseCase(
	imageBlockReferenceRepo repositories.ImageBlockReferenceRepository,
	logger logger.Logger,
) *GetImageBlockReferencesUseCase {
	return &GetImageBlockReferencesUseCase{
		imageBlockReferenceRepo: imageBlockReferenceRepo,
		logger:                  logger,
	}
}

// GetImageBlockReferencesInput represents the input for getting references
type GetImageBlockReferencesInput struct {
	ImageBlockID uuid.UUID
}

// GetImageBlockReferencesOutput represents the output of getting references
type GetImageBlockReferencesOutput struct {
	References []*story.ImageBlockReference
}

// Execute retrieves all references for an image block
func (uc *GetImageBlockReferencesUseCase) Execute(ctx context.Context, input GetImageBlockReferencesInput) (*GetImageBlockReferencesOutput, error) {
	references, err := uc.imageBlockReferenceRepo.ListByImageBlock(ctx, input.ImageBlockID)
	if err != nil {
		uc.logger.Error("failed to get image block references", "error", err, "image_block_id", input.ImageBlockID)
		return nil, err
	}

	return &GetImageBlockReferencesOutput{
		References: references,
	}, nil
}

