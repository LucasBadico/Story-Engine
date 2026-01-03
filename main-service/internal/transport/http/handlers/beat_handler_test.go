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
	beatapp "github.com/story-engine/main-service/internal/application/story/beat"
	sceneapp "github.com/story-engine/main-service/internal/application/story/scene"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func setupTestScene(t *testing.T, db *postgres.DB, tenantID, storyID string) string {
	sceneRepo := postgres.NewSceneRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	sceneReferenceRepo := postgres.NewSceneReferenceRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	log := logger.New()

	createSceneUseCase := sceneapp.NewCreateSceneUseCase(sceneRepo, chapterRepo, storyRepo, log)
	getSceneUseCase := sceneapp.NewGetSceneUseCase(sceneRepo, log)
	updateSceneUseCase := sceneapp.NewUpdateSceneUseCase(sceneRepo, log)
	deleteSceneUseCase := sceneapp.NewDeleteSceneUseCase(sceneRepo, log)
	listScenesUseCase := sceneapp.NewListScenesUseCase(sceneRepo, log)
	moveSceneUseCase := sceneapp.NewMoveSceneUseCase(sceneRepo, chapterRepo, log)
	addReferenceUC := sceneapp.NewAddSceneReferenceUseCase(sceneRepo, sceneReferenceRepo, characterRepo, locationRepo, artifactRepo, log)
	removeReferenceUC := sceneapp.NewRemoveSceneReferenceUseCase(sceneReferenceRepo, log)
	getReferencesUC := sceneapp.NewGetSceneReferencesUseCase(sceneReferenceRepo, log)
	sceneHandler := NewSceneHandler(createSceneUseCase, getSceneUseCase, updateSceneUseCase, deleteSceneUseCase, listScenesUseCase, moveSceneUseCase, addReferenceUC, removeReferenceUC, getReferencesUC, log)

	sceneBody := `{"story_id": "` + storyID + `", "order_num": 1, "goal": "Test Scene"}`
	sceneReq := httptest.NewRequest("POST", "/api/v1/scenes", strings.NewReader(sceneBody))
	sceneReq.Header.Set("Content-Type", "application/json")
	sceneReq.Header.Set("X-Tenant-ID", tenantID)
	sceneW := httptest.NewRecorder()
	withTenantMiddleware(sceneHandler.Create).ServeHTTP(sceneW, sceneReq)

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

	return sceneID
}

func TestBeatHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	sceneID := setupTestScene(t, db, tenantID, storyID)
	beatRepo := postgres.NewBeatRepository(db)
	sceneRepo := postgres.NewSceneRepository(db)
	log := logger.New()

	createBeatUseCase := beatapp.NewCreateBeatUseCase(beatRepo, sceneRepo, log)
	getBeatUseCase := beatapp.NewGetBeatUseCase(beatRepo, log)
	updateBeatUseCase := beatapp.NewUpdateBeatUseCase(beatRepo, log)
	deleteBeatUseCase := beatapp.NewDeleteBeatUseCase(beatRepo, log)
	listBeatsUseCase := beatapp.NewListBeatsUseCase(beatRepo, log)
	moveBeatUseCase := beatapp.NewMoveBeatUseCase(beatRepo, sceneRepo, log)
	handler := NewBeatHandler(createBeatUseCase, getBeatUseCase, updateBeatUseCase, deleteBeatUseCase, listBeatsUseCase, moveBeatUseCase, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"scene_id": "` + sceneID + `", "order_num": 1, "type": "action", "intent": "Test intent", "outcome": "Test outcome"}`
		req := httptest.NewRequest("POST", "/api/v1/beats", strings.NewReader(body))
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

		if beat, ok := resp["beat"].(map[string]interface{}); ok {
			if beat["order_num"] != float64(1) {
				t.Errorf("expected order_num 1, got %v", beat["order_num"])
			}
		} else {
			t.Error("response missing beat")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"scene_id": "` + sceneID + `", "order_num": 1, "type": "action"}`
		req := httptest.NewRequest("POST", "/api/v1/beats", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestBeatHandler_Get(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	sceneID := setupTestScene(t, db, tenantID, storyID)
	beatRepo := postgres.NewBeatRepository(db)
	sceneRepo := postgres.NewSceneRepository(db)
	log := logger.New()

	createBeatUseCase := beatapp.NewCreateBeatUseCase(beatRepo, sceneRepo, log)
	getBeatUseCase := beatapp.NewGetBeatUseCase(beatRepo, log)
	updateBeatUseCase := beatapp.NewUpdateBeatUseCase(beatRepo, log)
	deleteBeatUseCase := beatapp.NewDeleteBeatUseCase(beatRepo, log)
	listBeatsUseCase := beatapp.NewListBeatsUseCase(beatRepo, log)
	moveBeatUseCase := beatapp.NewMoveBeatUseCase(beatRepo, sceneRepo, log)
	handler := NewBeatHandler(createBeatUseCase, getBeatUseCase, updateBeatUseCase, deleteBeatUseCase, listBeatsUseCase, moveBeatUseCase, log)

	// Create beat
	beatBody := `{"scene_id": "` + sceneID + `", "order_num": 1, "type": "action", "intent": "Get Test Beat"}`
	beatReq := httptest.NewRequest("POST", "/api/v1/beats", strings.NewReader(beatBody))
	beatReq.Header.Set("Content-Type", "application/json")
	beatReq.Header.Set("X-Tenant-ID", tenantID)
	beatW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(beatW, beatReq)

	if beatW.Code != http.StatusCreated {
		t.Fatalf("failed to create beat: status %d, body: %s", beatW.Code, beatW.Body.String())
	}

	var beatResp map[string]interface{}
	if err := json.NewDecoder(beatW.Body).Decode(&beatResp); err != nil {
		t.Fatalf("failed to decode beat response: %v", err)
	}

	beatObj, ok := beatResp["beat"].(map[string]interface{})
	if !ok {
		t.Fatalf("beat response missing beat object: %v", beatResp)
	}

	beatID, ok := beatObj["id"].(string)
	if !ok {
		t.Fatalf("beat response missing id: %v", beatObj)
	}

	t.Run("existing beat", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/beats/"+beatID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", beatID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if beat, ok := resp["beat"].(map[string]interface{}); ok {
			if beat["id"] != beatID {
				t.Errorf("expected ID %s, got %v", beatID, beat["id"])
			}
		} else {
			t.Error("response missing beat")
		}
	})

	t.Run("non-existing beat", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/beats/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

func TestBeatHandler_List(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	sceneID := setupTestScene(t, db, tenantID, storyID)
	beatRepo := postgres.NewBeatRepository(db)
	sceneRepo := postgres.NewSceneRepository(db)
	log := logger.New()

	createBeatUseCase := beatapp.NewCreateBeatUseCase(beatRepo, sceneRepo, log)
	getBeatUseCase := beatapp.NewGetBeatUseCase(beatRepo, log)
	updateBeatUseCase := beatapp.NewUpdateBeatUseCase(beatRepo, log)
	deleteBeatUseCase := beatapp.NewDeleteBeatUseCase(beatRepo, log)
	listBeatsUseCase := beatapp.NewListBeatsUseCase(beatRepo, log)
	moveBeatUseCase := beatapp.NewMoveBeatUseCase(beatRepo, sceneRepo, log)
	handler := NewBeatHandler(createBeatUseCase, getBeatUseCase, updateBeatUseCase, deleteBeatUseCase, listBeatsUseCase, moveBeatUseCase, log)

	// Create multiple beats
	for i := 1; i <= 3; i++ {
		beatBody := `{"scene_id": "` + sceneID + `", "order_num": ` + strconv.Itoa(i) + `, "type": "action", "intent": "Beat ` + strconv.Itoa(i) + `"}`
		beatReq := httptest.NewRequest("POST", "/api/v1/beats", strings.NewReader(beatBody))
		beatReq.Header.Set("Content-Type", "application/json")
		beatReq.Header.Set("X-Tenant-ID", tenantID)
		beatW := httptest.NewRecorder()
		withTenantMiddleware(handler.Create).ServeHTTP(beatW, beatReq)

		if beatW.Code != http.StatusCreated {
			t.Fatalf("failed to create beat %d: status %d, body: %s", i, beatW.Code, beatW.Body.String())
		}
	}

	t.Run("list beats", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/scenes/"+sceneID+"/beats", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("scene_id", sceneID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.List).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if beats, ok := resp["beats"].([]interface{}); ok {
			if len(beats) < 3 {
				t.Errorf("expected at least 3 beats, got %d", len(beats))
			}
		} else {
			t.Error("response missing beats")
		}
	})
}

func TestBeatHandler_Update(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	sceneID := setupTestScene(t, db, tenantID, storyID)
	beatRepo := postgres.NewBeatRepository(db)
	sceneRepo := postgres.NewSceneRepository(db)
	log := logger.New()

	createBeatUseCase := beatapp.NewCreateBeatUseCase(beatRepo, sceneRepo, log)
	getBeatUseCase := beatapp.NewGetBeatUseCase(beatRepo, log)
	updateBeatUseCase := beatapp.NewUpdateBeatUseCase(beatRepo, log)
	deleteBeatUseCase := beatapp.NewDeleteBeatUseCase(beatRepo, log)
	listBeatsUseCase := beatapp.NewListBeatsUseCase(beatRepo, log)
	moveBeatUseCase := beatapp.NewMoveBeatUseCase(beatRepo, sceneRepo, log)
	handler := NewBeatHandler(createBeatUseCase, getBeatUseCase, updateBeatUseCase, deleteBeatUseCase, listBeatsUseCase, moveBeatUseCase, log)

	// Create beat
	beatBody := `{"scene_id": "` + sceneID + `", "order_num": 1, "type": "action", "intent": "Original Beat"}`
	beatReq := httptest.NewRequest("POST", "/api/v1/beats", strings.NewReader(beatBody))
	beatReq.Header.Set("Content-Type", "application/json")
	beatReq.Header.Set("X-Tenant-ID", tenantID)
	beatW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(beatW, beatReq)

	if beatW.Code != http.StatusCreated {
		t.Fatalf("failed to create beat: status %d, body: %s", beatW.Code, beatW.Body.String())
	}

	var beatResp map[string]interface{}
	if err := json.NewDecoder(beatW.Body).Decode(&beatResp); err != nil {
		t.Fatalf("failed to decode beat response: %v", err)
	}

	beatObj, ok := beatResp["beat"].(map[string]interface{})
	if !ok {
		t.Fatalf("beat response missing beat object: %v", beatResp)
	}

	beatID, ok := beatObj["id"].(string)
	if !ok {
		t.Fatalf("beat response missing id: %v", beatObj)
	}

	t.Run("update beat intent", func(t *testing.T) {
		body := `{"intent": "Updated Beat"}`
		req := httptest.NewRequest("PUT", "/api/v1/beats/"+beatID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", beatID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if beat, ok := resp["beat"].(map[string]interface{}); ok {
			if beat["intent"] != "Updated Beat" {
				t.Errorf("expected intent 'Updated Beat', got %v", beat["intent"])
			}
		} else {
			t.Error("response missing beat")
		}
	})

	t.Run("non-existing beat", func(t *testing.T) {
		body := `{"intent": "Non-existent"}`
		req := httptest.NewRequest("PUT", "/api/v1/beats/00000000-0000-0000-0000-000000000000", strings.NewReader(body))
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

func TestBeatHandler_Delete(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	sceneID := setupTestScene(t, db, tenantID, storyID)
	beatRepo := postgres.NewBeatRepository(db)
	sceneRepo := postgres.NewSceneRepository(db)
	log := logger.New()

	createBeatUseCase := beatapp.NewCreateBeatUseCase(beatRepo, sceneRepo, log)
	getBeatUseCase := beatapp.NewGetBeatUseCase(beatRepo, log)
	updateBeatUseCase := beatapp.NewUpdateBeatUseCase(beatRepo, log)
	deleteBeatUseCase := beatapp.NewDeleteBeatUseCase(beatRepo, log)
	listBeatsUseCase := beatapp.NewListBeatsUseCase(beatRepo, log)
	moveBeatUseCase := beatapp.NewMoveBeatUseCase(beatRepo, sceneRepo, log)
	handler := NewBeatHandler(createBeatUseCase, getBeatUseCase, updateBeatUseCase, deleteBeatUseCase, listBeatsUseCase, moveBeatUseCase, log)

	// Create beat
	beatBody := `{"scene_id": "` + sceneID + `", "order_num": 1, "type": "action", "intent": "Beat to Delete"}`
	beatReq := httptest.NewRequest("POST", "/api/v1/beats", strings.NewReader(beatBody))
	beatReq.Header.Set("Content-Type", "application/json")
	beatReq.Header.Set("X-Tenant-ID", tenantID)
	beatW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(beatW, beatReq)

	if beatW.Code != http.StatusCreated {
		t.Fatalf("failed to create beat: status %d, body: %s", beatW.Code, beatW.Body.String())
	}

	var beatResp map[string]interface{}
	if err := json.NewDecoder(beatW.Body).Decode(&beatResp); err != nil {
		t.Fatalf("failed to decode beat response: %v", err)
	}

	beatObj, ok := beatResp["beat"].(map[string]interface{})
	if !ok {
		t.Fatalf("beat response missing beat object: %v", beatResp)
	}

	beatID, ok := beatObj["id"].(string)
	if !ok {
		t.Fatalf("beat response missing id: %v", beatObj)
	}

	t.Run("delete existing beat", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/beats/"+beatID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", beatID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}

		// Verify beat is deleted
		getReq := httptest.NewRequest("GET", "/api/v1/beats/"+beatID, nil)
		getReq.Header.Set("X-Tenant-ID", tenantID)
		getReq.SetPathValue("id", beatID)
		getW := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(getW, getReq)

		if getW.Code != http.StatusNotFound {
			t.Errorf("expected status 404 when getting deleted beat, got %d", getW.Code)
		}
	})

	t.Run("delete non-existing beat", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/beats/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

