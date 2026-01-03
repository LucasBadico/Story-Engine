//go:build integration

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	rpgcharacterapp "github.com/story-engine/main-service/internal/application/rpg/character"
	proseblockapp "github.com/story-engine/main-service/internal/application/story/prose_block"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestProseBlockReferenceHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	proseBlockRepo := postgres.NewProseBlockRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	proseBlockReferenceRepo := postgres.NewProseBlockReferenceRepository(db)
	log := logger.New()

	createProseBlockUseCase := proseblockapp.NewCreateProseBlockUseCase(proseBlockRepo, chapterRepo, log)
	getProseBlockUseCase := proseblockapp.NewGetProseBlockUseCase(proseBlockRepo, log)
	updateProseBlockUseCase := proseblockapp.NewUpdateProseBlockUseCase(proseBlockRepo, log)
	deleteProseBlockUseCase := proseblockapp.NewDeleteProseBlockUseCase(proseBlockRepo, log)
	listProseBlocksUseCase := proseblockapp.NewListProseBlocksUseCase(proseBlockRepo, log)
	proseBlockHandler := NewProseBlockHandler(createProseBlockUseCase, getProseBlockUseCase, updateProseBlockUseCase, deleteProseBlockUseCase, listProseBlocksUseCase, log)

	// Create prose block
	proseBlockBody := `{"kind": "draft", "content": "Test content", "order_num": 1}`
	proseBlockReq := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/prose-blocks", strings.NewReader(proseBlockBody))
	proseBlockReq.Header.Set("Content-Type", "application/json")
	proseBlockReq.Header.Set("X-Tenant-ID", tenantID)
	proseBlockReq.SetPathValue("id", chapterID)
	proseBlockW := httptest.NewRecorder()
	withTenantMiddleware(proseBlockHandler.Create).ServeHTTP(proseBlockW, proseBlockReq)

	if proseBlockW.Code != http.StatusCreated {
		t.Fatalf("failed to create prose block: status %d, body: %s", proseBlockW.Code, proseBlockW.Body.String())
	}

	var proseBlockResp map[string]interface{}
	if err := json.NewDecoder(proseBlockW.Body).Decode(&proseBlockResp); err != nil {
		t.Fatalf("failed to decode prose block response: %v", err)
	}

	proseBlockObj, ok := proseBlockResp["prose_block"].(map[string]interface{})
	if !ok {
		t.Fatalf("prose block response missing prose_block object: %v", proseBlockResp)
	}

	proseBlockID, ok := proseBlockObj["id"].(string)
	if !ok {
		t.Fatalf("prose block response missing id: %v", proseBlockObj)
	}

	createReferenceUC := proseblockapp.NewCreateProseBlockReferenceUseCase(proseBlockReferenceRepo, proseBlockRepo, log)
	listByProseBlockUC := proseblockapp.NewListProseBlockReferencesByProseBlockUseCase(proseBlockReferenceRepo, proseBlockRepo, log)
	listByEntityUC := proseblockapp.NewListProseBlocksByEntityUseCase(proseBlockReferenceRepo, proseBlockRepo, log)
	deleteReferenceUC := proseblockapp.NewDeleteProseBlockReferenceUseCase(proseBlockReferenceRepo, log)
	handler := NewProseBlockReferenceHandler(createReferenceUC, listByProseBlockUC, listByEntityUC, deleteReferenceUC, log)

	t.Run("successful creation", func(t *testing.T) {
		// Create a character to reference
		worldID := setupTestWorld(t, db, tenantID)
		characterRepo := postgres.NewCharacterRepository(db)
		worldRepo := postgres.NewWorldRepository(db)
		archetypeRepo := postgres.NewArchetypeRepository(db)
		auditLogRepo := postgres.NewAuditLogRepository(db)
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

		body := `{"entity_type": "character", "entity_id": "` + characterID + `"}`
		req := httptest.NewRequest("POST", "/api/v1/prose-blocks/"+proseBlockID+"/references", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", proseBlockID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if reference, ok := resp["reference"].(map[string]interface{}); ok {
			if reference["entity_type"] != "character" {
				t.Errorf("expected entity_type 'character', got %v", reference["entity_type"])
			}
		} else {
			t.Error("response missing reference")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"entity_type": "character", "entity_id": "00000000-0000-0000-0000-000000000000"}`
		req := httptest.NewRequest("POST", "/api/v1/prose-blocks/"+proseBlockID+"/references", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("id", proseBlockID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestProseBlockReferenceHandler_ListByProseBlock(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	proseBlockRepo := postgres.NewProseBlockRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	proseBlockReferenceRepo := postgres.NewProseBlockReferenceRepository(db)
	log := logger.New()

	createProseBlockUseCase := proseblockapp.NewCreateProseBlockUseCase(proseBlockRepo, chapterRepo, log)
	getProseBlockUseCase := proseblockapp.NewGetProseBlockUseCase(proseBlockRepo, log)
	updateProseBlockUseCase := proseblockapp.NewUpdateProseBlockUseCase(proseBlockRepo, log)
	deleteProseBlockUseCase := proseblockapp.NewDeleteProseBlockUseCase(proseBlockRepo, log)
	listProseBlocksUseCase := proseblockapp.NewListProseBlocksUseCase(proseBlockRepo, log)
	proseBlockHandler := NewProseBlockHandler(createProseBlockUseCase, getProseBlockUseCase, updateProseBlockUseCase, deleteProseBlockUseCase, listProseBlocksUseCase, log)

	// Create prose block
	proseBlockBody := `{"kind": "draft", "content": "Test content", "order_num": 1}`
	proseBlockReq := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/prose-blocks", strings.NewReader(proseBlockBody))
	proseBlockReq.Header.Set("Content-Type", "application/json")
	proseBlockReq.Header.Set("X-Tenant-ID", tenantID)
	proseBlockReq.SetPathValue("id", chapterID)
	proseBlockW := httptest.NewRecorder()
	withTenantMiddleware(proseBlockHandler.Create).ServeHTTP(proseBlockW, proseBlockReq)

	if proseBlockW.Code != http.StatusCreated {
		t.Fatalf("failed to create prose block: status %d, body: %s", proseBlockW.Code, proseBlockW.Body.String())
	}

	var proseBlockResp map[string]interface{}
	if err := json.NewDecoder(proseBlockW.Body).Decode(&proseBlockResp); err != nil {
		t.Fatalf("failed to decode prose block response: %v", err)
	}

	proseBlockObj, ok := proseBlockResp["prose_block"].(map[string]interface{})
	if !ok {
		t.Fatalf("prose block response missing prose_block object: %v", proseBlockResp)
	}

	proseBlockID, ok := proseBlockObj["id"].(string)
	if !ok {
		t.Fatalf("prose block response missing id: %v", proseBlockObj)
	}

	createReferenceUC := proseblockapp.NewCreateProseBlockReferenceUseCase(proseBlockReferenceRepo, proseBlockRepo, log)
	listByProseBlockUC := proseblockapp.NewListProseBlockReferencesByProseBlockUseCase(proseBlockReferenceRepo, proseBlockRepo, log)
	listByEntityUC := proseblockapp.NewListProseBlocksByEntityUseCase(proseBlockReferenceRepo, proseBlockRepo, log)
	deleteReferenceUC := proseblockapp.NewDeleteProseBlockReferenceUseCase(proseBlockReferenceRepo, log)
	handler := NewProseBlockReferenceHandler(createReferenceUC, listByProseBlockUC, listByEntityUC, deleteReferenceUC, log)

	t.Run("list references", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/prose-blocks/"+proseBlockID+"/references", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", proseBlockID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.ListByProseBlock).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if _, ok := resp["references"]; !ok {
			t.Error("response missing references")
		}
	})
}
