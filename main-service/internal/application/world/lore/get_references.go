package lore

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetReferencesUseCase handles retrieving references for a lore
type GetReferencesUseCase struct {
	loreReferenceRepo repositories.LoreReferenceRepository
	logger            logger.Logger
}

// NewGetReferencesUseCase creates a new GetReferencesUseCase
func NewGetReferencesUseCase(
	loreReferenceRepo repositories.LoreReferenceRepository,
	logger logger.Logger,
) *GetReferencesUseCase {
	return &GetReferencesUseCase{
		loreReferenceRepo: loreReferenceRepo,
		logger:            logger,
	}
}

// GetReferencesInput represents the input for getting references
type GetReferencesInput struct {
	TenantID uuid.UUID
	LoreID   uuid.UUID
}

// GetReferencesOutput represents the output of getting references
type GetReferencesOutput struct {
	References []*world.LoreReference
}

// Execute retrieves references for a lore
func (uc *GetReferencesUseCase) Execute(ctx context.Context, input GetReferencesInput) (*GetReferencesOutput, error) {
	references, err := uc.loreReferenceRepo.ListByLore(ctx, input.TenantID, input.LoreID)
	if err != nil {
		uc.logger.Error("failed to get lore references", "error", err, "lore_id", input.LoreID)
		return nil, err
	}

	return &GetReferencesOutput{
		References: references,
	}, nil
}

