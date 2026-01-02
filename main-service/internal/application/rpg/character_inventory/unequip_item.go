package character_inventory

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UnequipItemUseCase handles unequipping an item
type UnequipItemUseCase struct {
	inventoryRepo repositories.CharacterInventoryRepository
	logger        logger.Logger
}

// NewUnequipItemUseCase creates a new UnequipItemUseCase
func NewUnequipItemUseCase(
	inventoryRepo repositories.CharacterInventoryRepository,
	logger logger.Logger,
) *UnequipItemUseCase {
	return &UnequipItemUseCase{
		inventoryRepo: inventoryRepo,
		logger:        logger,
	}
}

// UnequipItemInput represents the input for unequipping an item
type UnequipItemInput struct {
	ID uuid.UUID
}

// UnequipItemOutput represents the output of unequipping an item
type UnequipItemOutput struct {
	Inventory *rpg.CharacterInventory
}

// Execute unequips an item
func (uc *UnequipItemUseCase) Execute(ctx context.Context, input UnequipItemInput) (*UnequipItemOutput, error) {
	// Get inventory entry
	inventory, err := uc.inventoryRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// Unequip item
	inventory.SetEquipped(false)

	if err := uc.inventoryRepo.Update(ctx, inventory); err != nil {
		uc.logger.Error("failed to unequip item", "error", err, "id", input.ID)
		return nil, err
	}

	uc.logger.Info("item unequipped", "id", input.ID, "character_id", inventory.CharacterID)

	return &UnequipItemOutput{
		Inventory: inventory,
	}, nil
}

