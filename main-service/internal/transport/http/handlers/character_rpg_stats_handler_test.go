//go:build integration

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	characterstatsapp "github.com/story-engine/main-service/internal/application/rpg/character_stats"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	rpgcharacterapp "github.com/story-engine/main-service/internal/application/rpg/character"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestCharacterRPGStatsHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)
	characterRepo := postgres.NewCharacterRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createCharacterUseCase := characterapp.NewCreateCharacterUseCase(characterRepo, worldRepo, archetypeRepo, auditLogRepo, log)
	getCharacterUseCase := characterapp.NewGetCharacterUseCase(characterRepo, log)
	listCharactersUseCase := characterapp.NewListCharactersUseCase(characterRepo, log)
	updateCharacterUseCase := characterapp.NewUpdateCharacterUseCase(characterRepo, archetypeRepo, worldRepo, auditLogRepo, log)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, postgres.NewCharacterTraitRepository(db), worldRepo, auditLogRepo, log)
	addTraitUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, postgres.NewTraitRepository(db), postgres.NewCharacterTraitRepository(db), log)
	removeTraitUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(postgres.NewCharacterTraitRepository(db), log)
	updateTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(postgres.NewCharacterTraitRepository(db), log)
	getTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(postgres.NewCharacterTraitRepository(db), log)
	changeClassUseCase := rpgcharacterapp.NewChangeCharacterClassUseCase(characterRepo, postgres.NewRPGClassRepository(db), auditLogRepo, log)
	getAvailableClassesUseCase := rpgcharacterapp.NewGetAvailableClassesUseCase(characterRepo, postgres.NewRPGClassRepository(db), log)
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

	characterStatsRepo := postgres.NewCharacterRPGStatsRepository(db)
	eventRepo := postgres.NewEventRepository(db)
	createStatsUseCase := characterstatsapp.NewCreateCharacterStatsUseCase(characterStatsRepo, characterRepo, eventRepo, log)
	getActiveStatsUseCase := characterstatsapp.NewGetActiveCharacterStatsUseCase(characterStatsRepo, log)
	listHistoryUseCase := characterstatsapp.NewListCharacterStatsHistoryUseCase(characterStatsRepo, log)
	activateVersionUseCase := characterstatsapp.NewActivateCharacterStatsVersionUseCase(characterStatsRepo, log)
	deleteAllStatsUseCase := characterstatsapp.NewDeleteAllCharacterStatsUseCase(characterStatsRepo, log)
	handler := NewCharacterRPGStatsHandler(createStatsUseCase, getActiveStatsUseCase, listHistoryUseCase, activateVersionUseCase, deleteAllStatsUseCase, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"base_stats": {"strength": 15, "dexterity": 12}}`
		req := httptest.NewRequest("POST", "/api/v1/characters/"+characterID+"/rpg-stats", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", characterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if stats, ok := resp["character_rpg_stats"].(map[string]interface{}); ok {
			if stats["character_id"] != characterID {
				t.Errorf("expected character_id %s, got %v", characterID, stats["character_id"])
			}
		} else {
			t.Error("response missing character_rpg_stats")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"base_stats": {"strength": 15}}`
		req := httptest.NewRequest("POST", "/api/v1/characters/"+characterID+"/rpg-stats", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("id", characterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestCharacterRPGStatsHandler_GetActive(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)
	characterRepo := postgres.NewCharacterRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createCharacterUseCase := characterapp.NewCreateCharacterUseCase(characterRepo, worldRepo, archetypeRepo, auditLogRepo, log)
	getCharacterUseCase := characterapp.NewGetCharacterUseCase(characterRepo, log)
	listCharactersUseCase := characterapp.NewListCharactersUseCase(characterRepo, log)
	updateCharacterUseCase := characterapp.NewUpdateCharacterUseCase(characterRepo, archetypeRepo, worldRepo, auditLogRepo, log)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, postgres.NewCharacterTraitRepository(db), worldRepo, auditLogRepo, log)
	addTraitUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, postgres.NewTraitRepository(db), postgres.NewCharacterTraitRepository(db), log)
	removeTraitUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(postgres.NewCharacterTraitRepository(db), log)
	updateTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(postgres.NewCharacterTraitRepository(db), log)
	getTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(postgres.NewCharacterTraitRepository(db), log)
	changeClassUseCase := rpgcharacterapp.NewChangeCharacterClassUseCase(characterRepo, postgres.NewRPGClassRepository(db), auditLogRepo, log)
	getAvailableClassesUseCase := rpgcharacterapp.NewGetAvailableClassesUseCase(characterRepo, postgres.NewRPGClassRepository(db), log)
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

	characterStatsRepo := postgres.NewCharacterRPGStatsRepository(db)
	eventRepo := postgres.NewEventRepository(db)
	createStatsUseCase := characterstatsapp.NewCreateCharacterStatsUseCase(characterStatsRepo, characterRepo, eventRepo, log)
	getActiveStatsUseCase := characterstatsapp.NewGetActiveCharacterStatsUseCase(characterStatsRepo, log)
	listHistoryUseCase := characterstatsapp.NewListCharacterStatsHistoryUseCase(characterStatsRepo, log)
	activateVersionUseCase := characterstatsapp.NewActivateCharacterStatsVersionUseCase(characterStatsRepo, log)
	deleteAllStatsUseCase := characterstatsapp.NewDeleteAllCharacterStatsUseCase(characterStatsRepo, log)
	handler := NewCharacterRPGStatsHandler(createStatsUseCase, getActiveStatsUseCase, listHistoryUseCase, activateVersionUseCase, deleteAllStatsUseCase, log)

	// Create stats
	createBody := `{"base_stats": {"strength": 15, "dexterity": 12}}`
	createReq := httptest.NewRequest("POST", "/api/v1/characters/"+characterID+"/rpg-stats", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-Tenant-ID", tenantID)
	createReq.SetPathValue("id", characterID)
	createW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(createW, createReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("failed to create stats: status %d, body: %s", createW.Code, createW.Body.String())
	}

	t.Run("get active stats", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/characters/"+characterID+"/rpg-stats", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", characterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.GetActive).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if _, ok := resp["character_rpg_stats"]; !ok {
			t.Error("response missing character_rpg_stats")
		}
	})
}

