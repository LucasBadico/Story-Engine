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
	sceneapp "github.com/story-engine/main-service/internal/application/story/scene"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestSceneHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	sceneRepo := postgres.NewSceneRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	sceneReferenceRepo := postgres.NewSceneReferenceRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	log := logger.New()

	createSceneUseCase := sceneapp.NewCreateSceneUseCase(sceneRepo, chapterRepo, storyRepo, nil, log)
	getSceneUseCase := sceneapp.NewGetSceneUseCase(sceneRepo, log)
	updateSceneUseCase := sceneapp.NewUpdateSceneUseCase(sceneRepo, nil, log)
	deleteSceneUseCase := sceneapp.NewDeleteSceneUseCase(sceneRepo, log)
	listScenesUseCase := sceneapp.NewListScenesUseCase(sceneRepo, log)
	moveSceneUseCase := sceneapp.NewMoveSceneUseCase(sceneRepo, chapterRepo, log)
	addReferenceUC := sceneapp.NewAddSceneReferenceUseCase(sceneRepo, sceneReferenceRepo, characterRepo, locationRepo, artifactRepo, log)
	removeReferenceUC := sceneapp.NewRemoveSceneReferenceUseCase(sceneReferenceRepo, log)
	getReferencesUC := sceneapp.NewGetSceneReferencesUseCase(sceneReferenceRepo, log)
	handler := NewSceneHandler(createSceneUseCase, getSceneUseCase, updateSceneUseCase, deleteSceneUseCase, listScenesUseCase, moveSceneUseCase, addReferenceUC, removeReferenceUC, getReferencesUC, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"story_id": "` + storyID + `", "order_num": 1, "goal": "Test goal"}`
		req := httptest.NewRequest("POST", "/api/v1/scenes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if scene, ok := resp["scene"].(map[string]interface{}); ok {
			if scene["order_num"] != float64(1) {
				t.Errorf("expected order_num 1, got %v", scene["order_num"])
			}
		} else {
			t.Error("response missing scene")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"story_id": "` + storyID + `", "order_num": 1}`
		req := httptest.NewRequest("POST", "/api/v1/scenes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestSceneHandler_Get(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	sceneRepo := postgres.NewSceneRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	sceneReferenceRepo := postgres.NewSceneReferenceRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	log := logger.New()

	createSceneUseCase := sceneapp.NewCreateSceneUseCase(sceneRepo, chapterRepo, storyRepo, nil, log)
	getSceneUseCase := sceneapp.NewGetSceneUseCase(sceneRepo, log)
	updateSceneUseCase := sceneapp.NewUpdateSceneUseCase(sceneRepo, nil, log)
	deleteSceneUseCase := sceneapp.NewDeleteSceneUseCase(sceneRepo, log)
	listScenesUseCase := sceneapp.NewListScenesUseCase(sceneRepo, log)
	moveSceneUseCase := sceneapp.NewMoveSceneUseCase(sceneRepo, chapterRepo, log)
	addReferenceUC := sceneapp.NewAddSceneReferenceUseCase(sceneRepo, sceneReferenceRepo, characterRepo, locationRepo, artifactRepo, log)
	removeReferenceUC := sceneapp.NewRemoveSceneReferenceUseCase(sceneReferenceRepo, log)
	getReferencesUC := sceneapp.NewGetSceneReferencesUseCase(sceneReferenceRepo, log)
	handler := NewSceneHandler(createSceneUseCase, getSceneUseCase, updateSceneUseCase, deleteSceneUseCase, listScenesUseCase, moveSceneUseCase, addReferenceUC, removeReferenceUC, getReferencesUC, log)

	// Create scene
	sceneBody := `{"story_id": "` + storyID + `", "order_num": 1, "goal": "Get Test Scene"}`
	sceneReq := httptest.NewRequest("POST", "/api/v1/scenes", strings.NewReader(sceneBody))
	sceneReq.Header.Set("Content-Type", "application/json")
	sceneReq.Header.Set("X-Tenant-ID", tenantID)
	sceneW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(sceneW, sceneReq)

	if sceneW.Code != http.StatusCreated {
		t.Fatalf("failed to create scene: status %d, body: %s", sceneW.Code, sceneW.Body.String())
	}

	var sceneResp map[string]interface{}
	if err := json.NewDecoder(sceneW.Body).Decode(&sceneResp); err != nil {
		t.Fatalf("failed to decode scene response: %v", err)
	}

	sceneObj, ok := sceneResp["scene"].(map[string]interface{})
	if !ok {
		t.Fatalf("scene response missing scene object: %v", sceneResp)
	}

	sceneID, ok := sceneObj["id"].(string)
	if !ok {
		t.Fatalf("scene response missing id: %v", sceneObj)
	}

	t.Run("existing scene", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/scenes/"+sceneID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", sceneID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if scene, ok := resp["scene"].(map[string]interface{}); ok {
			if scene["id"] != sceneID {
				t.Errorf("expected ID %s, got %v", sceneID, scene["id"])
			}
		} else {
			t.Error("response missing scene")
		}
	})

	t.Run("non-existing scene", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/scenes/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

func TestSceneHandler_List(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	sceneRepo := postgres.NewSceneRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	sceneReferenceRepo := postgres.NewSceneReferenceRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	log := logger.New()

	createSceneUseCase := sceneapp.NewCreateSceneUseCase(sceneRepo, chapterRepo, storyRepo, nil, log)
	getSceneUseCase := sceneapp.NewGetSceneUseCase(sceneRepo, log)
	updateSceneUseCase := sceneapp.NewUpdateSceneUseCase(sceneRepo, nil, log)
	deleteSceneUseCase := sceneapp.NewDeleteSceneUseCase(sceneRepo, log)
	listScenesUseCase := sceneapp.NewListScenesUseCase(sceneRepo, log)
	moveSceneUseCase := sceneapp.NewMoveSceneUseCase(sceneRepo, chapterRepo, log)
	addReferenceUC := sceneapp.NewAddSceneReferenceUseCase(sceneRepo, sceneReferenceRepo, characterRepo, locationRepo, artifactRepo, log)
	removeReferenceUC := sceneapp.NewRemoveSceneReferenceUseCase(sceneReferenceRepo, log)
	getReferencesUC := sceneapp.NewGetSceneReferencesUseCase(sceneReferenceRepo, log)
	handler := NewSceneHandler(createSceneUseCase, getSceneUseCase, updateSceneUseCase, deleteSceneUseCase, listScenesUseCase, moveSceneUseCase, addReferenceUC, removeReferenceUC, getReferencesUC, log)

	// Create multiple scenes
	for i := 1; i <= 3; i++ {
		sceneBody := `{"story_id": "` + storyID + `", "order_num": ` + strconv.Itoa(i) + `, "goal": "Scene ` + strconv.Itoa(i) + `"}`
		sceneReq := httptest.NewRequest("POST", "/api/v1/scenes", strings.NewReader(sceneBody))
		sceneReq.Header.Set("Content-Type", "application/json")
		sceneReq.Header.Set("X-Tenant-ID", tenantID)
		sceneW := httptest.NewRecorder()
		withTenantMiddleware(handler.Create).ServeHTTP(sceneW, sceneReq)

		if sceneW.Code != http.StatusCreated {
			t.Fatalf("failed to create scene %d: status %d, body: %s", i, sceneW.Code, sceneW.Body.String())
		}
	}

	t.Run("list scenes", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/stories/"+storyID+"/scenes", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", storyID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.ListByStory).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if scenes, ok := resp["scenes"].([]interface{}); ok {
			if len(scenes) < 3 {
				t.Errorf("expected at least 3 scenes, got %d", len(scenes))
			}
		} else {
			t.Error("response missing scenes")
		}
	})
}

func TestSceneHandler_Update(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	sceneRepo := postgres.NewSceneRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	sceneReferenceRepo := postgres.NewSceneReferenceRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	log := logger.New()

	createSceneUseCase := sceneapp.NewCreateSceneUseCase(sceneRepo, chapterRepo, storyRepo, nil, log)
	getSceneUseCase := sceneapp.NewGetSceneUseCase(sceneRepo, log)
	updateSceneUseCase := sceneapp.NewUpdateSceneUseCase(sceneRepo, nil, log)
	deleteSceneUseCase := sceneapp.NewDeleteSceneUseCase(sceneRepo, log)
	listScenesUseCase := sceneapp.NewListScenesUseCase(sceneRepo, log)
	moveSceneUseCase := sceneapp.NewMoveSceneUseCase(sceneRepo, chapterRepo, log)
	addReferenceUC := sceneapp.NewAddSceneReferenceUseCase(sceneRepo, sceneReferenceRepo, characterRepo, locationRepo, artifactRepo, log)
	removeReferenceUC := sceneapp.NewRemoveSceneReferenceUseCase(sceneReferenceRepo, log)
	getReferencesUC := sceneapp.NewGetSceneReferencesUseCase(sceneReferenceRepo, log)
	handler := NewSceneHandler(createSceneUseCase, getSceneUseCase, updateSceneUseCase, deleteSceneUseCase, listScenesUseCase, moveSceneUseCase, addReferenceUC, removeReferenceUC, getReferencesUC, log)

	// Create scene
	sceneBody := `{"story_id": "` + storyID + `", "order_num": 1, "goal": "Original Scene"}`
	sceneReq := httptest.NewRequest("POST", "/api/v1/scenes", strings.NewReader(sceneBody))
	sceneReq.Header.Set("Content-Type", "application/json")
	sceneReq.Header.Set("X-Tenant-ID", tenantID)
	sceneW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(sceneW, sceneReq)

	if sceneW.Code != http.StatusCreated {
		t.Fatalf("failed to create scene: status %d, body: %s", sceneW.Code, sceneW.Body.String())
	}

	var sceneResp map[string]interface{}
	if err := json.NewDecoder(sceneW.Body).Decode(&sceneResp); err != nil {
		t.Fatalf("failed to decode scene response: %v", err)
	}

	sceneObj, ok := sceneResp["scene"].(map[string]interface{})
	if !ok {
		t.Fatalf("scene response missing scene object: %v", sceneResp)
	}

	sceneID, ok := sceneObj["id"].(string)
	if !ok {
		t.Fatalf("scene response missing id: %v", sceneObj)
	}

	t.Run("update scene goal", func(t *testing.T) {
		body := `{"goal": "Updated Scene"}`
		req := httptest.NewRequest("PUT", "/api/v1/scenes/"+sceneID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", sceneID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if scene, ok := resp["scene"].(map[string]interface{}); ok {
			if scene["goal"] != "Updated Scene" {
				t.Errorf("expected goal 'Updated Scene', got %v", scene["goal"])
			}
		} else {
			t.Error("response missing scene")
		}
	})

	t.Run("non-existing scene", func(t *testing.T) {
		body := `{"goal": "Non-existent"}`
		req := httptest.NewRequest("PUT", "/api/v1/scenes/00000000-0000-0000-0000-000000000000", strings.NewReader(body))
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

func TestSceneHandler_Delete(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	sceneRepo := postgres.NewSceneRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	sceneReferenceRepo := postgres.NewSceneReferenceRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	log := logger.New()

	createSceneUseCase := sceneapp.NewCreateSceneUseCase(sceneRepo, chapterRepo, storyRepo, nil, log)
	getSceneUseCase := sceneapp.NewGetSceneUseCase(sceneRepo, log)
	updateSceneUseCase := sceneapp.NewUpdateSceneUseCase(sceneRepo, nil, log)
	deleteSceneUseCase := sceneapp.NewDeleteSceneUseCase(sceneRepo, log)
	listScenesUseCase := sceneapp.NewListScenesUseCase(sceneRepo, log)
	moveSceneUseCase := sceneapp.NewMoveSceneUseCase(sceneRepo, chapterRepo, log)
	addReferenceUC := sceneapp.NewAddSceneReferenceUseCase(sceneRepo, sceneReferenceRepo, characterRepo, locationRepo, artifactRepo, log)
	removeReferenceUC := sceneapp.NewRemoveSceneReferenceUseCase(sceneReferenceRepo, log)
	getReferencesUC := sceneapp.NewGetSceneReferencesUseCase(sceneReferenceRepo, log)
	handler := NewSceneHandler(createSceneUseCase, getSceneUseCase, updateSceneUseCase, deleteSceneUseCase, listScenesUseCase, moveSceneUseCase, addReferenceUC, removeReferenceUC, getReferencesUC, log)

	// Create scene
	sceneBody := `{"story_id": "` + storyID + `", "order_num": 1, "goal": "Scene to Delete"}`
	sceneReq := httptest.NewRequest("POST", "/api/v1/scenes", strings.NewReader(sceneBody))
	sceneReq.Header.Set("Content-Type", "application/json")
	sceneReq.Header.Set("X-Tenant-ID", tenantID)
	sceneW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(sceneW, sceneReq)

	if sceneW.Code != http.StatusCreated {
		t.Fatalf("failed to create scene: status %d, body: %s", sceneW.Code, sceneW.Body.String())
	}

	var sceneResp map[string]interface{}
	if err := json.NewDecoder(sceneW.Body).Decode(&sceneResp); err != nil {
		t.Fatalf("failed to decode scene response: %v", err)
	}

	sceneObj, ok := sceneResp["scene"].(map[string]interface{})
	if !ok {
		t.Fatalf("scene response missing scene object: %v", sceneResp)
	}

	sceneID, ok := sceneObj["id"].(string)
	if !ok {
		t.Fatalf("scene response missing id: %v", sceneObj)
	}

	t.Run("delete existing scene", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/scenes/"+sceneID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", sceneID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}

		// Verify scene is deleted
		getReq := httptest.NewRequest("GET", "/api/v1/scenes/"+sceneID, nil)
		getReq.Header.Set("X-Tenant-ID", tenantID)
		getReq.SetPathValue("id", sceneID)
		getW := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(getW, getReq)

		if getW.Code != http.StatusNotFound {
			t.Errorf("expected status 404 when getting deleted scene, got %d", getW.Code)
		}
	})

	t.Run("delete non-existing scene", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/scenes/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

