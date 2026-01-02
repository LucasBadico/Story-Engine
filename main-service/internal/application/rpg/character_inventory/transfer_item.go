package character_inventory

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// TransferItemUseCase handles transferring an item between characters
type TransferItemUseCase struct {
	inventoryRepo repositories.CharacterInventoryRepository
	characterRepo repositories.CharacterRepository
	logger        logger.Logger
}

// NewTransferItemUseCase creates a new TransferItemUseCase
func NewTransferItemUseCase(
	inventoryRepo repositories.CharacterInventoryRepository,
	characterRepo repositories.CharacterRepository,
	logger logger.Logger,
) *TransferItemUseCase {
	return &TransferItemUseCase{
		inventoryRepo: inventoryRepo,
		characterRepo: characterRepo,
		logger:        logger,
	}
}

// TransferItemInput represents the input for transferring an item
type TransferItemInput struct {
	TenantID      uuid.UUID
	InventoryID   uuid.UUID
	ToCharacterID uuid.UUID
	Quantity      *int // optional: transfer partial quantity
}

// TransferItemOutput represents the output of transferring an item
type TransferItemOutput struct {
	FromInventory *rpg.CharacterInventory
	ToInventory   *rpg.CharacterInventory
}

// Execute transfers an item from one character to another
func (uc *TransferItemUseCase) Execute(ctx context.Context, input TransferItemInput) (*TransferItemOutput, error) {
	// Get source inventory entry
	fromInventory, err := uc.inventoryRepo.GetByID(ctx, input.TenantID, input.InventoryID)
	if err != nil {
		return nil, err
	}

	// Validate target character exists
	_, err = uc.characterRepo.GetByID(ctx, input.TenantID, input.ToCharacterID)
	if err != nil {
		return nil, err
	}

	// Determine quantity to transfer
	transferQuantity := fromInventory.Quantity
	if input.Quantity != nil && *input.Quantity < fromInventory.Quantity {
		transferQuantity = *input.Quantity
	}

	// Check if target character already has this item
	toInventory, err := uc.inventoryRepo.GetByCharacterAndItem(ctx, input.TenantID, input.ToCharacterID, fromInventory.ItemID, nil)
	if err == nil && toInventory != nil {
		// Target has item, add quantity
		if err := toInventory.AddQuantity(transferQuantity); err != nil {
			return nil, err
		}
		if err := uc.inventoryRepo.Update(ctx, toInventory); err != nil {
			return nil, err
		}
	} else {
		// Create new inventory entry for target
		toInventory, err = rpg.NewCharacterInventory(input.TenantID, input.ToCharacterID, fromInventory.ItemID)
		if err != nil {
			return nil, err
		}
		if err := toInventory.SetQuantity(transferQuantity); err != nil {
			return nil, err
		}
		if err := uc.inventoryRepo.Create(ctx, toInventory); err != nil {
			return nil, err
		}
	}

	// Update source inventory
	fromCharacterID := fromInventory.CharacterID
	itemID := fromInventory.ItemID
	if transferQuantity >= fromInventory.Quantity {
		// Transfer all, delete source
		if err := uc.inventoryRepo.Delete(ctx, input.TenantID, fromInventory.ID); err != nil {
			return nil, err
		}
		fromInventory = nil
	} else {
		// Transfer partial, reduce quantity
		if err := fromInventory.SetQuantity(fromInventory.Quantity - transferQuantity); err != nil {
			return nil, err
		}
		if err := uc.inventoryRepo.Update(ctx, fromInventory); err != nil {
			return nil, err
		}
	}

	uc.logger.Info("item transferred", "from_character_id", fromCharacterID, "to_character_id", input.ToCharacterID, "item_id", itemID)

	return &TransferItemOutput{
		FromInventory: fromInventory,
		ToInventory:   toInventory,
	}, nil
}

