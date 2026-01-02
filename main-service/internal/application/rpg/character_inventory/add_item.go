package character_inventory

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddItemToInventoryUseCase handles adding an item to character inventory
type AddItemToInventoryUseCase struct {
	inventoryRepo repositories.CharacterInventoryRepository
	characterRepo repositories.CharacterRepository
	itemRepo      repositories.InventoryItemRepository
	logger        logger.Logger
}

// NewAddItemToInventoryUseCase creates a new AddItemToInventoryUseCase
func NewAddItemToInventoryUseCase(
	inventoryRepo repositories.CharacterInventoryRepository,
	characterRepo repositories.CharacterRepository,
	itemRepo repositories.InventoryItemRepository,
	logger logger.Logger,
) *AddItemToInventoryUseCase {
	return &AddItemToInventoryUseCase{
		inventoryRepo: inventoryRepo,
		characterRepo: characterRepo,
		itemRepo:      itemRepo,
		logger:        logger,
	}
}

// AddItemToInventoryInput represents the input for adding an item
type AddItemToInventoryInput struct {
	TenantID    uuid.UUID
	CharacterID uuid.UUID
	ItemID      uuid.UUID
	Quantity    *int
	SlotID      *uuid.UUID
}

// AddItemToInventoryOutput represents the output of adding an item
type AddItemToInventoryOutput struct {
	Inventory *rpg.CharacterInventory
}

// Execute adds an item to character inventory
func (uc *AddItemToInventoryUseCase) Execute(ctx context.Context, input AddItemToInventoryInput) (*AddItemToInventoryOutput, error) {
	// Validate character exists
	_, err := uc.characterRepo.GetByID(ctx, input.TenantID, input.CharacterID)
	if err != nil {
		return nil, err
	}

	// Validate item exists
	item, err := uc.itemRepo.GetByID(ctx, input.TenantID, input.ItemID)
	if err != nil {
		return nil, err
	}

	// Check if item already exists in inventory (same item + slot)
	quantity := 1
	if input.Quantity != nil {
		quantity = *input.Quantity
	}

	existing, err := uc.inventoryRepo.GetByCharacterAndItem(ctx, input.TenantID, input.CharacterID, input.ItemID, input.SlotID)
	if err == nil && existing != nil {
		// Item exists, add to quantity if stackable
		if item.MaxStack > 1 {
			if err := existing.AddQuantity(quantity); err != nil {
				return nil, err
			}
			if err := uc.inventoryRepo.Update(ctx, existing); err != nil {
				return nil, err
			}
			return &AddItemToInventoryOutput{
				Inventory: existing,
			}, nil
		}
		// Not stackable, return existing
		return &AddItemToInventoryOutput{
			Inventory: existing,
		}, nil
	}

	// Create new inventory entry
	inventory, err := rpg.NewCharacterInventory(input.TenantID, input.CharacterID, input.ItemID)
	if err != nil {
		return nil, err
	}

	if err := inventory.SetQuantity(quantity); err != nil {
		return nil, err
	}
	if input.SlotID != nil {
		inventory.SetSlot(input.SlotID)
	}

	if err := inventory.Validate(); err != nil {
		return nil, err
	}

	if err := uc.inventoryRepo.Create(ctx, inventory); err != nil {
		uc.logger.Error("failed to add item to inventory", "error", err, "character_id", input.CharacterID, "item_id", input.ItemID)
		return nil, err
	}

	uc.logger.Info("item added to inventory", "character_id", input.CharacterID, "item_id", input.ItemID)

	return &AddItemToInventoryOutput{
		Inventory: inventory,
	}, nil
}


