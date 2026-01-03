package content_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListContentBlocksUseCase handles listing content blocks
type ListContentBlocksUseCase struct {
	contentBlockRepo repositories.ContentBlockRepository
	logger           logger.Logger
}

// NewListContentBlocksUseCase creates a new ListContentBlocksUseCase
func NewListContentBlocksUseCase(
	contentBlockRepo repositories.ContentBlockRepository,
	logger logger.Logger,
) *ListContentBlocksUseCase {
	return &ListContentBlocksUseCase{
		contentBlockRepo: contentBlockRepo,
		logger:           logger,
	}
}

// ListContentBlocksInput represents the input for listing content blocks
type ListContentBlocksInput struct {
	TenantID  uuid.UUID
	ChapterID uuid.UUID
}

// ListContentBlocksOutput represents the output of listing content blocks
type ListContentBlocksOutput struct {
	ContentBlocks []*story.ContentBlock
	Total         int
}

// Execute lists content blocks for a chapter
func (uc *ListContentBlocksUseCase) Execute(ctx context.Context, input ListContentBlocksInput) (*ListContentBlocksOutput, error) {
	contentBlocks, err := uc.contentBlockRepo.ListByChapter(ctx, input.TenantID, input.ChapterID)
	if err != nil {
		uc.logger.Error("failed to list content blocks", "error", err, "chapter_id", input.ChapterID, "tenant_id", input.TenantID)
		return nil, err
	}

	return &ListContentBlocksOutput{
		ContentBlocks: contentBlocks,
		Total:         len(contentBlocks),
	}, nil
}

