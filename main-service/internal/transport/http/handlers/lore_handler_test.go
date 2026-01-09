//go:build integration

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	loreapp "github.com/story-engine/main-service/internal/application/world/lore"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	rpgcharacterapp "github.com/story-engine/main-service/internal/application/rpg/character"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestLoreHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)

	loreRepo := postgres.NewLoreRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	eventRepo := postgres.NewEventRepository(db)
	factionRepo := postgres.NewFactionRepository(db)
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

	createLoreUseCase := loreapp.NewCreateLoreUseCase(loreRepo, worldRepo, auditLogRepo, log)
	getLoreUseCase := loreapp.NewGetLoreUseCase(loreRepo, log)
	listLoresUseCase := loreapp.NewListLoresUseCase(loreRepo, log)
	updateLoreUseCase := loreapp.NewUpdateLoreUseCase(loreRepo, auditLogRepo, log)
	deleteLoreUseCase := loreapp.NewDeleteLoreUseCase(loreRepo, entityRelationRepo, auditLogRepo, log)
	getLoreChildrenUseCase := loreapp.NewGetChildrenUseCase(loreRepo, log)
	addLoreReferenceUseCase := loreapp.NewAddReferenceUseCase(loreRepo, entityRelationRepo, createRelationUseCase, characterRepo, locationRepo, artifactRepo, eventRepo, factionRepo, log)
	removeLoreReferenceUseCase := loreapp.NewRemoveReferenceUseCase(listRelationsBySourceUseCase, deleteRelationUseCase, log)
	getLoreReferencesUseCase := loreapp.NewGetReferencesUseCase(listRelationsBySourceUseCase, log)
	updateLoreReferenceUseCase := loreapp.NewUpdateReferenceUseCase(getRelationUseCase, updateRelationUseCase, log)

	handler := NewLoreHandler(
		createLoreUseCase,
		getLoreUseCase,
		listLoresUseCase,
		updateLoreUseCase,
		deleteLoreUseCase,
		getLoreChildrenUseCase,
		addLoreReferenceUseCase,
		removeLoreReferenceUseCase,
		getLoreReferencesUseCase,
		updateLoreReferenceUseCase,
		log,
	)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"name": "Fire Magic", "category": "magic", "description": "The art of controlling flames"}`
		req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/lores", strings.NewReader(body))
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

		if lore, ok := resp["lore"].(map[string]interface{}); ok {
			if lore["name"] != "Fire Magic" {
				t.Errorf("expected name 'Fire Magic', got '%v'", lore["name"])
			}
		} else {
			t.Error("response missing lore")
		}
	})

	t.Run("empty name", func(t *testing.T) {
		body := `{"name": ""}`
		req := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/lores", strings.NewReader(body))
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

func TestLoreHandler_AddReference(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	worldID := setupTestWorld(t, db, tenantID)

	loreRepo := postgres.NewLoreRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	eventRepo := postgres.NewEventRepository(db)
	factionRepo := postgres.NewFactionRepository(db)
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

	createLoreUseCase := loreapp.NewCreateLoreUseCase(loreRepo, worldRepo, auditLogRepo, log)
	getLoreUseCase := loreapp.NewGetLoreUseCase(loreRepo, log)
	listLoresUseCase := loreapp.NewListLoresUseCase(loreRepo, log)
	updateLoreUseCase := loreapp.NewUpdateLoreUseCase(loreRepo, auditLogRepo, log)
	deleteLoreUseCase := loreapp.NewDeleteLoreUseCase(loreRepo, entityRelationRepo, auditLogRepo, log)
	getLoreChildrenUseCase := loreapp.NewGetChildrenUseCase(loreRepo, log)
	addLoreReferenceUseCase := loreapp.NewAddReferenceUseCase(loreRepo, entityRelationRepo, createRelationUseCase, characterRepo, locationRepo, artifactRepo, eventRepo, factionRepo, log)
	removeLoreReferenceUseCase := loreapp.NewRemoveReferenceUseCase(listRelationsBySourceUseCase, deleteRelationUseCase, log)
	getLoreReferencesUseCase := loreapp.NewGetReferencesUseCase(listRelationsBySourceUseCase, log)
	updateLoreReferenceUseCase := loreapp.NewUpdateReferenceUseCase(getRelationUseCase, updateRelationUseCase, log)

	handler := NewLoreHandler(
		createLoreUseCase,
		getLoreUseCase,
		listLoresUseCase,
		updateLoreUseCase,
		deleteLoreUseCase,
		getLoreChildrenUseCase,
		addLoreReferenceUseCase,
		removeLoreReferenceUseCase,
		getLoreReferencesUseCase,
		updateLoreReferenceUseCase,
		log,
	)

	// Create a lore first
	loreBody := `{"name": "Fire Magic"}`
	loreReq := httptest.NewRequest("POST", "/api/v1/worlds/"+worldID+"/lores", strings.NewReader(loreBody))
	loreReq.Header.Set("Content-Type", "application/json")
	loreReq.Header.Set("X-Tenant-ID", tenantID)
	loreReq.SetPathValue("world_id", worldID)
	loreW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(loreW, loreReq)

	if loreW.Code != http.StatusCreated {
		t.Fatalf("failed to create lore: status %d, body: %s", loreW.Code, loreW.Body.String())
	}

	var loreResp map[string]interface{}
	if err := json.NewDecoder(loreW.Body).Decode(&loreResp); err != nil {
		t.Fatalf("failed to decode lore response: %v", err)
	}

	loreObj, ok := loreResp["lore"].(map[string]interface{})
	if !ok {
		t.Fatalf("lore response missing lore object: %v", loreResp)
	}

	loreID, ok := loreObj["id"].(string)
	if !ok {
		t.Fatalf("lore response missing id: %v", loreObj)
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
		body := `{"entity_type": "character", "entity_id": "` + characterID + `", "relationship_type": "apprentice", "notes": "Learning fire magic"}`
		req := httptest.NewRequest("POST", "/api/v1/lores/"+loreID+"/references", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", loreID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.AddReference).ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("get references", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/lores/"+loreID+"/references", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", loreID)
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
