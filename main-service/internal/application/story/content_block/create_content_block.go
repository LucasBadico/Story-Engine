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

// CreateContentBlockUseCase handles content block creation
type CreateContentBlockUseCase struct {
	contentBlockRepo repositories.ContentBlockRepository
	chapterRepo      repositories.ChapterRepository
	ingestionQueue   queue.IngestionQueue
	logger           logger.Logger
}

// NewCreateContentBlockUseCase creates a new CreateContentBlockUseCase
func NewCreateContentBlockUseCase(
	contentBlockRepo repositories.ContentBlockRepository,
	chapterRepo repositories.ChapterRepository,
	ingestionQueue queue.IngestionQueue,
	logger logger.Logger,
) *CreateContentBlockUseCase {
	return &CreateContentBlockUseCase{
		contentBlockRepo: contentBlockRepo,
		chapterRepo:      chapterRepo,
		ingestionQueue:   ingestionQueue,
		logger:           logger,
	}
}

// CreateContentBlockInput represents the input for creating a content block
type CreateContentBlockInput struct {
	TenantID  uuid.UUID
	ChapterID *uuid.UUID
	OrderNum  *int
	Type      story.ContentType
	Kind      story.ContentKind
	Content   string
	Metadata  story.ContentMetadata
}

// CreateContentBlockOutput represents the output of creating a content block
type CreateContentBlockOutput struct {
	ContentBlock *story.ContentBlock
}

// Execute creates a new content block
func (uc *CreateContentBlockUseCase) Execute(ctx context.Context, input CreateContentBlockInput) (*CreateContentBlockOutput, error) {
	// Validate chapter exists if provided
	if input.ChapterID != nil {
		_, err := uc.chapterRepo.GetByID(ctx, input.TenantID, *input.ChapterID)
		if err != nil {
			return nil, err
		}
	}

	if input.OrderNum != nil && *input.OrderNum < 1 {
		return nil, &platformerrors.ValidationError{
			Field:   "order_num",
			Message: "must be greater than 0",
		}
	}

	contentType := input.Type
	if contentType == "" {
		contentType = story.ContentTypeText
	}

	kind := input.Kind
	if kind == "" {
		kind = story.ContentKindFinal
	}

	if input.Content == "" {
		return nil, &platformerrors.ValidationError{
			Field:   "content",
			Message: "content is required",
		}
	}

	contentBlock, err := story.NewContentBlock(input.TenantID, input.ChapterID, input.OrderNum, contentType, kind, input.Content)
	if err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "content_block",
			Message: err.Error(),
		}
	}

	// Update metadata if provided
	if input.Metadata.WordCount != nil || input.Metadata.AltText != nil {
		if err := contentBlock.UpdateMetadata(input.Metadata); err != nil {
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

	if err := uc.contentBlockRepo.Create(ctx, contentBlock); err != nil {
		uc.logger.Error("failed to create content block", "error", err, "tenant_id", input.TenantID)
		return nil, err
	}

	uc.logger.Info("content block created", "content_block_id", contentBlock.ID, "tenant_id", input.TenantID)
	uc.enqueueIngestion(ctx, input.TenantID, contentBlock.ID)

	return &CreateContentBlockOutput{
		ContentBlock: contentBlock,
	}, nil
}

func (uc *CreateContentBlockUseCase) enqueueIngestion(ctx context.Context, tenantID uuid.UUID, contentBlockID uuid.UUID) {
	if uc.ingestionQueue == nil {
		return
	}
	if err := uc.ingestionQueue.Push(ctx, tenantID, "content_block", contentBlockID); err != nil {
		uc.logger.Error("failed to enqueue content block ingestion", "error", err, "content_block_id", contentBlockID, "tenant_id", tenantID)
	}
}
