package image_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetImageBlockUseCase handles retrieving an image block
type GetImageBlockUseCase struct {
	imageBlockRepo repositories.ImageBlockRepository
	logger         logger.Logger
}

// NewGetImageBlockUseCase creates a new GetImageBlockUseCase
func NewGetImageBlockUseCase(
	imageBlockRepo repositories.ImageBlockRepository,
	logger logger.Logger,
) *GetImageBlockUseCase {
	return &GetImageBlockUseCase{
		imageBlockRepo: imageBlockRepo,
		logger:         logger,
	}
}

// GetImageBlockInput represents the input for getting an image block
type GetImageBlockInput struct {
	ID uuid.UUID
}

// GetImageBlockOutput represents the output of getting an image block
type GetImageBlockOutput struct {
	ImageBlock *story.ImageBlock
}

// Execute retrieves an image block by ID
func (uc *GetImageBlockUseCase) Execute(ctx context.Context, input GetImageBlockInput) (*GetImageBlockOutput, error) {
	imageBlock, err := uc.imageBlockRepo.GetByID(ctx, input.ID)
	if err != nil {
		uc.logger.Error("failed to get image block", "error", err, "image_block_id", input.ID)
		return nil, err
	}

	return &GetImageBlockOutput{
		ImageBlock: imageBlock,
	}, nil
}


