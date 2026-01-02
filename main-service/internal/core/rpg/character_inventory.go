package rpg

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCharacterInventoryRequired = errors.New("character and item IDs are required")
	ErrInvalidQuantity = errors.New("quantity must be at least 1")
	ErrInvalidCustomStatsJSON = errors.New("invalid custom_stats JSON")
)

// CharacterInventory represents an item in a character's inventory
type CharacterInventory struct {
	ID          uuid.UUID       `json:"id"`
	CharacterID uuid.UUID       `json:"character_id"`
	ItemID      uuid.UUID       `json:"item_id"`
	Quantity    int             `json:"quantity"`
	SlotID      *uuid.UUID      `json:"slot_id,omitempty"`
	IsEquipped  bool            `json:"is_equipped"`
	CustomName  *string         `json:"custom_name,omitempty"`
	CustomStats *json.RawMessage `json:"custom_stats,omitempty"` // JSONB
	AcquiredAt  time.Time       `json:"acquired_at"`
}

// NewCharacterInventory creates a new character inventory entry
func NewCharacterInventory(characterID, itemID uuid.UUID) (*CharacterInventory, error) {
	if characterID == uuid.Nil {
		return nil, ErrCharacterInventoryRequired
	}
	if itemID == uuid.Nil {
		return nil, ErrCharacterInventoryRequired
	}

	return &CharacterInventory{
		ID:          uuid.New(),
		CharacterID: characterID,
		ItemID:      itemID,
		Quantity:    1,
		IsEquipped:  false,
		AcquiredAt:  time.Now(),
	}, nil
}

// Validate validates the character inventory entity
func (ci *CharacterInventory) Validate() error {
	if ci.CharacterID == uuid.Nil {
		return ErrCharacterInventoryRequired
	}
	if ci.ItemID == uuid.Nil {
		return ErrCharacterInventoryRequired
	}
	if ci.Quantity < 1 {
		return ErrInvalidQuantity
	}

		// Validate custom_stats JSON if provided
	if ci.CustomStats != nil && len(*ci.CustomStats) > 0 {
		var stats map[string]interface{}
		if err := json.Unmarshal(*ci.CustomStats, &stats); err != nil {
			return ErrInvalidCustomStatsJSON
		}
	}

	return nil
}

// SetQuantity sets the quantity
func (ci *CharacterInventory) SetQuantity(quantity int) error {
	if quantity < 1 {
		return ErrInvalidQuantity
	}
	ci.Quantity = quantity
	return nil
}

// AddQuantity adds to the quantity
func (ci *CharacterInventory) AddQuantity(amount int) error {
	if ci.Quantity+amount < 1 {
		return ErrInvalidQuantity
	}
	ci.Quantity += amount
	return nil
}

// SetSlot sets the slot ID
func (ci *CharacterInventory) SetSlot(slotID *uuid.UUID) {
	ci.SlotID = slotID
}

// SetEquipped sets whether the item is equipped
func (ci *CharacterInventory) SetEquipped(isEquipped bool) {
	ci.IsEquipped = isEquipped
}

// SetCustomName sets the custom name
func (ci *CharacterInventory) SetCustomName(customName *string) {
	ci.CustomName = customName
}

// SetCustomStats sets the custom stats
func (ci *CharacterInventory) SetCustomStats(customStats *json.RawMessage) error {
	if customStats != nil && len(*customStats) > 0 {
		var stats map[string]interface{}
		if err := json.Unmarshal(*customStats, &stats); err != nil {
			return ErrInvalidCustomStatsJSON
		}
	}
	ci.CustomStats = customStats
	return nil
}

