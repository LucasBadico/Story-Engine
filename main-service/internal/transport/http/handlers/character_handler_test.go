//go:build integration

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	rpgcharacterapp "github.com/story-engine/main-service/internal/application/rpg/character"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestCharacterHandler_Create(t *testing.T) {
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
	handler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"name": "Test Character", "description": "A test character"}`
		req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/characters", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("world_id", worldID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if character, ok := resp["character"].(map[string]interface{}); ok {
			if character["name"] != "Test Character" {
				t.Errorf("expected name 'Test Character', got %v", character["name"])
			}
		} else {
			t.Error("response missing character")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"name": "Test Character"}`
		req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/characters", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("world_id", worldID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		body := `{"name": ""}`
		req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/characters", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("world_id", worldID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestCharacterHandler_Get(t *testing.T) {
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
	handler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

	// Create character
	characterBody := `{"name": "Get Test Character"}`
	characterReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/characters", strings.NewReader(characterBody))
	characterReq.Header.Set("Content-Type", "application/json")
	characterReq.Header.Set("X-Tenant-ID", tenantID)
	characterReq.SetPathValue("world_id", worldID)
	characterW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(characterW, characterReq)

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

	t.Run("existing character", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/characters/"+characterID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", characterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if character, ok := resp["character"].(map[string]interface{}); ok {
			if character["id"] != characterID {
				t.Errorf("expected ID %s, got %v", characterID, character["id"])
			}
		} else {
			t.Error("response missing character")
		}
	})

	t.Run("non-existing character", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/characters/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("invalid character ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/characters/not-a-uuid", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "not-a-uuid")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestCharacterHandler_List(t *testing.T) {
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
	handler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

	// Create multiple characters
	for i := 1; i <= 3; i++ {
		characterBody := `{"name": "Character ` + strconv.Itoa(i) + `"}`
		characterReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/characters", strings.NewReader(characterBody))
		characterReq.Header.Set("Content-Type", "application/json")
		characterReq.Header.Set("X-Tenant-ID", tenantID)
		characterReq.SetPathValue("world_id", worldID)
		characterW := httptest.NewRecorder()
		withTenantMiddleware(handler.Create).ServeHTTP(characterW, characterReq)

		if characterW.Code != http.StatusCreated {
			t.Fatalf("failed to create character %d: status %d, body: %s", i, characterW.Code, characterW.Body.String())
		}
	}

	t.Run("list characters", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/worlds/"+worldID+"/characters?limit=10&offset=0", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("world_id", worldID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.List).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if characters, ok := resp["characters"].([]interface{}); ok {
			if len(characters) < 3 {
				t.Errorf("expected at least 3 characters, got %d", len(characters))
			}
		} else {
			t.Error("response missing characters")
		}

		if total, ok := resp["total"].(float64); ok {
			if total < 3 {
				t.Errorf("expected total at least 3, got %v", total)
			}
		} else {
			t.Error("response missing total")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/worlds/"+worldID+"/characters", nil)
		req.SetPathValue("world_id", worldID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.List).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestCharacterHandler_Update(t *testing.T) {
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
	handler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

	// Create character
	characterBody := `{"name": "Original Character"}`
	characterReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/characters", strings.NewReader(characterBody))
	characterReq.Header.Set("Content-Type", "application/json")
	characterReq.Header.Set("X-Tenant-ID", tenantID)
	characterReq.SetPathValue("world_id", worldID)
	characterW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(characterW, characterReq)

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

	t.Run("update character name", func(t *testing.T) {
		body := `{"name": "Updated Character"}`
		req := httptest.NewRequest("PUT", "/api/v1/characters/"+characterID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", characterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if character, ok := resp["character"].(map[string]interface{}); ok {
			if character["name"] != "Updated Character" {
				t.Errorf("expected name 'Updated Character', got %v", character["name"])
			}
		} else {
			t.Error("response missing character")
		}
	})

	t.Run("non-existing character", func(t *testing.T) {
		body := `{"name": "Non-existent"}`
		req := httptest.NewRequest("PUT", "/api/v1/characters/00000000-0000-0000-0000-000000000000", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

func TestCharacterHandler_Delete(t *testing.T) {
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
	handler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

	// Create character
	characterBody := `{"name": "Character to Delete"}`
	characterReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/characters", strings.NewReader(characterBody))
	characterReq.Header.Set("Content-Type", "application/json")
	characterReq.Header.Set("X-Tenant-ID", tenantID)
	characterReq.SetPathValue("world_id", worldID)
	characterW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(characterW, characterReq)

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

	t.Run("delete existing character", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/characters/"+characterID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", characterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}

		// Verify character is deleted
		getReq := httptest.NewRequest("GET", "/api/v1/characters/"+characterID, nil)
		getReq.Header.Set("X-Tenant-ID", tenantID)
		getReq.SetPathValue("id", characterID)
		getW := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(getW, getReq)

		if getW.Code != http.StatusNotFound {
			t.Errorf("expected status 404 when getting deleted character, got %d", getW.Code)
		}
	})

	t.Run("delete non-existing character", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/characters/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

