package content_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetContentBlockUseCase handles retrieving a content block
type GetContentBlockUseCase struct {
	contentBlockRepo repositories.ContentBlockRepository
	logger           logger.Logger
}

// NewGetContentBlockUseCase creates a new GetContentBlockUseCase
func NewGetContentBlockUseCase(
	contentBlockRepo repositories.ContentBlockRepository,
	logger logger.Logger,
) *GetContentBlockUseCase {
	return &GetContentBlockUseCase{
		contentBlockRepo: contentBlockRepo,
		logger:           logger,
	}
}

// GetContentBlockInput represents the input for getting a content block
type GetContentBlockInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetContentBlockOutput represents the output of getting a content block
type GetContentBlockOutput struct {
	ContentBlock *story.ContentBlock
}

// Execute retrieves a content block by ID
func (uc *GetContentBlockUseCase) Execute(ctx context.Context, input GetContentBlockInput) (*GetContentBlockOutput, error) {
	contentBlock, err := uc.contentBlockRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get content block", "error", err, "content_block_id", input.ID, "tenant_id", input.TenantID)
		return nil, err
	}

	return &GetContentBlockOutput{
		ContentBlock: contentBlock,
	}, nil
}

