package inventory_item

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteInventoryItemUseCase handles inventory item deletion
type DeleteInventoryItemUseCase struct {
	itemRepo repositories.InventoryItemRepository
	logger   logger.Logger
}

// NewDeleteInventoryItemUseCase creates a new DeleteInventoryItemUseCase
func NewDeleteInventoryItemUseCase(
	itemRepo repositories.InventoryItemRepository,
	logger logger.Logger,
) *DeleteInventoryItemUseCase {
	return &DeleteInventoryItemUseCase{
		itemRepo: itemRepo,
		logger:   logger,
	}
}

// DeleteInventoryItemInput represents the input for deleting an inventory item
type DeleteInventoryItemInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes an inventory item
func (uc *DeleteInventoryItemUseCase) Execute(ctx context.Context, input DeleteInventoryItemInput) error {
	if err := uc.itemRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete inventory item", "error", err, "item_id", input.ID)
		return err
	}

	uc.logger.Info("inventory item deleted", "item_id", input.ID)

	return nil
}


