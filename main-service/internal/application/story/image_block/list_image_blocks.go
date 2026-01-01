package image_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListImageBlocksUseCase handles listing image blocks for a chapter
type ListImageBlocksUseCase struct {
	imageBlockRepo repositories.ImageBlockRepository
	logger         logger.Logger
}

// NewListImageBlocksUseCase creates a new ListImageBlocksUseCase
func NewListImageBlocksUseCase(
	imageBlockRepo repositories.ImageBlockRepository,
	logger logger.Logger,
) *ListImageBlocksUseCase {
	return &ListImageBlocksUseCase{
		imageBlockRepo: imageBlockRepo,
		logger:         logger,
	}
}

// ListImageBlocksInput represents the input for listing image blocks
type ListImageBlocksInput struct {
	ChapterID uuid.UUID
}

// ListImageBlocksOutput represents the output of listing image blocks
type ListImageBlocksOutput struct {
	ImageBlocks []*story.ImageBlock
}

// Execute lists image blocks for a chapter
func (uc *ListImageBlocksUseCase) Execute(ctx context.Context, input ListImageBlocksInput) (*ListImageBlocksOutput, error) {
	imageBlocks, err := uc.imageBlockRepo.ListByChapter(ctx, input.ChapterID)
	if err != nil {
		uc.logger.Error("failed to list image blocks", "error", err, "chapter_id", input.ChapterID)
		return nil, err
	}

	return &ListImageBlocksOutput{
		ImageBlocks: imageBlocks,
	}, nil
}

