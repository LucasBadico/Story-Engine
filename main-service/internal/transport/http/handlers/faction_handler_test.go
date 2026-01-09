//go:build integration

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	factionapp "github.com/story-engine/main-service/internal/application/world/faction"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	rpgcharacterapp "github.com/story-engine/main-service/internal/application/rpg/character"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestFactionHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)

	factionRepo := postgres.NewFactionRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	eventRepo := postgres.NewEventRepository(db)
	loreRepo := postgres.NewLoreRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	log := logger.New()

	// Create relation use cases
	summaryGenerator := relationapp.NewSummaryGenerator()
	createRelationUseCase := relationapp.NewCreateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	getRelationUseCase := relationapp.NewGetRelationUseCase(entityRelationRepo, log)
	listRelationsBySourceUseCase := relationapp.NewListRelationsBySourceUseCase(entityRelationRepo, log)
	updateRelationUseCase := relationapp.NewUpdateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	deleteRelationUseCase := relationapp.NewDeleteRelationUseCase(entityRelationRepo, log)

	createFactionUseCase := factionapp.NewCreateFactionUseCase(factionRepo, worldRepo, auditLogRepo, log)
	getFactionUseCase := factionapp.NewGetFactionUseCase(factionRepo, log)
	listFactionsUseCase := factionapp.NewListFactionsUseCase(factionRepo, log)
	updateFactionUseCase := factionapp.NewUpdateFactionUseCase(factionRepo, auditLogRepo, log)
	deleteFactionUseCase := factionapp.NewDeleteFactionUseCase(factionRepo, entityRelationRepo, auditLogRepo, log)
	getFactionChildrenUseCase := factionapp.NewGetChildrenUseCase(factionRepo, log)
	addFactionReferenceUseCase := factionapp.NewAddReferenceUseCase(factionRepo, entityRelationRepo, createRelationUseCase, characterRepo, locationRepo, artifactRepo, eventRepo, loreRepo, log)
	removeFactionReferenceUseCase := factionapp.NewRemoveReferenceUseCase(listRelationsBySourceUseCase, deleteRelationUseCase, log)
	getFactionReferencesUseCase := factionapp.NewGetReferencesUseCase(listRelationsBySourceUseCase, log)
	updateFactionReferenceUseCase := factionapp.NewUpdateReferenceUseCase(getRelationUseCase, updateRelationUseCase, log)

	handler := NewFactionHandler(
		createFactionUseCase,
		getFactionUseCase,
		listFactionsUseCase,
		updateFactionUseCase,
		deleteFactionUseCase,
		getFactionChildrenUseCase,
		addFactionReferenceUseCase,
		removeFactionReferenceUseCase,
		getFactionReferencesUseCase,
		updateFactionReferenceUseCase,
		log,
	)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"name": "Guild of Mages", "type": "guild", "description": "A powerful mage guild"}`
		req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/factions", strings.NewReader(body))
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

		if faction, ok := resp["faction"].(map[string]interface{}); ok {
			if faction["name"] != "Guild of Mages" {
				t.Errorf("expected name 'Guild of Mages', got '%v'", faction["name"])
			}
		} else {
			t.Error("response missing faction")
		}
	})

	t.Run("empty name", func(t *testing.T) {
		body := `{"name": ""}`
		req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/factions", strings.NewReader(body))
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

func TestFactionHandler_AddReference(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)

	factionRepo := postgres.NewFactionRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	eventRepo := postgres.NewEventRepository(db)
	loreRepo := postgres.NewLoreRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	log := logger.New()

	// Create relation use cases
	summaryGenerator := relationapp.NewSummaryGenerator()
	createRelationUseCase := relationapp.NewCreateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	getRelationUseCase := relationapp.NewGetRelationUseCase(entityRelationRepo, log)
	listRelationsBySourceUseCase := relationapp.NewListRelationsBySourceUseCase(entityRelationRepo, log)
	updateRelationUseCase := relationapp.NewUpdateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	deleteRelationUseCase := relationapp.NewDeleteRelationUseCase(entityRelationRepo, log)

	createFactionUseCase := factionapp.NewCreateFactionUseCase(factionRepo, worldRepo, auditLogRepo, log)
	getFactionUseCase := factionapp.NewGetFactionUseCase(factionRepo, log)
	listFactionsUseCase := factionapp.NewListFactionsUseCase(factionRepo, log)
	updateFactionUseCase := factionapp.NewUpdateFactionUseCase(factionRepo, auditLogRepo, log)
	deleteFactionUseCase := factionapp.NewDeleteFactionUseCase(factionRepo, entityRelationRepo, auditLogRepo, log)
	getFactionChildrenUseCase := factionapp.NewGetChildrenUseCase(factionRepo, log)
	addFactionReferenceUseCase := factionapp.NewAddReferenceUseCase(factionRepo, entityRelationRepo, createRelationUseCase, characterRepo, locationRepo, artifactRepo, eventRepo, loreRepo, log)
	removeFactionReferenceUseCase := factionapp.NewRemoveReferenceUseCase(listRelationsBySourceUseCase, deleteRelationUseCase, log)
	getFactionReferencesUseCase := factionapp.NewGetReferencesUseCase(listRelationsBySourceUseCase, log)
	updateFactionReferenceUseCase := factionapp.NewUpdateReferenceUseCase(getRelationUseCase, updateRelationUseCase, log)

	handler := NewFactionHandler(
		createFactionUseCase,
		getFactionUseCase,
		listFactionsUseCase,
		updateFactionUseCase,
		deleteFactionUseCase,
		getFactionChildrenUseCase,
		addFactionReferenceUseCase,
		removeFactionReferenceUseCase,
		getFactionReferencesUseCase,
		updateFactionReferenceUseCase,
		log,
	)

	// Create a faction first
	factionBody := `{"name": "Guild of Mages"}`
	factionReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/factions", strings.NewReader(factionBody))
	factionReq.Header.Set("Content-Type", "application/json")
	factionReq.Header.Set("X-Tenant-ID", tenantID)
	factionReq.SetPathValue("world_id", worldID)
	factionW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(factionW, factionReq)

	if factionW.Code != http.StatusCreated {
		t.Fatalf("failed to create faction: status %d, body: %s", factionW.Code, factionW.Body.String())
	}

	var factionResp map[string]interface{}
	if err := json.NewDecoder(factionW.Body).Decode(&factionResp); err != nil {
		t.Fatalf("failed to decode faction response: %v", err)
	}

	factionObj, ok := factionResp["faction"].(map[string]interface{})
	if !ok {
		t.Fatalf("faction response missing faction object: %v", factionResp)
	}

	factionID, ok := factionObj["id"].(string)
	if !ok {
		t.Fatalf("faction response missing id: %v", factionObj)
	}

	// Create a character to reference
	archetypeRepo := postgres.NewArchetypeRepository(db)
	createCharacterUseCase := characterapp.NewCreateCharacterUseCase(characterRepo, worldRepo, archetypeRepo, auditLogRepo, log)
	getCharacterUseCase := characterapp.NewGetCharacterUseCase(characterRepo, log)
	listCharactersUseCase := characterapp.NewListCharactersUseCase(characterRepo, log)
	updateCharacterUseCase := characterapp.NewUpdateCharacterUseCase(characterRepo, archetypeRepo, worldRepo, auditLogRepo, log)
	characterTraitRepo := postgres.NewCharacterTraitRepository(db)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo, entityRelationRepo, worldRepo, auditLogRepo, log)
	traitRepo := postgres.NewTraitRepository(db)
	addTraitUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, traitRepo, characterTraitRepo, log)
	removeTraitUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log)
	updateTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, traitRepo, log)
	getTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log)
	listRelationsByTargetUseCase := relationapp.NewListRelationsByTargetUseCase(entityRelationRepo, log)
	getEventsUseCase := characterapp.NewGetCharacterEventsUseCase(listRelationsByTargetUseCase, log)
	rpgClassRepo := postgres.NewRPGClassRepository(db)
	rpgSystemRepo := postgres.NewRPGSystemRepository(db)
	changeClassUseCase := rpgcharacterapp.NewChangeCharacterClassUseCase(characterRepo, rpgClassRepo, log)
	getAvailableClassesUseCase := rpgcharacterapp.NewGetAvailableClassesUseCase(characterRepo, rpgClassRepo, rpgSystemRepo, log)
	characterHandler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitUseCase, removeTraitUseCase, updateTraitUseCase, getTraitsUseCase, getEventsUseCase, createRelationUseCase, getRelationUseCase, listRelationsBySourceUseCase, listRelationsByTargetUseCase, updateRelationUseCase, deleteRelationUseCase, changeClassUseCase, getAvailableClassesUseCase, log)

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

	t.Run("add character reference", func(t *testing.T) {
		body := `{"entity_type": "character", "entity_id": "` + characterID + `", "role": "archmage", "notes": "Leader of the guild"}`
		req := httptest.NewRequest("POST", "/api/v1/factions/"+factionID+"/references", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", factionID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.AddReference).ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("get references", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/factions/"+factionID+"/references", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", factionID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.GetReferences).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		references, ok := resp["references"].([]interface{})
		if !ok {
			t.Fatalf("response missing references array: %v", resp)
		}

		if len(references) == 0 {
			t.Error("expected at least one reference")
		}
	})
}
