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
	relationapp "github.com/story-engine/main-service/internal/application/relation"
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
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo, entityRelationRepo, worldRepo, auditLogRepo, log)
	addTraitUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, traitRepo, characterTraitRepo, log)
	removeTraitUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log)
	updateTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, traitRepo, log)
	getTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log)
	// Entity relations use cases
	summaryGenerator := relationapp.NewSummaryGenerator()
	createRelationUseCase := relationapp.NewCreateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	getRelationUseCase := relationapp.NewGetRelationUseCase(entityRelationRepo, log)
	listRelationsBySourceUseCase := relationapp.NewListRelationsBySourceUseCase(entityRelationRepo, log)
	listRelationsByTargetUseCase := relationapp.NewListRelationsByTargetUseCase(entityRelationRepo, log)
	updateRelationUseCase := relationapp.NewUpdateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	deleteRelationUseCase := relationapp.NewDeleteRelationUseCase(entityRelationRepo, log)
	getEventsUseCase := characterapp.NewGetCharacterEventsUseCase(listRelationsByTargetUseCase, log)
	changeClassUseCase := rpgcharacterapp.NewChangeCharacterClassUseCase(characterRepo, rpgClassRepo, log)
	getAvailableClassesUseCase := rpgcharacterapp.NewGetAvailableClassesUseCase(characterRepo, rpgClassRepo, rpgSystemRepo, log)
	handler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, getEventsUseCase, createRelationUseCase, getRelationUseCase, listRelationsBySourceUseCase, listRelationsByTargetUseCase, updateRelationUseCase, deleteRelationUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

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
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo, entityRelationRepo, worldRepo, auditLogRepo, log)
	addTraitUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, traitRepo, characterTraitRepo, log)
	removeTraitUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log)
	updateTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, traitRepo, log)
	getTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log)
	// Entity relations use cases
	summaryGenerator := relationapp.NewSummaryGenerator()
	createRelationUseCase := relationapp.NewCreateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	getRelationUseCase := relationapp.NewGetRelationUseCase(entityRelationRepo, log)
	listRelationsBySourceUseCase := relationapp.NewListRelationsBySourceUseCase(entityRelationRepo, log)
	listRelationsByTargetUseCase := relationapp.NewListRelationsByTargetUseCase(entityRelationRepo, log)
	updateRelationUseCase := relationapp.NewUpdateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	deleteRelationUseCase := relationapp.NewDeleteRelationUseCase(entityRelationRepo, log)
	getEventsUseCase := characterapp.NewGetCharacterEventsUseCase(listRelationsByTargetUseCase, log)
	changeClassUseCase := rpgcharacterapp.NewChangeCharacterClassUseCase(characterRepo, rpgClassRepo, log)
	getAvailableClassesUseCase := rpgcharacterapp.NewGetAvailableClassesUseCase(characterRepo, rpgClassRepo, rpgSystemRepo, log)
	handler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, getEventsUseCase, createRelationUseCase, getRelationUseCase, listRelationsBySourceUseCase, listRelationsByTargetUseCase, updateRelationUseCase, deleteRelationUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

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
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo, entityRelationRepo, worldRepo, auditLogRepo, log)
	addTraitUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, traitRepo, characterTraitRepo, log)
	removeTraitUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log)
	updateTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, traitRepo, log)
	getTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log)
	// Entity relations use cases
	summaryGenerator := relationapp.NewSummaryGenerator()
	createRelationUseCase := relationapp.NewCreateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	getRelationUseCase := relationapp.NewGetRelationUseCase(entityRelationRepo, log)
	listRelationsBySourceUseCase := relationapp.NewListRelationsBySourceUseCase(entityRelationRepo, log)
	listRelationsByTargetUseCase := relationapp.NewListRelationsByTargetUseCase(entityRelationRepo, log)
	updateRelationUseCase := relationapp.NewUpdateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	deleteRelationUseCase := relationapp.NewDeleteRelationUseCase(entityRelationRepo, log)
	getEventsUseCase := characterapp.NewGetCharacterEventsUseCase(listRelationsByTargetUseCase, log)
	changeClassUseCase := rpgcharacterapp.NewChangeCharacterClassUseCase(characterRepo, rpgClassRepo, log)
	getAvailableClassesUseCase := rpgcharacterapp.NewGetAvailableClassesUseCase(characterRepo, rpgClassRepo, rpgSystemRepo, log)
	handler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, getEventsUseCase, createRelationUseCase, getRelationUseCase, listRelationsBySourceUseCase, listRelationsByTargetUseCase, updateRelationUseCase, deleteRelationUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

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
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo, entityRelationRepo, worldRepo, auditLogRepo, log)
	addTraitUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, traitRepo, characterTraitRepo, log)
	removeTraitUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log)
	updateTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, traitRepo, log)
	getTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log)
	// Entity relations use cases
	summaryGenerator := relationapp.NewSummaryGenerator()
	createRelationUseCase := relationapp.NewCreateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	getRelationUseCase := relationapp.NewGetRelationUseCase(entityRelationRepo, log)
	listRelationsBySourceUseCase := relationapp.NewListRelationsBySourceUseCase(entityRelationRepo, log)
	listRelationsByTargetUseCase := relationapp.NewListRelationsByTargetUseCase(entityRelationRepo, log)
	updateRelationUseCase := relationapp.NewUpdateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	deleteRelationUseCase := relationapp.NewDeleteRelationUseCase(entityRelationRepo, log)
	getEventsUseCase := characterapp.NewGetCharacterEventsUseCase(listRelationsByTargetUseCase, log)
	changeClassUseCase := rpgcharacterapp.NewChangeCharacterClassUseCase(characterRepo, rpgClassRepo, log)
	getAvailableClassesUseCase := rpgcharacterapp.NewGetAvailableClassesUseCase(characterRepo, rpgClassRepo, rpgSystemRepo, log)
	handler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, getEventsUseCase, createRelationUseCase, getRelationUseCase, listRelationsBySourceUseCase, listRelationsByTargetUseCase, updateRelationUseCase, deleteRelationUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

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
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo, entityRelationRepo, worldRepo, auditLogRepo, log)
	addTraitUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, traitRepo, characterTraitRepo, log)
	removeTraitUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log)
	updateTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, traitRepo, log)
	getTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log)
	// Entity relations use cases
	summaryGenerator := relationapp.NewSummaryGenerator()
	createRelationUseCase := relationapp.NewCreateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	getRelationUseCase := relationapp.NewGetRelationUseCase(entityRelationRepo, log)
	listRelationsBySourceUseCase := relationapp.NewListRelationsBySourceUseCase(entityRelationRepo, log)
	listRelationsByTargetUseCase := relationapp.NewListRelationsByTargetUseCase(entityRelationRepo, log)
	updateRelationUseCase := relationapp.NewUpdateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	deleteRelationUseCase := relationapp.NewDeleteRelationUseCase(entityRelationRepo, log)
	getEventsUseCase := characterapp.NewGetCharacterEventsUseCase(listRelationsByTargetUseCase, log)
	changeClassUseCase := rpgcharacterapp.NewChangeCharacterClassUseCase(characterRepo, rpgClassRepo, log)
	getAvailableClassesUseCase := rpgcharacterapp.NewGetAvailableClassesUseCase(characterRepo, rpgClassRepo, rpgSystemRepo, log)
	handler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, getEventsUseCase, createRelationUseCase, getRelationUseCase, listRelationsBySourceUseCase, listRelationsByTargetUseCase, updateRelationUseCase, deleteRelationUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

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

func TestCharacterHandler_ListRelationships(t *testing.T) {
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
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo, entityRelationRepo, worldRepo, auditLogRepo, log)
	addTraitUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, traitRepo, characterTraitRepo, log)
	removeTraitUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log)
	updateTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, traitRepo, log)
	getTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log)
	// Entity relations use cases
	summaryGenerator := relationapp.NewSummaryGenerator()
	createRelationUseCase := relationapp.NewCreateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	getRelationUseCase := relationapp.NewGetRelationUseCase(entityRelationRepo, log)
	listRelationsBySourceUseCase := relationapp.NewListRelationsBySourceUseCase(entityRelationRepo, log)
	listRelationsByTargetUseCase := relationapp.NewListRelationsByTargetUseCase(entityRelationRepo, log)
	updateRelationUseCase := relationapp.NewUpdateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	deleteRelationUseCase := relationapp.NewDeleteRelationUseCase(entityRelationRepo, log)
	getEventsUseCase := characterapp.NewGetCharacterEventsUseCase(listRelationsByTargetUseCase, log)
	changeClassUseCase := rpgcharacterapp.NewChangeCharacterClassUseCase(characterRepo, rpgClassRepo, log)
	getAvailableClassesUseCase := rpgcharacterapp.NewGetAvailableClassesUseCase(characterRepo, rpgClassRepo, rpgSystemRepo, log)
	handler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, getEventsUseCase, createRelationUseCase, getRelationUseCase, listRelationsBySourceUseCase, listRelationsByTargetUseCase, updateRelationUseCase, deleteRelationUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

	// Create two characters
	char1Body := `{"name": "Character 1"}`
	char1Req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/characters", strings.NewReader(char1Body))
	char1Req.Header.Set("Content-Type", "application/json")
	char1Req.Header.Set("X-Tenant-ID", tenantID)
	char1Req.SetPathValue("world_id", worldID)
	char1W := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(char1W, char1Req)
	if char1W.Code != http.StatusCreated {
		t.Fatalf("failed to create character 1: status %d", char1W.Code)
	}
	var char1Resp map[string]interface{}
	json.NewDecoder(char1W.Body).Decode(&char1Resp)
	char1ID := char1Resp["character"].(map[string]interface{})["id"].(string)

	char2Body := `{"name": "Character 2"}`
	char2Req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/characters", strings.NewReader(char2Body))
	char2Req.Header.Set("Content-Type", "application/json")
	char2Req.Header.Set("X-Tenant-ID", tenantID)
	char2Req.SetPathValue("world_id", worldID)
	char2W := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(char2W, char2Req)
	if char2W.Code != http.StatusCreated {
		t.Fatalf("failed to create character 2: status %d", char2W.Code)
	}
	var char2Resp map[string]interface{}
	json.NewDecoder(char2W.Body).Decode(&char2Resp)
	char2ID := char2Resp["character"].(map[string]interface{})["id"].(string)

	// Create a relationship
	relBody := `{"other_character_id": "` + char2ID + `", "relationship_type": "friend", "description": "They are friends", "bidirectional": false}`
	relReq := httptest.NewRequest("POST", "/api/v1/characters/"+char1ID+"/relationships", strings.NewReader(relBody))
	relReq.Header.Set("Content-Type", "application/json")
	relReq.Header.Set("X-Tenant-ID", tenantID)
	relReq.SetPathValue("id", char1ID)
	relW := httptest.NewRecorder()
	withTenantMiddleware(handler.CreateRelationship).ServeHTTP(relW, relReq)
	if relW.Code != http.StatusCreated {
		t.Fatalf("failed to create relationship: status %d, body: %s", relW.Code, relW.Body.String())
	}

	t.Run("list relationships", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/characters/"+char1ID+"/relationships", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", char1ID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.ListRelationships).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if relationships, ok := resp["relationships"].([]interface{}); ok {
			if len(relationships) < 1 {
				t.Errorf("expected at least 1 relationship, got %d", len(relationships))
			}
		} else {
			t.Error("response missing relationships")
		}
	})

	t.Run("list relationships for character with no relationships", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/characters/"+char2ID+"/relationships", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", char2ID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.ListRelationships).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if relationships, ok := resp["relationships"].([]interface{}); ok {
			if len(relationships) != 0 {
				t.Errorf("expected 0 relationships, got %d", len(relationships))
			}
		}
	})
}

func TestCharacterHandler_CreateRelationship(t *testing.T) {
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
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo, entityRelationRepo, worldRepo, auditLogRepo, log)
	addTraitUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, traitRepo, characterTraitRepo, log)
	removeTraitUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log)
	updateTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, traitRepo, log)
	getTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log)
	// Entity relations use cases
	summaryGenerator := relationapp.NewSummaryGenerator()
	createRelationUseCase := relationapp.NewCreateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	getRelationUseCase := relationapp.NewGetRelationUseCase(entityRelationRepo, log)
	listRelationsBySourceUseCase := relationapp.NewListRelationsBySourceUseCase(entityRelationRepo, log)
	listRelationsByTargetUseCase := relationapp.NewListRelationsByTargetUseCase(entityRelationRepo, log)
	updateRelationUseCase := relationapp.NewUpdateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	deleteRelationUseCase := relationapp.NewDeleteRelationUseCase(entityRelationRepo, log)
	getEventsUseCase := characterapp.NewGetCharacterEventsUseCase(listRelationsByTargetUseCase, log)
	changeClassUseCase := rpgcharacterapp.NewChangeCharacterClassUseCase(characterRepo, rpgClassRepo, log)
	getAvailableClassesUseCase := rpgcharacterapp.NewGetAvailableClassesUseCase(characterRepo, rpgClassRepo, rpgSystemRepo, log)
	handler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, getEventsUseCase, createRelationUseCase, getRelationUseCase, listRelationsBySourceUseCase, listRelationsByTargetUseCase, updateRelationUseCase, deleteRelationUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

	// Create two characters
	char1Body := `{"name": "Character 1"}`
	char1Req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/characters", strings.NewReader(char1Body))
	char1Req.Header.Set("Content-Type", "application/json")
	char1Req.Header.Set("X-Tenant-ID", tenantID)
	char1Req.SetPathValue("world_id", worldID)
	char1W := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(char1W, char1Req)
	if char1W.Code != http.StatusCreated {
		t.Fatalf("failed to create character 1: status %d", char1W.Code)
	}
	var char1Resp map[string]interface{}
	json.NewDecoder(char1W.Body).Decode(&char1Resp)
	char1ID := char1Resp["character"].(map[string]interface{})["id"].(string)

	char2Body := `{"name": "Character 2"}`
	char2Req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/characters", strings.NewReader(char2Body))
	char2Req.Header.Set("Content-Type", "application/json")
	char2Req.Header.Set("X-Tenant-ID", tenantID)
	char2Req.SetPathValue("world_id", worldID)
	char2W := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(char2W, char2Req)
	if char2W.Code != http.StatusCreated {
		t.Fatalf("failed to create character 2: status %d", char2W.Code)
	}
	var char2Resp map[string]interface{}
	json.NewDecoder(char2W.Body).Decode(&char2Resp)
	char2ID := char2Resp["character"].(map[string]interface{})["id"].(string)

	t.Run("create relationship", func(t *testing.T) {
		body := `{"other_character_id": "` + char2ID + `", "relationship_type": "friend", "description": "They are friends", "bidirectional": false}`
		req := httptest.NewRequest("POST", "/api/v1/characters/"+char1ID+"/relationships", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", char1ID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.CreateRelationship).ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["relationship_type"] != "friend" {
			t.Errorf("expected relationship_type 'friend', got %v", resp["relationship_type"])
		}
	})

	t.Run("create relationship with self", func(t *testing.T) {
		body := `{"other_character_id": "` + char1ID + `", "relationship_type": "self", "description": "Self relationship"}`
		req := httptest.NewRequest("POST", "/api/v1/characters/"+char1ID+"/relationships", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", char1ID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.CreateRelationship).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("create relationship missing relationship_type", func(t *testing.T) {
		body := `{"other_character_id": "` + char2ID + `", "description": "No type"}`
		req := httptest.NewRequest("POST", "/api/v1/characters/"+char1ID+"/relationships", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", char1ID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.CreateRelationship).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

