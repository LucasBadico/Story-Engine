package inventory_item

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateInventoryItemUseCase handles inventory item updates
type UpdateInventoryItemUseCase struct {
	itemRepo repositories.InventoryItemRepository
	logger   logger.Logger
}

// NewUpdateInventoryItemUseCase creates a new UpdateInventoryItemUseCase
func NewUpdateInventoryItemUseCase(
	itemRepo repositories.InventoryItemRepository,
	logger logger.Logger,
) *UpdateInventoryItemUseCase {
	return &UpdateInventoryItemUseCase{
		itemRepo: itemRepo,
		logger:   logger,
	}
}

// UpdateInventoryItemInput represents the input for updating an inventory item
type UpdateInventoryItemInput struct {
	ID            uuid.UUID
	ArtifactID    *uuid.UUID
	Name          *string
	Category      *rpg.ItemCategory
	Description   *string
	SlotsRequired *int
	Weight        *float64
	Size          *rpg.ItemSize
	MaxStack      *int
	EquipSlots    *json.RawMessage
	Requirements  *json.RawMessage
	ItemStats     *json.RawMessage
	IsTemplate    *bool
}

// UpdateInventoryItemOutput represents the output of updating an inventory item
type UpdateInventoryItemOutput struct {
	Item *rpg.InventoryItem
}

// Execute updates an inventory item
func (uc *UpdateInventoryItemUseCase) Execute(ctx context.Context, input UpdateInventoryItemInput) (*UpdateInventoryItemOutput, error) {
	// Get existing item
	item, err := uc.itemRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.ArtifactID != nil {
		item.SetArtifactID(input.ArtifactID)
	}
	if input.Name != nil {
		if err := item.UpdateName(*input.Name); err != nil {
			return nil, err
		}
	}
	if input.Category != nil {
		item.UpdateCategory(input.Category)
	}
	if input.Description != nil {
		item.UpdateDescription(input.Description)
	}
	if input.SlotsRequired != nil {
		if err := item.UpdateSlotsRequired(*input.SlotsRequired); err != nil {
			return nil, err
		}
	}
	if input.Weight != nil {
		item.UpdateWeight(input.Weight)
	}
	if input.Size != nil {
		item.UpdateSize(input.Size)
	}
	if input.MaxStack != nil {
		if err := item.UpdateMaxStack(*input.MaxStack); err != nil {
			return nil, err
		}
	}
	if input.EquipSlots != nil {
		if err := item.UpdateEquipSlots(input.EquipSlots); err != nil {
			return nil, err
		}
	}
	if input.Requirements != nil {
		if err := item.UpdateRequirements(input.Requirements); err != nil {
			return nil, err
		}
	}
	if input.ItemStats != nil {
		if err := item.UpdateItemStats(input.ItemStats); err != nil {
			return nil, err
		}
	}
	if input.IsTemplate != nil {
		item.SetTemplate(*input.IsTemplate)
	}

	if err := item.Validate(); err != nil {
		return nil, err
	}

	if err := uc.itemRepo.Update(ctx, item); err != nil {
		uc.logger.Error("failed to update inventory item", "error", err, "item_id", input.ID)
		return nil, err
	}

	uc.logger.Info("inventory item updated", "item_id", input.ID)

	return &UpdateInventoryItemOutput{
		Item: item,
	}, nil
}

