package image_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateImageBlockUseCase handles image block updates
type UpdateImageBlockUseCase struct {
	imageBlockRepo repositories.ImageBlockRepository
	logger         logger.Logger
}

// NewUpdateImageBlockUseCase creates a new UpdateImageBlockUseCase
func NewUpdateImageBlockUseCase(
	imageBlockRepo repositories.ImageBlockRepository,
	logger logger.Logger,
) *UpdateImageBlockUseCase {
	return &UpdateImageBlockUseCase{
		imageBlockRepo: imageBlockRepo,
		logger:         logger,
	}
}

// UpdateImageBlockInput represents the input for updating an image block
type UpdateImageBlockInput struct {
	ID        uuid.UUID
	ChapterID *uuid.UUID
	OrderNum  *int
	Kind      *story.ImageKind
	ImageURL  *string
	AltText   *string
	Caption   *string
	Width     *int
	Height    *int
}

// UpdateImageBlockOutput represents the output of updating an image block
type UpdateImageBlockOutput struct {
	ImageBlock *story.ImageBlock
}

// Execute updates an image block
func (uc *UpdateImageBlockUseCase) Execute(ctx context.Context, input UpdateImageBlockInput) (*UpdateImageBlockOutput, error) {
	// Get existing image block
	ib, err := uc.imageBlockRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.ChapterID != nil {
		ib.ChapterID = input.ChapterID
	}
	if input.OrderNum != nil {
		ib.OrderNum = input.OrderNum
	}
	if input.Kind != nil {
		ib.Kind = *input.Kind
	}
	if input.ImageURL != nil {
		if err := ib.UpdateImageURL(*input.ImageURL); err != nil {
			return nil, err
		}
	}
	if input.AltText != nil {
		ib.UpdateAltText(input.AltText)
	}
	if input.Caption != nil {
		ib.UpdateCaption(input.Caption)
	}
	if input.Width != nil || input.Height != nil {
		if err := ib.UpdateDimensions(input.Width, input.Height); err != nil {
			return nil, err
		}
	}

	if err := ib.Validate(); err != nil {
		return nil, err
	}

	if err := uc.imageBlockRepo.Update(ctx, ib); err != nil {
		uc.logger.Error("failed to update image block", "error", err, "image_block_id", input.ID)
		return nil, err
	}

	uc.logger.Info("image block updated", "image_block_id", input.ID)

	return &UpdateImageBlockOutput{
		ImageBlock: ib,
	}, nil
}

