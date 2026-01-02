package prose_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListProseBlocksByEntityUseCase handles listing prose blocks by entity
type ListProseBlocksByEntityUseCase struct {
	refRepo        repositories.ProseBlockReferenceRepository
	proseBlockRepo repositories.ProseBlockRepository
	logger         logger.Logger
}

// NewListProseBlocksByEntityUseCase creates a new ListProseBlocksByEntityUseCase
func NewListProseBlocksByEntityUseCase(
	refRepo repositories.ProseBlockReferenceRepository,
	proseBlockRepo repositories.ProseBlockRepository,
	logger logger.Logger,
) *ListProseBlocksByEntityUseCase {
	return &ListProseBlocksByEntityUseCase{
		refRepo:        refRepo,
		proseBlockRepo: proseBlockRepo,
		logger:         logger,
	}
}

// ListProseBlocksByEntityInput represents the input for listing prose blocks by entity
type ListProseBlocksByEntityInput struct {
	TenantID   uuid.UUID
	EntityType story.EntityType
	EntityID   uuid.UUID
}

// ListProseBlocksByEntityOutput represents the output of listing prose blocks by entity
type ListProseBlocksByEntityOutput struct {
	ProseBlocks []*story.ProseBlock
	Total       int
}

// Execute lists prose blocks associated with an entity
func (uc *ListProseBlocksByEntityUseCase) Execute(ctx context.Context, input ListProseBlocksByEntityInput) (*ListProseBlocksByEntityOutput, error) {
	if input.EntityType == "" {
		return nil, &platformerrors.ValidationError{
			Field:   "entity_type",
			Message: "entity_type is required",
		}
	}

	references, err := uc.refRepo.ListByEntity(ctx, input.TenantID, input.EntityType, input.EntityID)
	if err != nil {
		uc.logger.Error("failed to list prose block references by entity", "error", err, "entity_type", input.EntityType, "entity_id", input.EntityID, "tenant_id", input.TenantID)
		return nil, err
	}

	// Get prose blocks for each reference
	proseBlocks := make([]*story.ProseBlock, 0, len(references))
	for _, ref := range references {
		proseBlock, err := uc.proseBlockRepo.GetByID(ctx, input.TenantID, ref.ProseBlockID)
		if err != nil {
			uc.logger.Error("failed to get prose block", "prose_block_id", ref.ProseBlockID, "error", err)
			continue
		}
		proseBlocks = append(proseBlocks, proseBlock)
	}

	return &ListProseBlocksByEntityOutput{
		ProseBlocks: proseBlocks,
		Total:       len(proseBlocks),
	}, nil
}

