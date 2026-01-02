package inventory_item

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetInventoryItemUseCase handles retrieving an inventory item
type GetInventoryItemUseCase struct {
	itemRepo repositories.InventoryItemRepository
	logger   logger.Logger
}

// NewGetInventoryItemUseCase creates a new GetInventoryItemUseCase
func NewGetInventoryItemUseCase(
	itemRepo repositories.InventoryItemRepository,
	logger logger.Logger,
) *GetInventoryItemUseCase {
	return &GetInventoryItemUseCase{
		itemRepo: itemRepo,
		logger:   logger,
	}
}

// GetInventoryItemInput represents the input for getting an inventory item
type GetInventoryItemInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetInventoryItemOutput represents the output of getting an inventory item
type GetInventoryItemOutput struct {
	Item *rpg.InventoryItem
}

// Execute retrieves an inventory item by ID
func (uc *GetInventoryItemUseCase) Execute(ctx context.Context, input GetInventoryItemInput) (*GetInventoryItemOutput, error) {
	item, err := uc.itemRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get inventory item", "error", err, "item_id", input.ID)
		return nil, err
	}

	return &GetInventoryItemOutput{
		Item: item,
	}, nil
}


