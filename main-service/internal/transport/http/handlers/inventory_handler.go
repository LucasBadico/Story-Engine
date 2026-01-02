package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	inventoryslotapp "github.com/story-engine/main-service/internal/application/rpg/inventory_slot"
	inventoryitemapp "github.com/story-engine/main-service/internal/application/rpg/inventory_item"
	characterinventoryapp "github.com/story-engine/main-service/internal/application/rpg/character_inventory"
	"github.com/story-engine/main-service/internal/core/rpg"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
)

// InventoryHandler handles HTTP requests for inventory operations
type InventoryHandler struct {
	createSlotUseCase          *inventoryslotapp.CreateInventorySlotUseCase
	listSlotsUseCase           *inventoryslotapp.ListInventorySlotsUseCase
	createItemUseCase          *inventoryitemapp.CreateInventoryItemUseCase
	getItemUseCase             *inventoryitemapp.GetInventoryItemUseCase
	listItemsUseCase           *inventoryitemapp.ListInventoryItemsUseCase
	updateItemUseCase          *inventoryitemapp.UpdateInventoryItemUseCase
	deleteItemUseCase          *inventoryitemapp.DeleteInventoryItemUseCase
	addItemUseCase             *characterinventoryapp.AddItemToInventoryUseCase
	listInventoryUseCase       *characterinventoryapp.ListCharacterInventoryUseCase
	updateInventoryUseCase     *characterinventoryapp.UpdateCharacterInventoryUseCase
	equipItemUseCase           *characterinventoryapp.EquipItemUseCase
	unequipItemUseCase         *characterinventoryapp.UnequipItemUseCase
	deleteInventoryUseCase     *characterinventoryapp.DeleteCharacterInventoryUseCase
	transferItemUseCase        *characterinventoryapp.TransferItemUseCase
	logger                     logger.Logger
}

// NewInventoryHandler creates a new InventoryHandler
func NewInventoryHandler(
	createSlotUseCase *inventoryslotapp.CreateInventorySlotUseCase,
	listSlotsUseCase *inventoryslotapp.ListInventorySlotsUseCase,
	createItemUseCase *inventoryitemapp.CreateInventoryItemUseCase,
	getItemUseCase *inventoryitemapp.GetInventoryItemUseCase,
	listItemsUseCase *inventoryitemapp.ListInventoryItemsUseCase,
	updateItemUseCase *inventoryitemapp.UpdateInventoryItemUseCase,
	deleteItemUseCase *inventoryitemapp.DeleteInventoryItemUseCase,
	addItemUseCase *characterinventoryapp.AddItemToInventoryUseCase,
	listInventoryUseCase *characterinventoryapp.ListCharacterInventoryUseCase,
	updateInventoryUseCase *characterinventoryapp.UpdateCharacterInventoryUseCase,
	equipItemUseCase *characterinventoryapp.EquipItemUseCase,
	unequipItemUseCase *characterinventoryapp.UnequipItemUseCase,
	deleteInventoryUseCase *characterinventoryapp.DeleteCharacterInventoryUseCase,
	transferItemUseCase *characterinventoryapp.TransferItemUseCase,
	logger logger.Logger,
) *InventoryHandler {
	return &InventoryHandler{
		createSlotUseCase:      createSlotUseCase,
		listSlotsUseCase:       listSlotsUseCase,
		createItemUseCase:      createItemUseCase,
		getItemUseCase:         getItemUseCase,
		listItemsUseCase:       listItemsUseCase,
		updateItemUseCase:      updateItemUseCase,
		deleteItemUseCase:      deleteItemUseCase,
		addItemUseCase:         addItemUseCase,
		listInventoryUseCase:   listInventoryUseCase,
		updateInventoryUseCase: updateInventoryUseCase,
		equipItemUseCase:       equipItemUseCase,
		unequipItemUseCase:     unequipItemUseCase,
		deleteInventoryUseCase: deleteInventoryUseCase,
		transferItemUseCase:    transferItemUseCase,
		logger:                 logger,
	}
}

// CreateSlot handles POST /api/v1/rpg-systems/{id}/inventory-slots
func (h *InventoryHandler) CreateSlot(w http.ResponseWriter, r *http.Request) {
	rpgSystemIDStr := r.PathValue("id")
	rpgSystemID, err := uuid.Parse(rpgSystemIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Name     string         `json:"name"`
		SlotType *rpg.SlotType  `json:"slot_type,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.createSlotUseCase.Execute(r.Context(), inventoryslotapp.CreateInventorySlotInput{
		RPGSystemID: rpgSystemID,
		Name:        req.Name,
		SlotType:    req.SlotType,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"slot": output.Slot,
	})
}

// ListSlots handles GET /api/v1/rpg-systems/{id}/inventory-slots
func (h *InventoryHandler) ListSlots(w http.ResponseWriter, r *http.Request) {
	rpgSystemIDStr := r.PathValue("id")
	rpgSystemID, err := uuid.Parse(rpgSystemIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.listSlotsUseCase.Execute(r.Context(), inventoryslotapp.ListInventorySlotsInput{
		RPGSystemID: rpgSystemID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"slots": output.Slots,
		"total": len(output.Slots),
	})
}

// CreateItem handles POST /api/v1/rpg-systems/{id}/inventory-items
func (h *InventoryHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	rpgSystemIDStr := r.PathValue("id")
	rpgSystemID, err := uuid.Parse(rpgSystemIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		ArtifactID    *string          `json:"artifact_id,omitempty"`
		Name          string           `json:"name"`
		Category      *rpg.ItemCategory `json:"category,omitempty"`
		Description   *string          `json:"description,omitempty"`
		SlotsRequired *int             `json:"slots_required,omitempty"`
		Weight        *float64         `json:"weight,omitempty"`
		Size          *rpg.ItemSize    `json:"size,omitempty"`
		MaxStack      *int             `json:"max_stack,omitempty"`
		EquipSlots    *json.RawMessage `json:"equip_slots,omitempty"`
		Requirements  *json.RawMessage `json:"requirements,omitempty"`
		ItemStats     *json.RawMessage `json:"item_stats,omitempty"`
		IsTemplate    *bool            `json:"is_template,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var artifactID *uuid.UUID
	if req.ArtifactID != nil && *req.ArtifactID != "" {
		parsedID, err := uuid.Parse(*req.ArtifactID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "artifact_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		artifactID = &parsedID
	}

	output, err := h.createItemUseCase.Execute(r.Context(), inventoryitemapp.CreateInventoryItemInput{
		RPGSystemID:   rpgSystemID,
		ArtifactID:    artifactID,
		Name:          req.Name,
		Category:      req.Category,
		Description:   req.Description,
		SlotsRequired: req.SlotsRequired,
		Weight:        req.Weight,
		Size:          req.Size,
		MaxStack:      req.MaxStack,
		EquipSlots:    req.EquipSlots,
		Requirements:  req.Requirements,
		ItemStats:     req.ItemStats,
		IsTemplate:    req.IsTemplate,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"item": output.Item,
	})
}

// GetItem handles GET /api/v1/inventory-items/{id}
func (h *InventoryHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	itemID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getItemUseCase.Execute(r.Context(), inventoryitemapp.GetInventoryItemInput{
		ID: itemID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"item": output.Item,
	})
}

// ListItems handles GET /api/v1/rpg-systems/{id}/inventory-items
func (h *InventoryHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	rpgSystemIDStr := r.PathValue("id")
	rpgSystemID, err := uuid.Parse(rpgSystemIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	artifactIDStr := r.URL.Query().Get("artifact_id")
	var artifactID *uuid.UUID
	if artifactIDStr != "" {
		parsedID, err := uuid.Parse(artifactIDStr)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "artifact_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		artifactID = &parsedID
	}

	output, err := h.listItemsUseCase.Execute(r.Context(), inventoryitemapp.ListInventoryItemsInput{
		RPGSystemID: rpgSystemID,
		ArtifactID:  artifactID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": output.Items,
		"total": len(output.Items),
	})
}

// UpdateItem handles PUT /api/v1/inventory-items/{id}
func (h *InventoryHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	itemID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		ArtifactID    *string          `json:"artifact_id,omitempty"`
		Name          *string          `json:"name,omitempty"`
		Category      *rpg.ItemCategory `json:"category,omitempty"`
		Description   *string          `json:"description,omitempty"`
		SlotsRequired *int             `json:"slots_required,omitempty"`
		Weight        *float64         `json:"weight,omitempty"`
		Size          *rpg.ItemSize    `json:"size,omitempty"`
		MaxStack      *int             `json:"max_stack,omitempty"`
		EquipSlots    *json.RawMessage `json:"equip_slots,omitempty"`
		Requirements  *json.RawMessage `json:"requirements,omitempty"`
		ItemStats     *json.RawMessage `json:"item_stats,omitempty"`
		IsTemplate    *bool            `json:"is_template,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var artifactID *uuid.UUID
	if req.ArtifactID != nil && *req.ArtifactID != "" {
		parsedID, err := uuid.Parse(*req.ArtifactID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "artifact_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		artifactID = &parsedID
	}

	output, err := h.updateItemUseCase.Execute(r.Context(), inventoryitemapp.UpdateInventoryItemInput{
		ID:            itemID,
		ArtifactID:    artifactID,
		Name:          req.Name,
		Category:      req.Category,
		Description:   req.Description,
		SlotsRequired: req.SlotsRequired,
		Weight:        req.Weight,
		Size:          req.Size,
		MaxStack:      req.MaxStack,
		EquipSlots:    req.EquipSlots,
		Requirements:  req.Requirements,
		ItemStats:     req.ItemStats,
		IsTemplate:    req.IsTemplate,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"item": output.Item,
	})
}

// DeleteItem handles DELETE /api/v1/inventory-items/{id}
func (h *InventoryHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	itemID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.deleteItemUseCase.Execute(r.Context(), inventoryitemapp.DeleteInventoryItemInput{
		ID: itemID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddItem handles POST /api/v1/characters/{id}/inventory
func (h *InventoryHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	characterIDStr := r.PathValue("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		ItemID   uuid.UUID  `json:"item_id"`
		Quantity *int       `json:"quantity,omitempty"`
		SlotID   *uuid.UUID `json:"slot_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.addItemUseCase.Execute(r.Context(), characterinventoryapp.AddItemToInventoryInput{
		CharacterID: characterID,
		ItemID:      req.ItemID,
		Quantity:    req.Quantity,
		SlotID:      req.SlotID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"inventory": output.Inventory,
	})
}

// ListInventory handles GET /api/v1/characters/{id}/inventory
func (h *InventoryHandler) ListInventory(w http.ResponseWriter, r *http.Request) {
	characterIDStr := r.PathValue("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	equippedOnly := r.URL.Query().Get("equipped_only") == "true"

	output, err := h.listInventoryUseCase.Execute(r.Context(), characterinventoryapp.ListCharacterInventoryInput{
		CharacterID: characterID,
		EquippedOnly: equippedOnly,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": output.Items,
		"total":  len(output.Items),
	})
}

// UpdateInventory handles PUT /api/v1/character-inventory/{id}
func (h *InventoryHandler) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	inventoryID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Quantity    *int             `json:"quantity,omitempty"`
		SlotID      *uuid.UUID       `json:"slot_id,omitempty"`
		CustomName  *string          `json:"custom_name,omitempty"`
		CustomStats *json.RawMessage `json:"custom_stats,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.updateInventoryUseCase.Execute(r.Context(), characterinventoryapp.UpdateCharacterInventoryInput{
		ID:          inventoryID,
		Quantity:    req.Quantity,
		SlotID:      req.SlotID,
		CustomName:  req.CustomName,
		CustomStats: req.CustomStats,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"inventory": output.Inventory,
	})
}

// EquipItem handles PUT /api/v1/character-inventory/{id}/equip
func (h *InventoryHandler) EquipItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	inventoryID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.equipItemUseCase.Execute(r.Context(), characterinventoryapp.EquipItemInput{
		ID: inventoryID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"inventory": output.Inventory,
	})
}

// UnequipItem handles PUT /api/v1/character-inventory/{id}/unequip
func (h *InventoryHandler) UnequipItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	inventoryID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.unequipItemUseCase.Execute(r.Context(), characterinventoryapp.UnequipItemInput{
		ID: inventoryID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"inventory": output.Inventory,
	})
}

// DeleteInventory handles DELETE /api/v1/character-inventory/{id}
func (h *InventoryHandler) DeleteInventory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	inventoryID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.deleteInventoryUseCase.Execute(r.Context(), characterinventoryapp.DeleteCharacterInventoryInput{
		ID: inventoryID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TransferItem handles POST /api/v1/character-inventory/{id}/transfer
func (h *InventoryHandler) TransferItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	inventoryID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		ToCharacterID uuid.UUID `json:"to_character_id"`
		Quantity      *int      `json:"quantity,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.transferItemUseCase.Execute(r.Context(), characterinventoryapp.TransferItemInput{
		InventoryID:   inventoryID,
		ToCharacterID: req.ToCharacterID,
		Quantity:      req.Quantity,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"from_inventory": output.FromInventory,
		"to_inventory":   output.ToInventory,
	})
}


