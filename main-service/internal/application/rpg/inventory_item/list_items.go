package inventory_item

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListInventoryItemsUseCase handles listing inventory items
type ListInventoryItemsUseCase struct {
	itemRepo repositories.InventoryItemRepository
	logger   logger.Logger
}

// NewListInventoryItemsUseCase creates a new ListInventoryItemsUseCase
func NewListInventoryItemsUseCase(
	itemRepo repositories.InventoryItemRepository,
	logger logger.Logger,
) *ListInventoryItemsUseCase {
	return &ListInventoryItemsUseCase{
		itemRepo: itemRepo,
		logger:   logger,
	}
}

// ListInventoryItemsInput represents the input for listing items
type ListInventoryItemsInput struct {
	RPGSystemID uuid.UUID
	ArtifactID  *uuid.UUID // optional: filter by artifact
}

// ListInventoryItemsOutput represents the output of listing items
type ListInventoryItemsOutput struct {
	Items []*rpg.InventoryItem
}

// Execute lists inventory items
func (uc *ListInventoryItemsUseCase) Execute(ctx context.Context, input ListInventoryItemsInput) (*ListInventoryItemsOutput, error) {
	var items []*rpg.InventoryItem
	var err error

	if input.ArtifactID != nil {
		items, err = uc.itemRepo.ListByArtifact(ctx, *input.ArtifactID)
	} else {
		items, err = uc.itemRepo.ListBySystem(ctx, input.RPGSystemID)
	}

	if err != nil {
		uc.logger.Error("failed to list inventory items", "error", err, "rpg_system_id", input.RPGSystemID)
		return nil, err
	}

	return &ListInventoryItemsOutput{
		Items: items,
	}, nil
}


