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
	chapterapp "github.com/story-engine/main-service/internal/application/story/chapter"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestChapterHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterRepo := postgres.NewChapterRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	log := logger.New()

	createChapterUseCase := chapterapp.NewCreateChapterUseCase(chapterRepo, storyRepo, nil, log)
	getChapterUseCase := chapterapp.NewGetChapterUseCase(chapterRepo, log)
	updateChapterUseCase := chapterapp.NewUpdateChapterUseCase(chapterRepo, nil, log)
	deleteChapterUseCase := chapterapp.NewDeleteChapterUseCase(chapterRepo, log)
	listChaptersUseCase := chapterapp.NewListChaptersUseCase(chapterRepo, log)
	handler := NewChapterHandler(createChapterUseCase, getChapterUseCase, updateChapterUseCase, deleteChapterUseCase, listChaptersUseCase, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"story_id": "` + storyID + `", "number": 1, "title": "Chapter 1"}`
		req := httptest.NewRequest("POST", "/api/v1/chapters", strings.NewReader(body))
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

		if chapter, ok := resp["chapter"].(map[string]interface{}); ok {
			if chapter["title"] != "Chapter 1" {
				t.Errorf("expected title 'Chapter 1', got %v", chapter["title"])
			}
		} else {
			t.Error("response missing chapter")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"story_id": "` + storyID + `", "number": 1, "title": "Chapter 1"}`
		req := httptest.NewRequest("POST", "/api/v1/chapters", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestChapterHandler_Get(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterRepo := postgres.NewChapterRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	log := logger.New()

	createChapterUseCase := chapterapp.NewCreateChapterUseCase(chapterRepo, storyRepo, nil, log)
	getChapterUseCase := chapterapp.NewGetChapterUseCase(chapterRepo, log)
	updateChapterUseCase := chapterapp.NewUpdateChapterUseCase(chapterRepo, nil, log)
	deleteChapterUseCase := chapterapp.NewDeleteChapterUseCase(chapterRepo, log)
	listChaptersUseCase := chapterapp.NewListChaptersUseCase(chapterRepo, log)
	handler := NewChapterHandler(createChapterUseCase, getChapterUseCase, updateChapterUseCase, deleteChapterUseCase, listChaptersUseCase, log)

	// Create chapter
	chapterBody := `{"story_id": "` + storyID + `", "number": 1, "title": "Get Test Chapter"}`
	chapterReq := httptest.NewRequest("POST", "/api/v1/chapters", strings.NewReader(chapterBody))
	chapterReq.Header.Set("Content-Type", "application/json")
	chapterReq.Header.Set("X-Tenant-ID", tenantID)
	chapterW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(chapterW, chapterReq)

	if chapterW.Code != http.StatusCreated {
		t.Fatalf("failed to create chapter: status %d, body: %s", chapterW.Code, chapterW.Body.String())
	}

	var chapterResp map[string]interface{}
	if err := json.NewDecoder(chapterW.Body).Decode(&chapterResp); err != nil {
		t.Fatalf("failed to decode chapter response: %v", err)
	}

	chapterObj, ok := chapterResp["chapter"].(map[string]interface{})
	if !ok {
		t.Fatalf("chapter response missing chapter object: %v", chapterResp)
	}

	chapterID, ok := chapterObj["id"].(string)
	if !ok {
		t.Fatalf("chapter response missing id: %v", chapterObj)
	}

	t.Run("existing chapter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/chapters/"+chapterID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", chapterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if chapter, ok := resp["chapter"].(map[string]interface{}); ok {
			if chapter["id"] != chapterID {
				t.Errorf("expected ID %s, got %v", chapterID, chapter["id"])
			}
		} else {
			t.Error("response missing chapter")
		}
	})

	t.Run("non-existing chapter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/chapters/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

func TestChapterHandler_List(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterRepo := postgres.NewChapterRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	log := logger.New()

	createChapterUseCase := chapterapp.NewCreateChapterUseCase(chapterRepo, storyRepo, nil, log)
	getChapterUseCase := chapterapp.NewGetChapterUseCase(chapterRepo, log)
	updateChapterUseCase := chapterapp.NewUpdateChapterUseCase(chapterRepo, nil, log)
	deleteChapterUseCase := chapterapp.NewDeleteChapterUseCase(chapterRepo, log)
	listChaptersUseCase := chapterapp.NewListChaptersUseCase(chapterRepo, log)
	handler := NewChapterHandler(createChapterUseCase, getChapterUseCase, updateChapterUseCase, deleteChapterUseCase, listChaptersUseCase, log)

	// Create multiple chapters
	for i := 1; i <= 3; i++ {
		chapterBody := `{"story_id": "` + storyID + `", "number": ` + strconv.Itoa(i) + `, "title": "Chapter ` + strconv.Itoa(i) + `"}`
		chapterReq := httptest.NewRequest("POST", "/api/v1/chapters", strings.NewReader(chapterBody))
		chapterReq.Header.Set("Content-Type", "application/json")
		chapterReq.Header.Set("X-Tenant-ID", tenantID)
		chapterW := httptest.NewRecorder()
		withTenantMiddleware(handler.Create).ServeHTTP(chapterW, chapterReq)

		if chapterW.Code != http.StatusCreated {
			t.Fatalf("failed to create chapter %d: status %d, body: %s", i, chapterW.Code, chapterW.Body.String())
		}
	}

	t.Run("list chapters", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/stories/"+storyID+"/chapters", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", storyID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.List).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if chapters, ok := resp["chapters"].([]interface{}); ok {
			if len(chapters) < 3 {
				t.Errorf("expected at least 3 chapters, got %d", len(chapters))
			}
		} else {
			t.Error("response missing chapters")
		}
	})
}

func TestChapterHandler_Update(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterRepo := postgres.NewChapterRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	log := logger.New()

	createChapterUseCase := chapterapp.NewCreateChapterUseCase(chapterRepo, storyRepo, nil, log)
	getChapterUseCase := chapterapp.NewGetChapterUseCase(chapterRepo, log)
	updateChapterUseCase := chapterapp.NewUpdateChapterUseCase(chapterRepo, nil, log)
	deleteChapterUseCase := chapterapp.NewDeleteChapterUseCase(chapterRepo, log)
	listChaptersUseCase := chapterapp.NewListChaptersUseCase(chapterRepo, log)
	handler := NewChapterHandler(createChapterUseCase, getChapterUseCase, updateChapterUseCase, deleteChapterUseCase, listChaptersUseCase, log)

	// Create chapter
	chapterBody := `{"story_id": "` + storyID + `", "number": 1, "title": "Original Chapter"}`
	chapterReq := httptest.NewRequest("POST", "/api/v1/chapters", strings.NewReader(chapterBody))
	chapterReq.Header.Set("Content-Type", "application/json")
	chapterReq.Header.Set("X-Tenant-ID", tenantID)
	chapterW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(chapterW, chapterReq)

	if chapterW.Code != http.StatusCreated {
		t.Fatalf("failed to create chapter: status %d, body: %s", chapterW.Code, chapterW.Body.String())
	}

	var chapterResp map[string]interface{}
	if err := json.NewDecoder(chapterW.Body).Decode(&chapterResp); err != nil {
		t.Fatalf("failed to decode chapter response: %v", err)
	}

	chapterObj, ok := chapterResp["chapter"].(map[string]interface{})
	if !ok {
		t.Fatalf("chapter response missing chapter object: %v", chapterResp)
	}

	chapterID, ok := chapterObj["id"].(string)
	if !ok {
		t.Fatalf("chapter response missing id: %v", chapterObj)
	}

	t.Run("update chapter title", func(t *testing.T) {
		body := `{"title": "Updated Chapter"}`
		req := httptest.NewRequest("PUT", "/api/v1/chapters/"+chapterID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", chapterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if chapter, ok := resp["chapter"].(map[string]interface{}); ok {
			if chapter["title"] != "Updated Chapter" {
				t.Errorf("expected title 'Updated Chapter', got %v", chapter["title"])
			}
		} else {
			t.Error("response missing chapter")
		}
	})

	t.Run("non-existing chapter", func(t *testing.T) {
		body := `{"title": "Non-existent"}`
		req := httptest.NewRequest("PUT", "/api/v1/chapters/00000000-0000-0000-0000-000000000000", strings.NewReader(body))
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

func TestChapterHandler_Delete(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterRepo := postgres.NewChapterRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	log := logger.New()

	createChapterUseCase := chapterapp.NewCreateChapterUseCase(chapterRepo, storyRepo, nil, log)
	getChapterUseCase := chapterapp.NewGetChapterUseCase(chapterRepo, log)
	updateChapterUseCase := chapterapp.NewUpdateChapterUseCase(chapterRepo, nil, log)
	deleteChapterUseCase := chapterapp.NewDeleteChapterUseCase(chapterRepo, log)
	listChaptersUseCase := chapterapp.NewListChaptersUseCase(chapterRepo, log)
	handler := NewChapterHandler(createChapterUseCase, getChapterUseCase, updateChapterUseCase, deleteChapterUseCase, listChaptersUseCase, log)

	// Create chapter
	chapterBody := `{"story_id": "` + storyID + `", "number": 1, "title": "Chapter to Delete"}`
	chapterReq := httptest.NewRequest("POST", "/api/v1/chapters", strings.NewReader(chapterBody))
	chapterReq.Header.Set("Content-Type", "application/json")
	chapterReq.Header.Set("X-Tenant-ID", tenantID)
	chapterW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(chapterW, chapterReq)

	if chapterW.Code != http.StatusCreated {
		t.Fatalf("failed to create chapter: status %d, body: %s", chapterW.Code, chapterW.Body.String())
	}

	var chapterResp map[string]interface{}
	if err := json.NewDecoder(chapterW.Body).Decode(&chapterResp); err != nil {
		t.Fatalf("failed to decode chapter response: %v", err)
	}

	chapterObj, ok := chapterResp["chapter"].(map[string]interface{})
	if !ok {
		t.Fatalf("chapter response missing chapter object: %v", chapterResp)
	}

	chapterID, ok := chapterObj["id"].(string)
	if !ok {
		t.Fatalf("chapter response missing id: %v", chapterObj)
	}

	t.Run("delete existing chapter", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/chapters/"+chapterID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", chapterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}

		// Verify chapter is deleted
		getReq := httptest.NewRequest("GET", "/api/v1/chapters/"+chapterID, nil)
		getReq.Header.Set("X-Tenant-ID", tenantID)
		getReq.SetPathValue("id", chapterID)
		getW := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(getW, getReq)

		if getW.Code != http.StatusNotFound {
			t.Errorf("expected status 404 when getting deleted chapter, got %d", getW.Code)
		}
	})

	t.Run("delete non-existing chapter", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/chapters/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

