package character_inventory

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// EquipItemUseCase handles equipping an item
type EquipItemUseCase struct {
	inventoryRepo repositories.CharacterInventoryRepository
	logger        logger.Logger
}

// NewEquipItemUseCase creates a new EquipItemUseCase
func NewEquipItemUseCase(
	inventoryRepo repositories.CharacterInventoryRepository,
	logger logger.Logger,
) *EquipItemUseCase {
	return &EquipItemUseCase{
		inventoryRepo: inventoryRepo,
		logger:        logger,
	}
}

// EquipItemInput represents the input for equipping an item
type EquipItemInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// EquipItemOutput represents the output of equipping an item
type EquipItemOutput struct {
	Inventory *rpg.CharacterInventory
}

// Execute equips an item
func (uc *EquipItemUseCase) Execute(ctx context.Context, input EquipItemInput) (*EquipItemOutput, error) {
	// Get inventory entry
	inventory, err := uc.inventoryRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return nil, err
	}

	// Equip item
	inventory.SetEquipped(true)

	if err := uc.inventoryRepo.Update(ctx, inventory); err != nil {
		uc.logger.Error("failed to equip item", "error", err, "id", input.ID)
		return nil, err
	}

	uc.logger.Info("item equipped", "id", input.ID, "character_id", inventory.CharacterID)

	return &EquipItemOutput{
		Inventory: inventory,
	}, nil
}


