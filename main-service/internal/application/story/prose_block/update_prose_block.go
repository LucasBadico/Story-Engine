package prose_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateProseBlockUseCase handles prose block updates
type UpdateProseBlockUseCase struct {
	proseBlockRepo repositories.ProseBlockRepository
	logger         logger.Logger
}

// NewUpdateProseBlockUseCase creates a new UpdateProseBlockUseCase
func NewUpdateProseBlockUseCase(
	proseBlockRepo repositories.ProseBlockRepository,
	logger logger.Logger,
) *UpdateProseBlockUseCase {
	return &UpdateProseBlockUseCase{
		proseBlockRepo: proseBlockRepo,
		logger:         logger,
	}
}

// UpdateProseBlockInput represents the input for updating a prose block
type UpdateProseBlockInput struct {
	TenantID  uuid.UUID
	ID        uuid.UUID
	OrderNum  *int
	Kind      *story.ProseKind
	Content   *string
}

// UpdateProseBlockOutput represents the output of updating a prose block
type UpdateProseBlockOutput struct {
	ProseBlock *story.ProseBlock
}

// Execute updates a prose block
func (uc *UpdateProseBlockUseCase) Execute(ctx context.Context, input UpdateProseBlockInput) (*UpdateProseBlockOutput, error) {
	// Get existing prose block
	proseBlock, err := uc.proseBlockRepo.GetByID(ctx, input.TenantID, input.ID)
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
		proseBlock.OrderNum = input.OrderNum
	}

	if input.Kind != nil {
		proseBlock.Kind = *input.Kind
	}

	if input.Content != nil {
		proseBlock.UpdateContent(*input.Content)
	}

	if err := proseBlock.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "prose_block",
			Message: err.Error(),
		}
	}

	if err := uc.proseBlockRepo.Update(ctx, proseBlock); err != nil {
		uc.logger.Error("failed to update prose block", "error", err, "prose_block_id", input.ID, "tenant_id", input.TenantID)
		return nil, err
	}

	uc.logger.Info("prose block updated", "prose_block_id", input.ID, "tenant_id", input.TenantID)

	return &UpdateProseBlockOutput{
		ProseBlock: proseBlock,
	}, nil
}

