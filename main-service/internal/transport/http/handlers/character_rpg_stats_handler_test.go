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
	relationapp "github.com/story-engine/main-service/internal/application/relation"
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
	characterHandler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, getEventsUseCase, createRelationUseCase, getRelationUseCase, listRelationsBySourceUseCase, listRelationsByTargetUseCase, updateRelationUseCase, deleteRelationUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

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

		if stats, ok := resp["stats"].(map[string]interface{}); ok {
			if stats["character_id"] != characterID {
				t.Errorf("expected character_id %s, got %v", characterID, stats["character_id"])
			}
		} else {
			t.Error("response missing stats")
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

	traitRepo2 := postgres.NewTraitRepository(db)
	characterTraitRepo2 := postgres.NewCharacterTraitRepository(db)
	rpgClassRepo2 := postgres.NewRPGClassRepository(db)
	rpgSystemRepo2 := postgres.NewRPGSystemRepository(db)
	createCharacterUseCase := characterapp.NewCreateCharacterUseCase(characterRepo, worldRepo, archetypeRepo, auditLogRepo, log)
	getCharacterUseCase := characterapp.NewGetCharacterUseCase(characterRepo, log)
	listCharactersUseCase := characterapp.NewListCharactersUseCase(characterRepo, log)
	updateCharacterUseCase := characterapp.NewUpdateCharacterUseCase(characterRepo, archetypeRepo, worldRepo, auditLogRepo, log)
	entityRelationRepo2 := postgres.NewEntityRelationRepository(db)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo2, entityRelationRepo2, worldRepo, auditLogRepo, log)
	addTraitUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, traitRepo2, characterTraitRepo2, log)
	removeTraitUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo2, log)
	updateTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo2, traitRepo2, log)
	getTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo2, log)
	summaryGenerator2 := relationapp.NewSummaryGenerator()
	createRelationUseCase := relationapp.NewCreateRelationUseCase(entityRelationRepo2, summaryGenerator2, nil, log)
	getRelationUseCase := relationapp.NewGetRelationUseCase(entityRelationRepo2, log)
	listRelationsBySourceUseCase := relationapp.NewListRelationsBySourceUseCase(entityRelationRepo2, log)
	listRelationsByTargetUseCase := relationapp.NewListRelationsByTargetUseCase(entityRelationRepo2, log)
	updateRelationUseCase := relationapp.NewUpdateRelationUseCase(entityRelationRepo2, summaryGenerator2, nil, log)
	deleteRelationUseCase := relationapp.NewDeleteRelationUseCase(entityRelationRepo2, log)
	getEventsUseCase := characterapp.NewGetCharacterEventsUseCase(listRelationsByTargetUseCase, log)
	changeClassUseCase := rpgcharacterapp.NewChangeCharacterClassUseCase(characterRepo, rpgClassRepo2, log)
	getAvailableClassesUseCase := rpgcharacterapp.NewGetAvailableClassesUseCase(characterRepo, rpgClassRepo2, rpgSystemRepo2, log)
	characterHandler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, getEventsUseCase, createRelationUseCase, getRelationUseCase, listRelationsBySourceUseCase, listRelationsByTargetUseCase, updateRelationUseCase, deleteRelationUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

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

		if _, ok := resp["stats"]; !ok {
			t.Error("response missing stats")
		}
	})
}
