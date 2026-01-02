package prose_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateProseBlockUseCase handles prose block creation
type CreateProseBlockUseCase struct {
	proseBlockRepo repositories.ProseBlockRepository
	chapterRepo    repositories.ChapterRepository
	logger         logger.Logger
}

// NewCreateProseBlockUseCase creates a new CreateProseBlockUseCase
func NewCreateProseBlockUseCase(
	proseBlockRepo repositories.ProseBlockRepository,
	chapterRepo repositories.ChapterRepository,
	logger logger.Logger,
) *CreateProseBlockUseCase {
	return &CreateProseBlockUseCase{
		proseBlockRepo: proseBlockRepo,
		chapterRepo:    chapterRepo,
		logger:         logger,
	}
}

// CreateProseBlockInput represents the input for creating a prose block
type CreateProseBlockInput struct {
	TenantID  uuid.UUID
	ChapterID *uuid.UUID
	OrderNum  *int
	Kind      story.ProseKind
	Content   string
}

// CreateProseBlockOutput represents the output of creating a prose block
type CreateProseBlockOutput struct {
	ProseBlock *story.ProseBlock
}

// Execute creates a new prose block
func (uc *CreateProseBlockUseCase) Execute(ctx context.Context, input CreateProseBlockInput) (*CreateProseBlockOutput, error) {
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

	kind := input.Kind
	if kind == "" {
		kind = story.ProseKindFinal
	}

	if input.Content == "" {
		return nil, &platformerrors.ValidationError{
			Field:   "content",
			Message: "content is required",
		}
	}

	proseBlock, err := story.NewProseBlock(input.TenantID, input.ChapterID, input.OrderNum, kind, input.Content)
	if err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "prose_block",
			Message: err.Error(),
		}
	}

	if err := proseBlock.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "prose_block",
			Message: err.Error(),
		}
	}

	if err := uc.proseBlockRepo.Create(ctx, proseBlock); err != nil {
		uc.logger.Error("failed to create prose block", "error", err, "tenant_id", input.TenantID)
		return nil, err
	}

	uc.logger.Info("prose block created", "prose_block_id", proseBlock.ID, "tenant_id", input.TenantID)

	return &CreateProseBlockOutput{
		ProseBlock: proseBlock,
	}, nil
}

