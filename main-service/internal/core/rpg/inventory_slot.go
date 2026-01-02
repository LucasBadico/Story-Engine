package rpg

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSlotNameRequired = errors.New("slot name is required")
	ErrSlotSystemRequired = errors.New("RPG system ID is required")
)

// SlotType represents the type of inventory slot
type SlotType string

const (
	SlotTypeEquipment  SlotType = "equipment"
	SlotTypeConsumable SlotType = "consumable"
	SlotTypeQuest      SlotType = "quest"
)

// InventorySlot represents an inventory slot definition
type InventorySlot struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	RPGSystemID uuid.UUID  `json:"rpg_system_id"`
	Name        string     `json:"name"`
	SlotType    *SlotType  `json:"slot_type,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// NewInventorySlot creates a new inventory slot
func NewInventorySlot(tenantID, rpgSystemID uuid.UUID, name string) (*InventorySlot, error) {
	if name == "" {
		return nil, ErrSlotNameRequired
	}
	if rpgSystemID == uuid.Nil {
		return nil, ErrSlotSystemRequired
	}

	return &InventorySlot{
		ID:          uuid.New(),
		TenantID:    tenantID,
		RPGSystemID: rpgSystemID,
		Name:        name,
		CreatedAt:   time.Now(),
	}, nil
}

// Validate validates the inventory slot entity
func (s *InventorySlot) Validate() error {
	if s.Name == "" {
		return ErrSlotNameRequired
	}
	if s.RPGSystemID == uuid.Nil {
		return ErrSlotSystemRequired
	}
	return nil
}

// UpdateName updates the slot name
func (s *InventorySlot) UpdateName(name string) error {
	if name == "" {
		return ErrSlotNameRequired
	}
	s.Name = name
	return nil
}

// UpdateSlotType updates the slot type
func (s *InventorySlot) UpdateSlotType(slotType *SlotType) {
	s.SlotType = slotType
}


