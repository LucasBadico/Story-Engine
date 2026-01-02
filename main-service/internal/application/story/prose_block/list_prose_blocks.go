package prose_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListProseBlocksUseCase handles listing prose blocks
type ListProseBlocksUseCase struct {
	proseBlockRepo repositories.ProseBlockRepository
	logger         logger.Logger
}

// NewListProseBlocksUseCase creates a new ListProseBlocksUseCase
func NewListProseBlocksUseCase(
	proseBlockRepo repositories.ProseBlockRepository,
	logger logger.Logger,
) *ListProseBlocksUseCase {
	return &ListProseBlocksUseCase{
		proseBlockRepo: proseBlockRepo,
		logger:         logger,
	}
}

// ListProseBlocksInput represents the input for listing prose blocks
type ListProseBlocksInput struct {
	TenantID  uuid.UUID
	ChapterID uuid.UUID
}

// ListProseBlocksOutput represents the output of listing prose blocks
type ListProseBlocksOutput struct {
	ProseBlocks []*story.ProseBlock
	Total       int
}

// Execute lists prose blocks for a chapter
func (uc *ListProseBlocksUseCase) Execute(ctx context.Context, input ListProseBlocksInput) (*ListProseBlocksOutput, error) {
	proseBlocks, err := uc.proseBlockRepo.ListByChapter(ctx, input.TenantID, input.ChapterID)
	if err != nil {
		uc.logger.Error("failed to list prose blocks", "error", err, "chapter_id", input.ChapterID, "tenant_id", input.TenantID)
		return nil, err
	}

	return &ListProseBlocksOutput{
		ProseBlocks: proseBlocks,
		Total:       len(proseBlocks),
	}, nil
}

