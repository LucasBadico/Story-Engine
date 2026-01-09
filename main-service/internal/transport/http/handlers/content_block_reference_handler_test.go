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
	contentblockapp "github.com/story-engine/main-service/internal/application/story/content_block"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestContentBlockReferenceHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	contentBlockRepo := postgres.NewContentBlockRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	contentBlockReferenceRepo := postgres.NewContentBlockReferenceRepository(db)
	log := logger.New()

	createContentBlockUseCase := contentblockapp.NewCreateContentBlockUseCase(contentBlockRepo, chapterRepo, nil, log)
	getContentBlockUseCase := contentblockapp.NewGetContentBlockUseCase(contentBlockRepo, log)
	updateContentBlockUseCase := contentblockapp.NewUpdateContentBlockUseCase(contentBlockRepo, nil, log)
	deleteContentBlockUseCase := contentblockapp.NewDeleteContentBlockUseCase(contentBlockRepo, log)
	listContentBlocksUseCase := contentblockapp.NewListContentBlocksUseCase(contentBlockRepo, log)
	contentBlockHandler := NewContentBlockHandler(createContentBlockUseCase, getContentBlockUseCase, updateContentBlockUseCase, deleteContentBlockUseCase, listContentBlocksUseCase, log)

	// Create content block
	contentBlockBody := `{"type": "text", "kind": "draft", "content": "Test content", "order_num": 1}`
	contentBlockReq := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/content-blocks", strings.NewReader(contentBlockBody))
	contentBlockReq.Header.Set("Content-Type", "application/json")
	contentBlockReq.Header.Set("X-Tenant-ID", tenantID)
	contentBlockReq.SetPathValue("id", chapterID)
	contentBlockW := httptest.NewRecorder()
	withTenantMiddleware(contentBlockHandler.Create).ServeHTTP(contentBlockW, contentBlockReq)

	if contentBlockW.Code != http.StatusCreated {
		t.Fatalf("failed to create content block: status %d, body: %s", contentBlockW.Code, contentBlockW.Body.String())
	}

	var contentBlockResp map[string]interface{}
	if err := json.NewDecoder(contentBlockW.Body).Decode(&contentBlockResp); err != nil {
		t.Fatalf("failed to decode content block response: %v", err)
	}

	contentBlockObj, ok := contentBlockResp["content_block"].(map[string]interface{})
	if !ok {
		t.Fatalf("content block response missing content_block object: %v", contentBlockResp)
	}

	contentBlockID, ok := contentBlockObj["id"].(string)
	if !ok {
		t.Fatalf("content block response missing id: %v", contentBlockObj)
	}

	createReferenceUC := contentblockapp.NewCreateContentBlockReferenceUseCase(contentBlockReferenceRepo, contentBlockRepo, log)
	listByContentBlockUC := contentblockapp.NewListContentBlockReferencesByContentBlockUseCase(contentBlockReferenceRepo, contentBlockRepo, log)
	listByEntityUC := contentblockapp.NewListContentBlocksByEntityUseCase(contentBlockReferenceRepo, contentBlockRepo, log)
	deleteReferenceUC := contentblockapp.NewDeleteContentBlockReferenceUseCase(contentBlockReferenceRepo, log)
	handler := NewContentBlockReferenceHandler(createReferenceUC, listByContentBlockUC, listByEntityUC, deleteReferenceUC, log)

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
		req := httptest.NewRequest("POST", "/api/v1/content-blocks/"+contentBlockID+"/references", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", contentBlockID)
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
		req := httptest.NewRequest("POST", "/api/v1/content-blocks/"+contentBlockID+"/references", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("id", contentBlockID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestContentBlockReferenceHandler_ListByContentBlock(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	contentBlockRepo := postgres.NewContentBlockRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	contentBlockReferenceRepo := postgres.NewContentBlockReferenceRepository(db)
	log := logger.New()

	createContentBlockUseCase := contentblockapp.NewCreateContentBlockUseCase(contentBlockRepo, chapterRepo, nil, log)
	getContentBlockUseCase := contentblockapp.NewGetContentBlockUseCase(contentBlockRepo, log)
	updateContentBlockUseCase := contentblockapp.NewUpdateContentBlockUseCase(contentBlockRepo, nil, log)
	deleteContentBlockUseCase := contentblockapp.NewDeleteContentBlockUseCase(contentBlockRepo, log)
	listContentBlocksUseCase := contentblockapp.NewListContentBlocksUseCase(contentBlockRepo, log)
	contentBlockHandler := NewContentBlockHandler(createContentBlockUseCase, getContentBlockUseCase, updateContentBlockUseCase, deleteContentBlockUseCase, listContentBlocksUseCase, log)

	// Create content block
	contentBlockBody := `{"type": "text", "kind": "draft", "content": "Test content", "order_num": 1}`
	contentBlockReq := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/content-blocks", strings.NewReader(contentBlockBody))
	contentBlockReq.Header.Set("Content-Type", "application/json")
	contentBlockReq.Header.Set("X-Tenant-ID", tenantID)
	contentBlockReq.SetPathValue("id", chapterID)
	contentBlockW := httptest.NewRecorder()
	withTenantMiddleware(contentBlockHandler.Create).ServeHTTP(contentBlockW, contentBlockReq)

	if contentBlockW.Code != http.StatusCreated {
		t.Fatalf("failed to create content block: status %d, body: %s", contentBlockW.Code, contentBlockW.Body.String())
	}

	var contentBlockResp map[string]interface{}
	if err := json.NewDecoder(contentBlockW.Body).Decode(&contentBlockResp); err != nil {
		t.Fatalf("failed to decode content block response: %v", err)
	}

	contentBlockObj, ok := contentBlockResp["content_block"].(map[string]interface{})
	if !ok {
		t.Fatalf("content block response missing content_block object: %v", contentBlockResp)
	}

	contentBlockID, ok := contentBlockObj["id"].(string)
	if !ok {
		t.Fatalf("content block response missing id: %v", contentBlockObj)
	}

	createReferenceUC := contentblockapp.NewCreateContentBlockReferenceUseCase(contentBlockReferenceRepo, contentBlockRepo, log)
	listByContentBlockUC := contentblockapp.NewListContentBlockReferencesByContentBlockUseCase(contentBlockReferenceRepo, contentBlockRepo, log)
	listByEntityUC := contentblockapp.NewListContentBlocksByEntityUseCase(contentBlockReferenceRepo, contentBlockRepo, log)
	deleteReferenceUC := contentblockapp.NewDeleteContentBlockReferenceUseCase(contentBlockReferenceRepo, log)
	handler := NewContentBlockReferenceHandler(createReferenceUC, listByContentBlockUC, listByEntityUC, deleteReferenceUC, log)

	t.Run("list references", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/content-blocks/"+contentBlockID+"/references", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", contentBlockID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.ListByContentBlock).ServeHTTP(w, req)

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
