package content_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListContentAnchorsByContentBlockUseCase handles listing anchors for a content block
type ListContentAnchorsByContentBlockUseCase struct {
	anchorRepo      repositories.ContentAnchorRepository
	contentBlockRepo repositories.ContentBlockRepository
	logger          logger.Logger
}

// NewListContentAnchorsByContentBlockUseCase creates a new ListContentAnchorsByContentBlockUseCase
func NewListContentAnchorsByContentBlockUseCase(
	anchorRepo repositories.ContentAnchorRepository,
	contentBlockRepo repositories.ContentBlockRepository,
	logger logger.Logger,
) *ListContentAnchorsByContentBlockUseCase {
	return &ListContentAnchorsByContentBlockUseCase{
		anchorRepo:      anchorRepo,
		contentBlockRepo: contentBlockRepo,
		logger:          logger,
	}
}

// ListContentAnchorsByContentBlockInput represents the input for listing anchors
type ListContentAnchorsByContentBlockInput struct {
	TenantID      uuid.UUID
	ContentBlockID uuid.UUID
}

// ListContentAnchorsByContentBlockOutput represents the output of listing anchors
type ListContentAnchorsByContentBlockOutput struct {
	Anchors []*story.ContentAnchor
	Total      int
}

// Execute lists anchors for a content block
func (uc *ListContentAnchorsByContentBlockUseCase) Execute(ctx context.Context, input ListContentAnchorsByContentBlockInput) (*ListContentAnchorsByContentBlockOutput, error) {
	// Validate content block exists
	_, err := uc.contentBlockRepo.GetByID(ctx, input.TenantID, input.ContentBlockID)
	if err != nil {
		return nil, err
	}

	anchors, err := uc.anchorRepo.ListByContentBlock(ctx, input.TenantID, input.ContentBlockID)
	if err != nil {
		uc.logger.Error("failed to list content anchors", "error", err, "content_block_id", input.ContentBlockID, "tenant_id", input.TenantID)
		return nil, err
	}

	return &ListContentAnchorsByContentBlockOutput{
		Anchors: anchors,
		Total:   len(anchors),
	}, nil
}

