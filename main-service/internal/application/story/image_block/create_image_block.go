package image_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateImageBlockUseCase handles image block creation
type CreateImageBlockUseCase struct {
	imageBlockRepo repositories.ImageBlockRepository
	chapterRepo    repositories.ChapterRepository
	logger         logger.Logger
}

// NewCreateImageBlockUseCase creates a new CreateImageBlockUseCase
func NewCreateImageBlockUseCase(
	imageBlockRepo repositories.ImageBlockRepository,
	chapterRepo repositories.ChapterRepository,
	logger logger.Logger,
) *CreateImageBlockUseCase {
	return &CreateImageBlockUseCase{
		imageBlockRepo: imageBlockRepo,
		chapterRepo:    chapterRepo,
		logger:         logger,
	}
}

// CreateImageBlockInput represents the input for creating an image block
type CreateImageBlockInput struct {
	TenantID  uuid.UUID
	ChapterID *uuid.UUID
	OrderNum  *int
	Kind      story.ImageKind
	ImageURL  string
	AltText   *string
	Caption   *string
	Width     *int
	Height    *int
}

// CreateImageBlockOutput represents the output of creating an image block
type CreateImageBlockOutput struct {
	ImageBlock *story.ImageBlock
}

// Execute creates a new image block
func (uc *CreateImageBlockUseCase) Execute(ctx context.Context, input CreateImageBlockInput) (*CreateImageBlockOutput, error) {
	// Validate chapter exists if provided
	if input.ChapterID != nil {
		_, err := uc.chapterRepo.GetByID(ctx, input.TenantID, *input.ChapterID)
		if err != nil {
			return nil, err
		}
	}

	// Create image block
	ib, err := story.NewImageBlock(input.TenantID, input.ChapterID, input.OrderNum, input.Kind, input.ImageURL)
	if err != nil {
		return nil, err
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

	if err := uc.imageBlockRepo.Create(ctx, ib); err != nil {
		uc.logger.Error("failed to create image block", "error", err)
		return nil, err
	}

	uc.logger.Info("image block created", "image_block_id", ib.ID)

	return &CreateImageBlockOutput{
		ImageBlock: ib,
	}, nil
}


