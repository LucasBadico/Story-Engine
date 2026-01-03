package content_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListContentBlocksByEntityUseCase handles listing content blocks by entity
type ListContentBlocksByEntityUseCase struct {
	refRepo         repositories.ContentBlockReferenceRepository
	contentBlockRepo repositories.ContentBlockRepository
	logger          logger.Logger
}

// NewListContentBlocksByEntityUseCase creates a new ListContentBlocksByEntityUseCase
func NewListContentBlocksByEntityUseCase(
	refRepo repositories.ContentBlockReferenceRepository,
	contentBlockRepo repositories.ContentBlockRepository,
	logger logger.Logger,
) *ListContentBlocksByEntityUseCase {
	return &ListContentBlocksByEntityUseCase{
		refRepo:         refRepo,
		contentBlockRepo: contentBlockRepo,
		logger:          logger,
	}
}

// ListContentBlocksByEntityInput represents the input for listing content blocks by entity
type ListContentBlocksByEntityInput struct {
	TenantID   uuid.UUID
	EntityType story.EntityType
	EntityID   uuid.UUID
}

// ListContentBlocksByEntityOutput represents the output of listing content blocks by entity
type ListContentBlocksByEntityOutput struct {
	ContentBlocks []*story.ContentBlock
	Total         int
}

// Execute lists content blocks associated with an entity
func (uc *ListContentBlocksByEntityUseCase) Execute(ctx context.Context, input ListContentBlocksByEntityInput) (*ListContentBlocksByEntityOutput, error) {
	if input.EntityType == "" {
		return nil, &platformerrors.ValidationError{
			Field:   "entity_type",
			Message: "entity_type is required",
		}
	}

	references, err := uc.refRepo.ListByEntity(ctx, input.TenantID, input.EntityType, input.EntityID)
	if err != nil {
		uc.logger.Error("failed to list content block references by entity", "error", err, "entity_type", input.EntityType, "entity_id", input.EntityID, "tenant_id", input.TenantID)
		return nil, err
	}

	// Get content blocks for each reference
	contentBlocks := make([]*story.ContentBlock, 0, len(references))
	for _, ref := range references {
		contentBlock, err := uc.contentBlockRepo.GetByID(ctx, input.TenantID, ref.ContentBlockID)
		if err != nil {
			uc.logger.Error("failed to get content block", "content_block_id", ref.ContentBlockID, "error", err)
			continue
		}
		contentBlocks = append(contentBlocks, contentBlock)
	}

	return &ListContentBlocksByEntityOutput{
		ContentBlocks: contentBlocks,
		Total:         len(contentBlocks),
	}, nil
}

