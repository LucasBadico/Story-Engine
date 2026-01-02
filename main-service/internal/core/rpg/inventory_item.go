package rpg

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrItemNameRequired = errors.New("item name is required")
	ErrItemSystemRequired = errors.New("RPG system ID is required")
	ErrInvalidItemJSON = errors.New("invalid item JSON")
	ErrInvalidSlotsRequired = errors.New("slots_required must be at least 1")
	ErrInvalidMaxStack = errors.New("max_stack must be at least 1")
)

// ItemCategory represents the category of an inventory item
type ItemCategory string

const (
	ItemCategoryWeapon     ItemCategory = "weapon"
	ItemCategoryArmor      ItemCategory = "armor"
	ItemCategoryConsumable ItemCategory = "consumable"
	ItemCategoryQuest      ItemCategory = "quest"
	ItemCategoryMisc      ItemCategory = "misc"
)

// ItemSize represents the size of an inventory item
type ItemSize string

const (
	ItemSizeTiny   ItemSize = "tiny"
	ItemSizeSmall  ItemSize = "small"
	ItemSizeMedium ItemSize = "medium"
	ItemSizeLarge  ItemSize = "large"
	ItemSizeHuge   ItemSize = "huge"
)

// InventoryItem represents an inventory item definition (mechanical)
type InventoryItem struct {
	ID            uuid.UUID       `json:"id"`
	TenantID      uuid.UUID       `json:"tenant_id"`
	RPGSystemID   uuid.UUID       `json:"rpg_system_id"`
	ArtifactID    *uuid.UUID      `json:"artifact_id,omitempty"` // optional: link to narrative
	Name          string          `json:"name"`
	Category      *ItemCategory   `json:"category,omitempty"`
	Description   *string         `json:"description,omitempty"`
	SlotsRequired int             `json:"slots_required"`
	Weight        *float64        `json:"weight,omitempty"`
	Size          *ItemSize       `json:"size,omitempty"`
	MaxStack      int             `json:"max_stack"`
	EquipSlots    *json.RawMessage `json:"equip_slots,omitempty"` // JSONB
	Requirements  *json.RawMessage `json:"requirements,omitempty"` // JSONB
	ItemStats     *json.RawMessage `json:"item_stats,omitempty"` // JSONB
	IsTemplate    bool            `json:"is_template"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// NewInventoryItem creates a new inventory item
func NewInventoryItem(tenantID, rpgSystemID uuid.UUID, name string) (*InventoryItem, error) {
	if name == "" {
		return nil, ErrItemNameRequired
	}
	if rpgSystemID == uuid.Nil {
		return nil, ErrItemSystemRequired
	}

	now := time.Now()
	return &InventoryItem{
		ID:            uuid.New(),
		TenantID:      tenantID,
		RPGSystemID:   rpgSystemID,
		Name:          name,
		SlotsRequired: 1,
		MaxStack:      1,
		IsTemplate:    false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// Validate validates the inventory item entity
func (i *InventoryItem) Validate() error {
	if i.Name == "" {
		return ErrItemNameRequired
	}
	if i.RPGSystemID == uuid.Nil {
		return ErrItemSystemRequired
	}
	if i.SlotsRequired < 1 {
		return ErrInvalidSlotsRequired
	}
	if i.MaxStack < 1 {
		return ErrInvalidMaxStack
	}

	// Validate JSON fields
	if i.EquipSlots != nil && len(*i.EquipSlots) > 0 {
		var slots []interface{}
		if err := json.Unmarshal(*i.EquipSlots, &slots); err != nil {
			return ErrInvalidItemJSON
		}
	}

	if i.Requirements != nil && len(*i.Requirements) > 0 {
		var req map[string]interface{}
		if err := json.Unmarshal(*i.Requirements, &req); err != nil {
			return ErrInvalidItemJSON
		}
	}

	if i.ItemStats != nil && len(*i.ItemStats) > 0 {
		var stats map[string]interface{}
		if err := json.Unmarshal(*i.ItemStats, &stats); err != nil {
			return ErrInvalidItemJSON
		}
	}

	return nil
}

// UpdateName updates the item name
func (i *InventoryItem) UpdateName(name string) error {
	if name == "" {
		return ErrItemNameRequired
	}
	i.Name = name
	i.UpdatedAt = time.Now()
	return nil
}

// UpdateCategory updates the category
func (i *InventoryItem) UpdateCategory(category *ItemCategory) {
	i.Category = category
	i.UpdatedAt = time.Now()
}

// UpdateDescription updates the description
func (i *InventoryItem) UpdateDescription(description *string) {
	i.Description = description
	i.UpdatedAt = time.Now()
}

// UpdateSlotsRequired updates slots required
func (i *InventoryItem) UpdateSlotsRequired(slots int) error {
	if slots < 1 {
		return ErrInvalidSlotsRequired
	}
	i.SlotsRequired = slots
	i.UpdatedAt = time.Now()
	return nil
}

// UpdateWeight updates the weight
func (i *InventoryItem) UpdateWeight(weight *float64) {
	i.Weight = weight
	i.UpdatedAt = time.Now()
}

// UpdateSize updates the size
func (i *InventoryItem) UpdateSize(size *ItemSize) {
	i.Size = size
	i.UpdatedAt = time.Now()
}

// UpdateMaxStack updates max stack
func (i *InventoryItem) UpdateMaxStack(maxStack int) error {
	if maxStack < 1 {
		return ErrInvalidMaxStack
	}
	i.MaxStack = maxStack
	i.UpdatedAt = time.Now()
	return nil
}

// UpdateEquipSlots updates equip slots
func (i *InventoryItem) UpdateEquipSlots(equipSlots *json.RawMessage) error {
	if equipSlots != nil && len(*equipSlots) > 0 {
		var slots []interface{}
		if err := json.Unmarshal(*equipSlots, &slots); err != nil {
			return ErrInvalidItemJSON
		}
	}
	i.EquipSlots = equipSlots
	i.UpdatedAt = time.Now()
	return nil
}

// UpdateRequirements updates requirements
func (i *InventoryItem) UpdateRequirements(requirements *json.RawMessage) error {
	if requirements != nil && len(*requirements) > 0 {
		var req map[string]interface{}
		if err := json.Unmarshal(*requirements, &req); err != nil {
			return ErrInvalidItemJSON
		}
	}
	i.Requirements = requirements
	i.UpdatedAt = time.Now()
	return nil
}

// UpdateItemStats updates item stats
func (i *InventoryItem) UpdateItemStats(itemStats *json.RawMessage) error {
	if itemStats != nil && len(*itemStats) > 0 {
		var stats map[string]interface{}
		if err := json.Unmarshal(*itemStats, &stats); err != nil {
			return ErrInvalidItemJSON
		}
	}
	i.ItemStats = itemStats
	i.UpdatedAt = time.Now()
	return nil
}

// SetArtifactID sets the artifact ID (link to narrative)
func (i *InventoryItem) SetArtifactID(artifactID *uuid.UUID) {
	i.ArtifactID = artifactID
	i.UpdatedAt = time.Now()
}

// SetTemplate sets the template flag
func (i *InventoryItem) SetTemplate(isTemplate bool) {
	i.IsTemplate = isTemplate
	i.UpdatedAt = time.Now()
}


