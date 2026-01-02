package handlers

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	inventoryapp "github.com/story-engine/main-service/internal/application/rpg/character_inventory"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	"github.com/story-engine/main-service/internal/platform/logger"
	inventorypb "github.com/story-engine/main-service/proto/inventory"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InventoryHandler implements the InventoryService gRPC service
type InventoryHandler struct {
	inventorypb.UnimplementedInventoryServiceServer
	addItemUseCase    *inventoryapp.AddItemToInventoryUseCase
	updateItemUseCase *inventoryapp.UpdateCharacterInventoryUseCase
	deleteItemUseCase *inventoryapp.DeleteCharacterInventoryUseCase
	listInventoryUseCase *inventoryapp.ListCharacterInventoryUseCase
	logger           logger.Logger
}

// NewInventoryHandler creates a new InventoryHandler
func NewInventoryHandler(
	addItemUseCase *inventoryapp.AddItemToInventoryUseCase,
	updateItemUseCase *inventoryapp.UpdateCharacterInventoryUseCase,
	deleteItemUseCase *inventoryapp.DeleteCharacterInventoryUseCase,
	listInventoryUseCase *inventoryapp.ListCharacterInventoryUseCase,
	logger logger.Logger,
) *InventoryHandler {
	return &InventoryHandler{
		addItemUseCase:     addItemUseCase,
		updateItemUseCase:  updateItemUseCase,
		deleteItemUseCase:  deleteItemUseCase,
		listInventoryUseCase: listInventoryUseCase,
		logger:             logger,
	}
}

// AddItemToInventory adds an item to a character's inventory
func (h *InventoryHandler) AddItemToInventory(ctx context.Context, req *inventorypb.AddItemToInventoryRequest) (*inventorypb.AddItemToInventoryResponse, error) {
	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	itemID, err := uuid.Parse(req.ItemId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid item_id: %v", err)
	}

	var quantity *int
	if req.Quantity != nil && *req.Quantity > 0 {
		q := int(*req.Quantity)
		quantity = &q
	}

	var slotID *uuid.UUID
	if req.SlotId != nil && *req.SlotId != "" {
		sid, err := uuid.Parse(*req.SlotId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid slot_id: %v", err)
		}
		slotID = &sid
	}

	input := inventoryapp.AddItemToInventoryInput{
		CharacterID: characterID,
		ItemID:      itemID,
		Quantity:    quantity,
		SlotID:      slotID,
	}

	output, err := h.addItemUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &inventorypb.AddItemToInventoryResponse{
		CharacterInventory: mappers.CharacterInventoryToProto(output.Inventory),
	}, nil
}

// UpdateInventoryItem updates an inventory item
func (h *InventoryHandler) UpdateInventoryItem(ctx context.Context, req *inventorypb.UpdateInventoryItemRequest) (*inventorypb.UpdateInventoryItemResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid inventory id: %v", err)
	}

	var quantity *int
	if req.Quantity != nil && *req.Quantity > 0 {
		q := int(*req.Quantity)
		quantity = &q
	}

	var slotID *uuid.UUID
	if req.SlotId != nil && *req.SlotId != "" {
		sid, err := uuid.Parse(*req.SlotId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid slot_id: %v", err)
		}
		slotID = &sid
	}

	var customName *string
	if req.CustomName != nil {
		customName = req.CustomName
	}

	var customStats *json.RawMessage
	if req.CustomStats != nil && *req.CustomStats != "" {
		stats := json.RawMessage(*req.CustomStats)
		customStats = &stats
	}

	var isEquipped *bool
	if req.IsEquipped != nil {
		isEquipped = req.IsEquipped
	}

	input := inventoryapp.UpdateCharacterInventoryInput{
		ID:          id,
		Quantity:    quantity,
		SlotID:      slotID,
		CustomName:  customName,
		CustomStats: customStats,
	}

	output, err := h.updateItemUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &inventorypb.UpdateInventoryItemResponse{
		CharacterInventory: mappers.CharacterInventoryToProto(output.Inventory),
	}, nil
}

// RemoveItemFromInventory removes an item from inventory
func (h *InventoryHandler) RemoveItemFromInventory(ctx context.Context, req *inventorypb.RemoveItemFromInventoryRequest) (*inventorypb.RemoveItemFromInventoryResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid inventory id: %v", err)
	}

	err = h.deleteItemUseCase.Execute(ctx, inventoryapp.DeleteCharacterInventoryInput{
		ID: id,
	})
	if err != nil {
		return nil, err
	}

	return &inventorypb.RemoveItemFromInventoryResponse{}, nil
}

// ListInventory lists items in a character's inventory
func (h *InventoryHandler) ListInventory(ctx context.Context, req *inventorypb.ListInventoryRequest) (*inventorypb.ListInventoryResponse, error) {
	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	output, err := h.listInventoryUseCase.Execute(ctx, inventoryapp.ListCharacterInventoryInput{
		CharacterID: characterID,
		EquippedOnly: false,
	})
	if err != nil {
		return nil, err
	}

	inventory := make([]*inventorypb.CharacterInventory, len(output.Items))
	for i, inv := range output.Items {
		inventory[i] = mappers.CharacterInventoryToProto(inv)
	}

	return &inventorypb.ListInventoryResponse{
		CharacterInventory: inventory,
		TotalCount:         int32(len(output.Items)),
	}, nil
}

