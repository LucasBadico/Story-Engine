package content_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListContentBlockReferencesByContentBlockUseCase handles listing references for a content block
type ListContentBlockReferencesByContentBlockUseCase struct {
	refRepo         repositories.ContentBlockReferenceRepository
	contentBlockRepo repositories.ContentBlockRepository
	logger          logger.Logger
}

// NewListContentBlockReferencesByContentBlockUseCase creates a new ListContentBlockReferencesByContentBlockUseCase
func NewListContentBlockReferencesByContentBlockUseCase(
	refRepo repositories.ContentBlockReferenceRepository,
	contentBlockRepo repositories.ContentBlockRepository,
	logger logger.Logger,
) *ListContentBlockReferencesByContentBlockUseCase {
	return &ListContentBlockReferencesByContentBlockUseCase{
		refRepo:         refRepo,
		contentBlockRepo: contentBlockRepo,
		logger:          logger,
	}
}

// ListContentBlockReferencesByContentBlockInput represents the input for listing references
type ListContentBlockReferencesByContentBlockInput struct {
	TenantID      uuid.UUID
	ContentBlockID uuid.UUID
}

// ListContentBlockReferencesByContentBlockOutput represents the output of listing references
type ListContentBlockReferencesByContentBlockOutput struct {
	References []*story.ContentBlockReference
	Total      int
}

// Execute lists references for a content block
func (uc *ListContentBlockReferencesByContentBlockUseCase) Execute(ctx context.Context, input ListContentBlockReferencesByContentBlockInput) (*ListContentBlockReferencesByContentBlockOutput, error) {
	// Validate content block exists
	_, err := uc.contentBlockRepo.GetByID(ctx, input.TenantID, input.ContentBlockID)
	if err != nil {
		return nil, err
	}

	references, err := uc.refRepo.ListByContentBlock(ctx, input.TenantID, input.ContentBlockID)
	if err != nil {
		uc.logger.Error("failed to list content block references", "error", err, "content_block_id", input.ContentBlockID, "tenant_id", input.TenantID)
		return nil, err
	}

	return &ListContentBlockReferencesByContentBlockOutput{
		References: references,
		Total:      len(references),
	}, nil
}

