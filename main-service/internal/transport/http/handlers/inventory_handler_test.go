//go:build integration

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	characterinventoryapp "github.com/story-engine/main-service/internal/application/rpg/character_inventory"
	inventoryitemapp "github.com/story-engine/main-service/internal/application/rpg/inventory_item"
	inventoryslotapp "github.com/story-engine/main-service/internal/application/rpg/inventory_slot"
	rpgsystemapp "github.com/story-engine/main-service/internal/application/rpg/rpg_system"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	rpgcharacterapp "github.com/story-engine/main-service/internal/application/rpg/character"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestInventoryHandler_AddItem(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)
	characterRepo := postgres.NewCharacterRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	traitRepo := postgres.NewTraitRepository(db)
	characterTraitRepo := postgres.NewCharacterTraitRepository(db)
	rpgClassRepo := postgres.NewRPGClassRepository(db)
	rpgSystemRepo := postgres.NewRPGSystemRepository(db)
	createCharacterUseCase := characterapp.NewCreateCharacterUseCase(characterRepo, worldRepo, archetypeRepo, auditLogRepo, log)
	getCharacterUseCase := characterapp.NewGetCharacterUseCase(characterRepo, log)
	listCharactersUseCase := characterapp.NewListCharactersUseCase(characterRepo, log)
	updateCharacterUseCase := characterapp.NewUpdateCharacterUseCase(characterRepo, archetypeRepo, worldRepo, auditLogRepo, log)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo, worldRepo, auditLogRepo, log)
	addTraitUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, traitRepo, characterTraitRepo, log)
	removeTraitUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log)
	updateTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, traitRepo, log)
	getTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log)
	changeClassUseCase := rpgcharacterapp.NewChangeCharacterClassUseCase(characterRepo, rpgClassRepo, log)
	getAvailableClassesUseCase := rpgcharacterapp.NewGetAvailableClassesUseCase(characterRepo, rpgClassRepo, rpgSystemRepo, log)
	characterHandler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

	// Create character
	characterBody := `{"name": "Test Character"}`
	characterReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/characters", strings.NewReader(characterBody))
	characterReq.Header.Set("Content-Type", "application/json")
	characterReq.Header.Set("X-Tenant-ID", tenantID)
	characterReq.SetPathValue("world_id", worldID)
	characterW := httptest.NewRecorder()
	withTenantMiddleware(characterHandler.Create).ServeHTTP(characterW, characterReq)

	if characterW.Code != http.StatusCreated {
		t.Fatalf("failed to create character: status %d, body: %s", characterW.Code, characterW.Body.String())
	}

	var characterResp map[string]interface{}
	if err := json.NewDecoder(characterW.Body).Decode(&characterResp); err != nil {
		t.Fatalf("failed to decode character response: %v", err)
	}

	characterObj, ok := characterResp["character"].(map[string]interface{})
	if !ok {
		t.Fatalf("character response missing character object: %v", characterResp)
	}

	characterID, ok := characterObj["id"].(string)
	if !ok {
		t.Fatalf("character response missing id: %v", characterObj)
	}

	// Create RPG system and inventory item
	tenantRepo := postgres.NewTenantRepository(db)
	createRPGSystemUseCase := rpgsystemapp.NewCreateRPGSystemUseCase(rpgSystemRepo, tenantRepo, log)
	getRPGSystemUseCase := rpgsystemapp.NewGetRPGSystemUseCase(rpgSystemRepo, log)
	listRPGSystemsUseCase := rpgsystemapp.NewListRPGSystemsUseCase(rpgSystemRepo, log)
	updateRPGSystemUseCase := rpgsystemapp.NewUpdateRPGSystemUseCase(rpgSystemRepo, log)
	deleteRPGSystemUseCase := rpgsystemapp.NewDeleteRPGSystemUseCase(rpgSystemRepo, log)
	rpgSystemHandler := NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)

	rpgSystemBody := `{"name": "Test RPG System", "base_stats_schema": {}}`
	rpgSystemReq := httptest.NewRequest("POST", "/api/v1/rpg-systems", strings.NewReader(rpgSystemBody))
	rpgSystemReq.Header.Set("Content-Type", "application/json")
	rpgSystemReq.Header.Set("X-Tenant-ID", tenantID)
	rpgSystemW := httptest.NewRecorder()
	withTenantMiddleware(rpgSystemHandler.Create).ServeHTTP(rpgSystemW, rpgSystemReq)

	if rpgSystemW.Code != http.StatusCreated {
		t.Fatalf("failed to create RPG system: status %d, body: %s", rpgSystemW.Code, rpgSystemW.Body.String())
	}

	var rpgSystemResp map[string]interface{}
	if err := json.NewDecoder(rpgSystemW.Body).Decode(&rpgSystemResp); err != nil {
		t.Fatalf("failed to decode RPG system response: %v", err)
	}

	rpgSystemObj, ok := rpgSystemResp["rpg_system"].(map[string]interface{})
	if !ok {
		t.Fatalf("RPG system response missing rpg_system object: %v", rpgSystemResp)
	}

	rpgSystemID, ok := rpgSystemObj["id"].(string)
	if !ok {
		t.Fatalf("RPG system response missing id: %v", rpgSystemObj)
	}

	inventorySlotRepo := postgres.NewInventorySlotRepository(db)
	inventoryItemRepo := postgres.NewInventoryItemRepository(db)
	characterInventoryRepo := postgres.NewCharacterInventoryRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	createSlotUseCase := inventoryslotapp.NewCreateInventorySlotUseCase(inventorySlotRepo, rpgSystemRepo, log)
	listSlotsUseCase := inventoryslotapp.NewListInventorySlotsUseCase(inventorySlotRepo, log)
	createItemUseCase := inventoryitemapp.NewCreateInventoryItemUseCase(inventoryItemRepo, rpgSystemRepo, artifactRepo, log)
	getItemUseCase := inventoryitemapp.NewGetInventoryItemUseCase(inventoryItemRepo, log)
	listItemsUseCase := inventoryitemapp.NewListInventoryItemsUseCase(inventoryItemRepo, log)
	updateItemUseCase := inventoryitemapp.NewUpdateInventoryItemUseCase(inventoryItemRepo, log)
	deleteItemUseCase := inventoryitemapp.NewDeleteInventoryItemUseCase(inventoryItemRepo, log)
	addItemUseCase := characterinventoryapp.NewAddItemToInventoryUseCase(characterInventoryRepo, characterRepo, inventoryItemRepo, log)
	listInventoryUseCase := characterinventoryapp.NewListCharacterInventoryUseCase(characterInventoryRepo, log)
	updateInventoryUseCase := characterinventoryapp.NewUpdateCharacterInventoryUseCase(characterInventoryRepo, log)
	equipItemUseCase := characterinventoryapp.NewEquipItemUseCase(characterInventoryRepo, log)
	unequipItemUseCase := characterinventoryapp.NewUnequipItemUseCase(characterInventoryRepo, log)
	deleteInventoryUseCase := characterinventoryapp.NewDeleteCharacterInventoryUseCase(characterInventoryRepo, log)
	transferItemUseCase := characterinventoryapp.NewTransferItemUseCase(characterInventoryRepo, characterRepo, log)
	handler := NewInventoryHandler(createSlotUseCase, listSlotsUseCase, createItemUseCase, getItemUseCase, listItemsUseCase, updateItemUseCase, deleteItemUseCase, addItemUseCase, listInventoryUseCase, updateInventoryUseCase, equipItemUseCase, unequipItemUseCase, deleteInventoryUseCase, transferItemUseCase, log)

	// Create inventory item
	itemBody := `{"name": "Test Item", "category": "weapon"}`
	itemReq := httptest.NewRequest("POST", "/api/v1/rpg-systems/"+rpgSystemID+"/inventory-items", strings.NewReader(itemBody))
	itemReq.Header.Set("Content-Type", "application/json")
	itemReq.Header.Set("X-Tenant-ID", tenantID)
	itemReq.SetPathValue("id", rpgSystemID)
	itemW := httptest.NewRecorder()
	withTenantMiddleware(handler.CreateItem).ServeHTTP(itemW, itemReq)

	if itemW.Code != http.StatusCreated {
		t.Fatalf("failed to create inventory item: status %d, body: %s", itemW.Code, itemW.Body.String())
	}

	var itemResp map[string]interface{}
	if err := json.NewDecoder(itemW.Body).Decode(&itemResp); err != nil {
		t.Fatalf("failed to decode item response: %v", err)
	}

	itemObj, ok := itemResp["item"].(map[string]interface{})
	if !ok {
		t.Fatalf("item response missing item object: %v", itemResp)
	}

	itemID, ok := itemObj["id"].(string)
	if !ok {
		t.Fatalf("item response missing id: %v", itemObj)
	}

	t.Run("successful add item", func(t *testing.T) {
		body := `{"item_id": "` + itemID + `", "quantity": 1}`
		req := httptest.NewRequest("POST", "/api/v1/characters/"+characterID+"/inventory", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", characterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.AddItem).ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if inventory, ok := resp["inventory"].(map[string]interface{}); ok {
			if inventory["item_id"] != itemID {
				t.Errorf("expected item_id %s, got %v", itemID, inventory["item_id"])
			}
		} else {
			t.Error("response missing inventory")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"item_id": "` + itemID + `"}`
		req := httptest.NewRequest("POST", "/api/v1/characters/"+characterID+"/inventory", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("id", characterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.AddItem).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestInventoryHandler_ListInventory(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)
	characterRepo := postgres.NewCharacterRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	traitRepo := postgres.NewTraitRepository(db)
	characterTraitRepo := postgres.NewCharacterTraitRepository(db)
	rpgClassRepo := postgres.NewRPGClassRepository(db)
	rpgSystemRepo := postgres.NewRPGSystemRepository(db)
	createCharacterUseCase := characterapp.NewCreateCharacterUseCase(characterRepo, worldRepo, archetypeRepo, auditLogRepo, log)
	getCharacterUseCase := characterapp.NewGetCharacterUseCase(characterRepo, log)
	listCharactersUseCase := characterapp.NewListCharactersUseCase(characterRepo, log)
	updateCharacterUseCase := characterapp.NewUpdateCharacterUseCase(characterRepo, archetypeRepo, worldRepo, auditLogRepo, log)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo, worldRepo, auditLogRepo, log)
	addTraitUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, traitRepo, characterTraitRepo, log)
	removeTraitUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log)
	updateTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, traitRepo, log)
	getTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log)
	changeClassUseCase := rpgcharacterapp.NewChangeCharacterClassUseCase(characterRepo, rpgClassRepo, log)
	getAvailableClassesUseCase := rpgcharacterapp.NewGetAvailableClassesUseCase(characterRepo, rpgClassRepo, rpgSystemRepo, log)
	characterHandler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

	// Create character
	characterBody := `{"name": "Test Character"}`
	characterReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/characters", strings.NewReader(characterBody))
	characterReq.Header.Set("Content-Type", "application/json")
	characterReq.Header.Set("X-Tenant-ID", tenantID)
	characterReq.SetPathValue("world_id", worldID)
	characterW := httptest.NewRecorder()
	withTenantMiddleware(characterHandler.Create).ServeHTTP(characterW, characterReq)

	if characterW.Code != http.StatusCreated {
		t.Fatalf("failed to create character: status %d, body: %s", characterW.Code, characterW.Body.String())
	}

	var characterResp map[string]interface{}
	if err := json.NewDecoder(characterW.Body).Decode(&characterResp); err != nil {
		t.Fatalf("failed to decode character response: %v", err)
	}

	characterObj, ok := characterResp["character"].(map[string]interface{})
	if !ok {
		t.Fatalf("character response missing character object: %v", characterResp)
	}

	characterID, ok := characterObj["id"].(string)
	if !ok {
		t.Fatalf("character response missing id: %v", characterObj)
	}

	characterInventoryRepo := postgres.NewCharacterInventoryRepository(db)
	inventoryItemRepo := postgres.NewInventoryItemRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	inventorySlotRepo := postgres.NewInventorySlotRepository(db)
	createSlotUseCase := inventoryslotapp.NewCreateInventorySlotUseCase(inventorySlotRepo, rpgSystemRepo, log)
	listSlotsUseCase := inventoryslotapp.NewListInventorySlotsUseCase(inventorySlotRepo, log)
	createItemUseCase := inventoryitemapp.NewCreateInventoryItemUseCase(inventoryItemRepo, rpgSystemRepo, artifactRepo, log)
	getItemUseCase := inventoryitemapp.NewGetInventoryItemUseCase(inventoryItemRepo, log)
	listItemsUseCase := inventoryitemapp.NewListInventoryItemsUseCase(inventoryItemRepo, log)
	updateItemUseCase := inventoryitemapp.NewUpdateInventoryItemUseCase(inventoryItemRepo, log)
	deleteItemUseCase := inventoryitemapp.NewDeleteInventoryItemUseCase(inventoryItemRepo, log)
	addItemUseCase := characterinventoryapp.NewAddItemToInventoryUseCase(characterInventoryRepo, characterRepo, inventoryItemRepo, log)
	listInventoryUseCase := characterinventoryapp.NewListCharacterInventoryUseCase(characterInventoryRepo, log)
	updateInventoryUseCase := characterinventoryapp.NewUpdateCharacterInventoryUseCase(characterInventoryRepo, log)
	equipItemUseCase := characterinventoryapp.NewEquipItemUseCase(characterInventoryRepo, log)
	unequipItemUseCase := characterinventoryapp.NewUnequipItemUseCase(characterInventoryRepo, log)
	deleteInventoryUseCase := characterinventoryapp.NewDeleteCharacterInventoryUseCase(characterInventoryRepo, log)
	transferItemUseCase := characterinventoryapp.NewTransferItemUseCase(characterInventoryRepo, characterRepo, log)
	handler := NewInventoryHandler(createSlotUseCase, listSlotsUseCase, createItemUseCase, getItemUseCase, listItemsUseCase, updateItemUseCase, deleteItemUseCase, addItemUseCase, listInventoryUseCase, updateInventoryUseCase, equipItemUseCase, unequipItemUseCase, deleteInventoryUseCase, transferItemUseCase, log)

	t.Run("list inventory", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/characters/"+characterID+"/inventory", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", characterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.ListInventory).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if _, ok := resp["items"]; !ok {
			t.Error("response missing items")
		}
	})
}

func TestInventoryHandler_DeleteInventory(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)
	characterRepo := postgres.NewCharacterRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	traitRepo := postgres.NewTraitRepository(db)
	characterTraitRepo := postgres.NewCharacterTraitRepository(db)
	rpgClassRepo := postgres.NewRPGClassRepository(db)
	rpgSystemRepo := postgres.NewRPGSystemRepository(db)
	createCharacterUseCase := characterapp.NewCreateCharacterUseCase(characterRepo, worldRepo, archetypeRepo, auditLogRepo, log)
	getCharacterUseCase := characterapp.NewGetCharacterUseCase(characterRepo, log)
	listCharactersUseCase := characterapp.NewListCharactersUseCase(characterRepo, log)
	updateCharacterUseCase := characterapp.NewUpdateCharacterUseCase(characterRepo, archetypeRepo, worldRepo, auditLogRepo, log)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo, worldRepo, auditLogRepo, log)
	addTraitUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, traitRepo, characterTraitRepo, log)
	removeTraitUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log)
	updateTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, traitRepo, log)
	getTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log)
	changeClassUseCase := rpgcharacterapp.NewChangeCharacterClassUseCase(characterRepo, rpgClassRepo, log)
	getAvailableClassesUseCase := rpgcharacterapp.NewGetAvailableClassesUseCase(characterRepo, rpgClassRepo, rpgSystemRepo, log)
	characterHandler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

	// Create character
	characterBody := `{"name": "Test Character"}`
	characterReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/characters", strings.NewReader(characterBody))
	characterReq.Header.Set("Content-Type", "application/json")
	characterReq.Header.Set("X-Tenant-ID", tenantID)
	characterReq.SetPathValue("world_id", worldID)
	characterW := httptest.NewRecorder()
	withTenantMiddleware(characterHandler.Create).ServeHTTP(characterW, characterReq)

	if characterW.Code != http.StatusCreated {
		t.Fatalf("failed to create character: status %d, body: %s", characterW.Code, characterW.Body.String())
	}

	var characterResp map[string]interface{}
	if err := json.NewDecoder(characterW.Body).Decode(&characterResp); err != nil {
		t.Fatalf("failed to decode character response: %v", err)
	}

	characterObj, ok := characterResp["character"].(map[string]interface{})
	if !ok {
		t.Fatalf("character response missing character object: %v", characterResp)
	}

	characterID, ok := characterObj["id"].(string)
	if !ok {
		t.Fatalf("character response missing id: %v", characterObj)
	}

	// Create RPG system and inventory item
	tenantRepo := postgres.NewTenantRepository(db)
	createRPGSystemUseCase := rpgsystemapp.NewCreateRPGSystemUseCase(rpgSystemRepo, tenantRepo, log)
	getRPGSystemUseCase := rpgsystemapp.NewGetRPGSystemUseCase(rpgSystemRepo, log)
	listRPGSystemsUseCase := rpgsystemapp.NewListRPGSystemsUseCase(rpgSystemRepo, log)
	updateRPGSystemUseCase := rpgsystemapp.NewUpdateRPGSystemUseCase(rpgSystemRepo, log)
	deleteRPGSystemUseCase := rpgsystemapp.NewDeleteRPGSystemUseCase(rpgSystemRepo, log)
	rpgSystemHandler := NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)

	rpgSystemBody := `{"name": "Test RPG System", "base_stats_schema": {}}`
	rpgSystemReq := httptest.NewRequest("POST", "/api/v1/rpg-systems", strings.NewReader(rpgSystemBody))
	rpgSystemReq.Header.Set("Content-Type", "application/json")
	rpgSystemReq.Header.Set("X-Tenant-ID", tenantID)
	rpgSystemW := httptest.NewRecorder()
	withTenantMiddleware(rpgSystemHandler.Create).ServeHTTP(rpgSystemW, rpgSystemReq)

	if rpgSystemW.Code != http.StatusCreated {
		t.Fatalf("failed to create RPG system: status %d, body: %s", rpgSystemW.Code, rpgSystemW.Body.String())
	}

	var rpgSystemResp map[string]interface{}
	if err := json.NewDecoder(rpgSystemW.Body).Decode(&rpgSystemResp); err != nil {
		t.Fatalf("failed to decode RPG system response: %v", err)
	}

	rpgSystemObj, ok := rpgSystemResp["rpg_system"].(map[string]interface{})
	if !ok {
		t.Fatalf("RPG system response missing rpg_system object: %v", rpgSystemResp)
	}

	rpgSystemID, ok := rpgSystemObj["id"].(string)
	if !ok {
		t.Fatalf("RPG system response missing id: %v", rpgSystemObj)
	}

	inventorySlotRepo := postgres.NewInventorySlotRepository(db)
	inventoryItemRepo := postgres.NewInventoryItemRepository(db)
	characterInventoryRepo := postgres.NewCharacterInventoryRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	createSlotUseCase := inventoryslotapp.NewCreateInventorySlotUseCase(inventorySlotRepo, rpgSystemRepo, log)
	listSlotsUseCase := inventoryslotapp.NewListInventorySlotsUseCase(inventorySlotRepo, log)
	createItemUseCase := inventoryitemapp.NewCreateInventoryItemUseCase(inventoryItemRepo, rpgSystemRepo, artifactRepo, log)
	getItemUseCase := inventoryitemapp.NewGetInventoryItemUseCase(inventoryItemRepo, log)
	listItemsUseCase := inventoryitemapp.NewListInventoryItemsUseCase(inventoryItemRepo, log)
	updateItemUseCase := inventoryitemapp.NewUpdateInventoryItemUseCase(inventoryItemRepo, log)
	deleteItemUseCase := inventoryitemapp.NewDeleteInventoryItemUseCase(inventoryItemRepo, log)
	addItemUseCase := characterinventoryapp.NewAddItemToInventoryUseCase(characterInventoryRepo, characterRepo, inventoryItemRepo, log)
	listInventoryUseCase := characterinventoryapp.NewListCharacterInventoryUseCase(characterInventoryRepo, log)
	updateInventoryUseCase := characterinventoryapp.NewUpdateCharacterInventoryUseCase(characterInventoryRepo, log)
	equipItemUseCase := characterinventoryapp.NewEquipItemUseCase(characterInventoryRepo, log)
	unequipItemUseCase := characterinventoryapp.NewUnequipItemUseCase(characterInventoryRepo, log)
	deleteInventoryUseCase := characterinventoryapp.NewDeleteCharacterInventoryUseCase(characterInventoryRepo, log)
	transferItemUseCase := characterinventoryapp.NewTransferItemUseCase(characterInventoryRepo, characterRepo, log)
	handler := NewInventoryHandler(createSlotUseCase, listSlotsUseCase, createItemUseCase, getItemUseCase, listItemsUseCase, updateItemUseCase, deleteItemUseCase, addItemUseCase, listInventoryUseCase, updateInventoryUseCase, equipItemUseCase, unequipItemUseCase, deleteInventoryUseCase, transferItemUseCase, log)

	// Create inventory item
	itemBody := `{"name": "Test Item", "category": "weapon"}`
	itemReq := httptest.NewRequest("POST", "/api/v1/rpg-systems/"+rpgSystemID+"/inventory-items", strings.NewReader(itemBody))
	itemReq.Header.Set("Content-Type", "application/json")
	itemReq.Header.Set("X-Tenant-ID", tenantID)
	itemReq.SetPathValue("id", rpgSystemID)
	itemW := httptest.NewRecorder()
	withTenantMiddleware(handler.CreateItem).ServeHTTP(itemW, itemReq)

	if itemW.Code != http.StatusCreated {
		t.Fatalf("failed to create inventory item: status %d, body: %s", itemW.Code, itemW.Body.String())
	}

	var itemResp map[string]interface{}
	if err := json.NewDecoder(itemW.Body).Decode(&itemResp); err != nil {
		t.Fatalf("failed to decode item response: %v", err)
	}

	itemObj, ok := itemResp["item"].(map[string]interface{})
	if !ok {
		t.Fatalf("item response missing item object: %v", itemResp)
	}

	itemID, ok := itemObj["id"].(string)
	if !ok {
		t.Fatalf("item response missing id: %v", itemObj)
	}

	// Add item to inventory
	addItemBody := `{"item_id": "` + itemID + `", "quantity": 1}`
	addItemReq := httptest.NewRequest("POST", "/api/v1/characters/"+characterID+"/inventory", strings.NewReader(addItemBody))
	addItemReq.Header.Set("Content-Type", "application/json")
	addItemReq.Header.Set("X-Tenant-ID", tenantID)
	addItemReq.SetPathValue("id", characterID)
	addItemW := httptest.NewRecorder()
	withTenantMiddleware(handler.AddItem).ServeHTTP(addItemW, addItemReq)

	if addItemW.Code != http.StatusCreated {
		t.Fatalf("failed to add item to inventory: status %d, body: %s", addItemW.Code, addItemW.Body.String())
	}

	var addItemResp map[string]interface{}
	if err := json.NewDecoder(addItemW.Body).Decode(&addItemResp); err != nil {
		t.Fatalf("failed to decode add item response: %v", err)
	}

	inventoryObj, ok := addItemResp["inventory"].(map[string]interface{})
	if !ok {
		t.Fatalf("add item response missing inventory object: %v", addItemResp)
	}

	inventoryID, ok := inventoryObj["id"].(string)
	if !ok {
		t.Fatalf("add item response missing id: %v", inventoryObj)
	}

	t.Run("delete existing inventory item", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/character-inventory/"+inventoryID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", inventoryID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.DeleteInventory).ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}
	})

	t.Run("delete non-existing inventory item", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/character-inventory/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.DeleteInventory).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

