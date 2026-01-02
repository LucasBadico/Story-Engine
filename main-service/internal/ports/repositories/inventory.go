package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
)

// InventorySlotRepository defines the interface for inventory slot persistence
type InventorySlotRepository interface {
	Create(ctx context.Context, slot *rpg.InventorySlot) error
	GetByID(ctx context.Context, id uuid.UUID) (*rpg.InventorySlot, error)
	ListBySystem(ctx context.Context, rpgSystemID uuid.UUID) ([]*rpg.InventorySlot, error)
	Update(ctx context.Context, slot *rpg.InventorySlot) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// InventoryItemRepository defines the interface for inventory item persistence
type InventoryItemRepository interface {
	Create(ctx context.Context, item *rpg.InventoryItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*rpg.InventoryItem, error)
	ListBySystem(ctx context.Context, rpgSystemID uuid.UUID) ([]*rpg.InventoryItem, error)
	ListByArtifact(ctx context.Context, artifactID uuid.UUID) ([]*rpg.InventoryItem, error)
	Update(ctx context.Context, item *rpg.InventoryItem) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// CharacterInventoryRepository defines the interface for character inventory persistence
type CharacterInventoryRepository interface {
	Create(ctx context.Context, inventory *rpg.CharacterInventory) error
	GetByID(ctx context.Context, id uuid.UUID) (*rpg.CharacterInventory, error)
	GetByCharacterAndItem(ctx context.Context, characterID, itemID uuid.UUID, slotID *uuid.UUID) (*rpg.CharacterInventory, error)
	ListByCharacter(ctx context.Context, characterID uuid.UUID) ([]*rpg.CharacterInventory, error)
	ListEquippedByCharacter(ctx context.Context, characterID uuid.UUID) ([]*rpg.CharacterInventory, error)
	Update(ctx context.Context, inventory *rpg.CharacterInventory) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByCharacter(ctx context.Context, characterID uuid.UUID) error
}


