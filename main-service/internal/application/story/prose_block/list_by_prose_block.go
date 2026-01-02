package prose_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListProseBlockReferencesByProseBlockUseCase handles listing references for a prose block
type ListProseBlockReferencesByProseBlockUseCase struct {
	refRepo        repositories.ProseBlockReferenceRepository
	proseBlockRepo repositories.ProseBlockRepository
	logger         logger.Logger
}

// NewListProseBlockReferencesByProseBlockUseCase creates a new ListProseBlockReferencesByProseBlockUseCase
func NewListProseBlockReferencesByProseBlockUseCase(
	refRepo repositories.ProseBlockReferenceRepository,
	proseBlockRepo repositories.ProseBlockRepository,
	logger logger.Logger,
) *ListProseBlockReferencesByProseBlockUseCase {
	return &ListProseBlockReferencesByProseBlockUseCase{
		refRepo:        refRepo,
		proseBlockRepo: proseBlockRepo,
		logger:         logger,
	}
}

// ListProseBlockReferencesByProseBlockInput represents the input for listing references
type ListProseBlockReferencesByProseBlockInput struct {
	TenantID     uuid.UUID
	ProseBlockID uuid.UUID
}

// ListProseBlockReferencesByProseBlockOutput represents the output of listing references
type ListProseBlockReferencesByProseBlockOutput struct {
	References []*story.ProseBlockReference
	Total      int
}

// Execute lists references for a prose block
func (uc *ListProseBlockReferencesByProseBlockUseCase) Execute(ctx context.Context, input ListProseBlockReferencesByProseBlockInput) (*ListProseBlockReferencesByProseBlockOutput, error) {
	// Validate prose block exists
	_, err := uc.proseBlockRepo.GetByID(ctx, input.TenantID, input.ProseBlockID)
	if err != nil {
		return nil, err
	}

	references, err := uc.refRepo.ListByProseBlock(ctx, input.TenantID, input.ProseBlockID)
	if err != nil {
		uc.logger.Error("failed to list prose block references", "error", err, "prose_block_id", input.ProseBlockID, "tenant_id", input.TenantID)
		return nil, err
	}

	return &ListProseBlockReferencesByProseBlockOutput{
		References: references,
		Total:      len(references),
	}, nil
}

