package content_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/queue"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateContentBlockUseCase handles content block updates
type UpdateContentBlockUseCase struct {
	contentBlockRepo repositories.ContentBlockRepository
	ingestionQueue   queue.IngestionQueue
	logger           logger.Logger
}

// NewUpdateContentBlockUseCase creates a new UpdateContentBlockUseCase
func NewUpdateContentBlockUseCase(
	contentBlockRepo repositories.ContentBlockRepository,
	ingestionQueue queue.IngestionQueue,
	logger logger.Logger,
) *UpdateContentBlockUseCase {
	return &UpdateContentBlockUseCase{
		contentBlockRepo: contentBlockRepo,
		ingestionQueue:   ingestionQueue,
		logger:           logger,
	}
}

// UpdateContentBlockInput represents the input for updating a content block
type UpdateContentBlockInput struct {
	TenantID  uuid.UUID
	ID        uuid.UUID
	OrderNum  *int
	Type      *story.ContentType
	Kind      *story.ContentKind
	Content   *string
	Metadata  *story.ContentMetadata
}

// UpdateContentBlockOutput represents the output of updating a content block
type UpdateContentBlockOutput struct {
	ContentBlock *story.ContentBlock
}

// Execute updates a content block
func (uc *UpdateContentBlockUseCase) Execute(ctx context.Context, input UpdateContentBlockInput) (*UpdateContentBlockOutput, error) {
	// Get existing content block
	contentBlock, err := uc.contentBlockRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.OrderNum != nil {
		if *input.OrderNum < 1 {
			return nil, &platformerrors.ValidationError{
				Field:   "order_num",
				Message: "must be greater than 0",
			}
		}
		contentBlock.OrderNum = input.OrderNum
	}

	if input.Type != nil {
		contentBlock.Type = *input.Type
	}

	if input.Kind != nil {
		contentBlock.Kind = *input.Kind
	}

	if input.Content != nil {
		if err := contentBlock.UpdateContent(*input.Content); err != nil {
			return nil, &platformerrors.ValidationError{
				Field:   "content",
				Message: err.Error(),
			}
		}
	}

	if input.Metadata != nil {
		if err := contentBlock.UpdateMetadata(*input.Metadata); err != nil {
			return nil, &platformerrors.ValidationError{
				Field:   "metadata",
				Message: err.Error(),
			}
		}
	}

	if err := contentBlock.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "content_block",
			Message: err.Error(),
		}
	}

	if err := uc.contentBlockRepo.Update(ctx, contentBlock); err != nil {
		uc.logger.Error("failed to update content block", "error", err, "content_block_id", input.ID, "tenant_id", input.TenantID)
		return nil, err
	}

	uc.logger.Info("content block updated", "content_block_id", input.ID, "tenant_id", input.TenantID)
	uc.enqueueIngestion(ctx, input.TenantID, contentBlock.ID)

	return &UpdateContentBlockOutput{
		ContentBlock: contentBlock,
	}, nil
}

func (uc *UpdateContentBlockUseCase) enqueueIngestion(ctx context.Context, tenantID uuid.UUID, contentBlockID uuid.UUID) {
	if uc.ingestionQueue == nil {
		return
	}
	if err := uc.ingestionQueue.Push(ctx, tenantID, "content_block", contentBlockID); err != nil {
		uc.logger.Error("failed to enqueue content block ingestion", "error", err, "content_block_id", contentBlockID, "tenant_id", tenantID)
	}
}
