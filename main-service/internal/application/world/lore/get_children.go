package lore

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetChildrenUseCase handles getting children of a lore
type GetChildrenUseCase struct {
	loreRepo repositories.LoreRepository
	logger   logger.Logger
}

// NewGetChildrenUseCase creates a new GetChildrenUseCase
func NewGetChildrenUseCase(
	loreRepo repositories.LoreRepository,
	logger logger.Logger,
) *GetChildrenUseCase {
	return &GetChildrenUseCase{
		loreRepo: loreRepo,
		logger:   logger,
	}
}

// GetChildrenInput represents the input for getting children
type GetChildrenInput struct {
	TenantID uuid.UUID
	LoreID   uuid.UUID
}

// GetChildrenOutput represents the output of getting children
type GetChildrenOutput struct {
	Children []*world.Lore
}

// Execute retrieves direct children of a lore
func (uc *GetChildrenUseCase) Execute(ctx context.Context, input GetChildrenInput) (*GetChildrenOutput, error) {
	children, err := uc.loreRepo.GetChildren(ctx, input.TenantID, input.LoreID)
	if err != nil {
		uc.logger.Error("failed to get children", "error", err, "lore_id", input.LoreID)
		return nil, err
	}

	return &GetChildrenOutput{
		Children: children,
	}, nil
}

