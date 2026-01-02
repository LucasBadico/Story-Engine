package character_inventory

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateCharacterInventoryUseCase handles updating character inventory entry
type UpdateCharacterInventoryUseCase struct {
	inventoryRepo repositories.CharacterInventoryRepository
	logger        logger.Logger
}

// NewUpdateCharacterInventoryUseCase creates a new UpdateCharacterInventoryUseCase
func NewUpdateCharacterInventoryUseCase(
	inventoryRepo repositories.CharacterInventoryRepository,
	logger logger.Logger,
) *UpdateCharacterInventoryUseCase {
	return &UpdateCharacterInventoryUseCase{
		inventoryRepo: inventoryRepo,
		logger:        logger,
	}
}

// UpdateCharacterInventoryInput represents the input for updating inventory
type UpdateCharacterInventoryInput struct {
	ID          uuid.UUID
	Quantity    *int
	SlotID      *uuid.UUID
	CustomName  *string
	CustomStats *json.RawMessage
}

// UpdateCharacterInventoryOutput represents the output of updating inventory
type UpdateCharacterInventoryOutput struct {
	Inventory *rpg.CharacterInventory
}

// Execute updates a character inventory entry
func (uc *UpdateCharacterInventoryUseCase) Execute(ctx context.Context, input UpdateCharacterInventoryInput) (*UpdateCharacterInventoryOutput, error) {
	// Get existing inventory entry
	inventory, err := uc.inventoryRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Quantity != nil {
		if err := inventory.SetQuantity(*input.Quantity); err != nil {
			return nil, err
		}
	}
	if input.SlotID != nil {
		inventory.SetSlot(input.SlotID)
	}
	if input.CustomName != nil {
		inventory.SetCustomName(input.CustomName)
	}
	if input.CustomStats != nil {
		if err := inventory.SetCustomStats(input.CustomStats); err != nil {
			return nil, err
		}
	}

	if err := inventory.Validate(); err != nil {
		return nil, err
	}

	if err := uc.inventoryRepo.Update(ctx, inventory); err != nil {
		uc.logger.Error("failed to update character inventory", "error", err, "id", input.ID)
		return nil, err
	}

	uc.logger.Info("character inventory updated", "id", input.ID)

	return &UpdateCharacterInventoryOutput{
		Inventory: inventory,
	}, nil
}

