package prose_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetProseBlockUseCase handles retrieving a prose block
type GetProseBlockUseCase struct {
	proseBlockRepo repositories.ProseBlockRepository
	logger         logger.Logger
}

// NewGetProseBlockUseCase creates a new GetProseBlockUseCase
func NewGetProseBlockUseCase(
	proseBlockRepo repositories.ProseBlockRepository,
	logger logger.Logger,
) *GetProseBlockUseCase {
	return &GetProseBlockUseCase{
		proseBlockRepo: proseBlockRepo,
		logger:         logger,
	}
}

// GetProseBlockInput represents the input for getting a prose block
type GetProseBlockInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetProseBlockOutput represents the output of getting a prose block
type GetProseBlockOutput struct {
	ProseBlock *story.ProseBlock
}

// Execute retrieves a prose block by ID
func (uc *GetProseBlockUseCase) Execute(ctx context.Context, input GetProseBlockInput) (*GetProseBlockOutput, error) {
	proseBlock, err := uc.proseBlockRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get prose block", "error", err, "prose_block_id", input.ID, "tenant_id", input.TenantID)
		return nil, err
	}

	return &GetProseBlockOutput{
		ProseBlock: proseBlock,
	}, nil
}

