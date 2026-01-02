package mappers

import (
	"github.com/story-engine/main-service/internal/core/rpg"
	inventorypb "github.com/story-engine/main-service/proto/inventory"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// InventoryItemToProto converts an inventory item domain entity to a protobuf message
func InventoryItemToProto(i *rpg.InventoryItem) *inventorypb.InventoryItem {
	if i == nil {
		return nil
	}

	var artifactID *string
	if i.ArtifactID != nil {
		id := i.ArtifactID.String()
		artifactID = &id
	}

	var category *string
	if i.Category != nil {
		cat := string(*i.Category)
		category = &cat
	}

	var description *string
	if i.Description != nil {
		desc := *i.Description
		description = &desc
	}

	var weight *float64
	if i.Weight != nil {
		w := *i.Weight
		weight = &w
	}

	var size *string
	if i.Size != nil {
		s := string(*i.Size)
		size = &s
	}

	var equipSlots *string
	if i.EquipSlots != nil {
		slots := string(*i.EquipSlots)
		equipSlots = &slots
	}

	var requirements *string
	if i.Requirements != nil {
		req := string(*i.Requirements)
		requirements = &req
	}

	var itemStats *string
	if i.ItemStats != nil {
		stats := string(*i.ItemStats)
		itemStats = &stats
	}

	return &inventorypb.InventoryItem{
		Id:            i.ID.String(),
		RpgSystemId:   i.RPGSystemID.String(),
		ArtifactId:    artifactID,
		Name:          i.Name,
		Category:      category,
		Description:   description,
		SlotsRequired: int32(i.SlotsRequired),
		Weight:        weight,
		Size:          size,
		MaxStack:      int32(i.MaxStack),
		EquipSlots:    equipSlots,
		Requirements:  requirements,
		ItemStats:     itemStats,
		IsTemplate:    i.IsTemplate,
		CreatedAt:     timestamppb.New(i.CreatedAt),
		UpdatedAt:     timestamppb.New(i.UpdatedAt),
	}
}

// InventorySlotToProto converts an inventory slot domain entity to a protobuf message
func InventorySlotToProto(s *rpg.InventorySlot) *inventorypb.InventorySlot {
	if s == nil {
		return nil
	}

	var slotType *string
	if s.SlotType != nil {
		st := string(*s.SlotType)
		slotType = &st
	}

	return &inventorypb.InventorySlot{
		Id:          s.ID.String(),
		RpgSystemId: s.RPGSystemID.String(),
		Name:        s.Name,
		SlotType:    slotType,
		CreatedAt:   timestamppb.New(s.CreatedAt),
	}
}

// CharacterInventoryToProto converts a character inventory domain entity to a protobuf message
func CharacterInventoryToProto(ci *rpg.CharacterInventory) *inventorypb.CharacterInventory {
	if ci == nil {
		return nil
	}

	var slotID *string
	if ci.SlotID != nil {
		id := ci.SlotID.String()
		slotID = &id
	}

	var customName *string
	if ci.CustomName != nil {
		name := *ci.CustomName
		customName = &name
	}

	var customStats *string
	if ci.CustomStats != nil {
		stats := string(*ci.CustomStats)
		customStats = &stats
	}

	return &inventorypb.CharacterInventory{
		Id:           ci.ID.String(),
		CharacterId:  ci.CharacterID.String(),
		ItemId:       ci.ItemID.String(),
		Quantity:     int32(ci.Quantity),
		SlotId:       slotID,
		IsEquipped:   ci.IsEquipped,
		CustomName:   customName,
		CustomStats:  customStats,
		AcquiredAt:   timestamppb.New(ci.AcquiredAt),
	}
}

