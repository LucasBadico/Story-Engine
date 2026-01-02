package inventory_item

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateInventoryItemUseCase handles inventory item creation
type CreateInventoryItemUseCase struct {
	itemRepo    repositories.InventoryItemRepository
	rpgSystemRepo repositories.RPGSystemRepository
	artifactRepo repositories.ArtifactRepository
	logger      logger.Logger
}

// NewCreateInventoryItemUseCase creates a new CreateInventoryItemUseCase
func NewCreateInventoryItemUseCase(
	itemRepo repositories.InventoryItemRepository,
	rpgSystemRepo repositories.RPGSystemRepository,
	artifactRepo repositories.ArtifactRepository,
	logger logger.Logger,
) *CreateInventoryItemUseCase {
	return &CreateInventoryItemUseCase{
		itemRepo:    itemRepo,
		rpgSystemRepo: rpgSystemRepo,
		artifactRepo: artifactRepo,
		logger:      logger,
	}
}

// CreateInventoryItemInput represents the input for creating an inventory item
type CreateInventoryItemInput struct {
	RPGSystemID   uuid.UUID
	ArtifactID    *uuid.UUID
	Name          string
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

// CreateInventoryItemOutput represents the output of creating an inventory item
type CreateInventoryItemOutput struct {
	Item *rpg.InventoryItem
}

// Execute creates a new inventory item
func (uc *CreateInventoryItemUseCase) Execute(ctx context.Context, input CreateInventoryItemInput) (*CreateInventoryItemOutput, error) {
	// Validate RPG system exists
	_, err := uc.rpgSystemRepo.GetByID(ctx, input.RPGSystemID)
	if err != nil {
		return nil, err
	}

	// Validate artifact exists if provided
	if input.ArtifactID != nil {
		_, err := uc.artifactRepo.GetByID(ctx, *input.ArtifactID)
		if err != nil {
			return nil, err
		}
	}

	// Create item
	item, err := rpg.NewInventoryItem(input.RPGSystemID, input.Name)
	if err != nil {
		return nil, err
	}

	if input.ArtifactID != nil {
		item.SetArtifactID(input.ArtifactID)
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

	if err := uc.itemRepo.Create(ctx, item); err != nil {
		uc.logger.Error("failed to create inventory item", "error", err, "rpg_system_id", input.RPGSystemID)
		return nil, err
	}

	uc.logger.Info("inventory item created", "item_id", item.ID, "name", item.Name)

	return &CreateInventoryItemOutput{
		Item: item,
	}, nil
}

