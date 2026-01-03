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
	"github.com/story-engine/main-service/internal/application/story"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/application/world"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestStoryHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantRepo := postgres.NewTenantRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	sceneRepo := postgres.NewSceneRepository(db)
	beatRepo := postgres.NewBeatRepository(db)
	proseBlockRepo := postgres.NewProseBlockRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	log := logger.New()

	// Create a tenant first
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	tenantBody := `{"name": "Test Tenant"}`
	tenantReq := httptest.NewRequest("POST", "/api/v1/tenants", strings.NewReader(tenantBody))
	tenantReq.Header.Set("Content-Type", "application/json")
	tenantW := httptest.NewRecorder()
	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	tenantHandler.Create(tenantW, tenantReq)

	if tenantW.Code != http.StatusCreated {
		t.Fatalf("failed to create tenant: status %d, body: %s", tenantW.Code, tenantW.Body.String())
	}

	var tenantResp map[string]interface{}
	if err := json.NewDecoder(tenantW.Body).Decode(&tenantResp); err != nil {
		t.Fatalf("failed to decode tenant response: %v", err)
	}

	tenantObj, ok := tenantResp["tenant"].(map[string]interface{})
	if !ok {
		t.Fatalf("tenant response missing tenant object: %v", tenantResp)
	}

	tenantID, ok := tenantObj["id"].(string)
	if !ok {
		t.Fatalf("tenant response missing id: %v", tenantObj)
	}

	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	createStoryUseCase := story.NewCreateStoryUseCase(storyRepo, tenantRepo, worldRepo, createWorldUseCase, auditLogRepo, log)
	getStoryUseCase := story.NewGetStoryUseCase(storyRepo, log)
	updateStoryUseCase := story.NewUpdateStoryUseCase(storyRepo, log)
	listStoriesUseCase := story.NewListStoriesUseCase(storyRepo, log)
	transactionRepo := postgres.NewTransactionRepository(db)
	cloneStoryUseCase := story.NewCloneStoryUseCase(storyRepo, chapterRepo, sceneRepo, beatRepo, proseBlockRepo, auditLogRepo, transactionRepo, log)
	handler := NewStoryHandler(createStoryUseCase, getStoryUseCase, updateStoryUseCase, listStoriesUseCase, cloneStoryUseCase, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"title": "Test Story"}`
		req := httptest.NewRequest("POST", "/api/v1/stories", strings.NewReader(body))
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

		if story, ok := resp["story"].(map[string]interface{}); ok {
			if story["title"] != "Test Story" {
				t.Errorf("expected title 'Test Story', got %v", story["title"])
			}
		} else {
			t.Error("response missing story")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"title": "Test Story"}`
		req := httptest.NewRequest("POST", "/api/v1/stories", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		// X-Tenant-ID header not set
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		// Middleware returns 400 (ValidationError) when X-Tenant-ID is missing
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("empty title", func(t *testing.T) {
		body := `{"title": ""}`
		req := httptest.NewRequest("POST", "/api/v1/stories", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestStoryHandler_Get(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantRepo := postgres.NewTenantRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	sceneRepo := postgres.NewSceneRepository(db)
	beatRepo := postgres.NewBeatRepository(db)
	proseBlockRepo := postgres.NewProseBlockRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	log := logger.New()

	// Create a tenant and story
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	tenantBody := `{"name": "Test Tenant"}`
	tenantReq := httptest.NewRequest("POST", "/api/v1/tenants", strings.NewReader(tenantBody))
	tenantReq.Header.Set("Content-Type", "application/json")
	tenantW := httptest.NewRecorder()
	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	tenantHandler.Create(tenantW, tenantReq)

	if tenantW.Code != http.StatusCreated {
		t.Fatalf("failed to create tenant: status %d, body: %s", tenantW.Code, tenantW.Body.String())
	}

	var tenantResp map[string]interface{}
	if err := json.NewDecoder(tenantW.Body).Decode(&tenantResp); err != nil {
		t.Fatalf("failed to decode tenant response: %v", err)
	}

	tenantObj, ok := tenantResp["tenant"].(map[string]interface{})
	if !ok {
		t.Fatalf("tenant response missing tenant object: %v", tenantResp)
	}

	tenantID, ok := tenantObj["id"].(string)
	if !ok {
		t.Fatalf("tenant response missing id: %v", tenantObj)
	}

	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	createStoryUseCase := story.NewCreateStoryUseCase(storyRepo, tenantRepo, worldRepo, createWorldUseCase, auditLogRepo, log)
	getStoryUseCase := story.NewGetStoryUseCase(storyRepo, log)
	updateStoryUseCase := story.NewUpdateStoryUseCase(storyRepo, log)
	listStoriesUseCase := story.NewListStoriesUseCase(storyRepo, log)
	transactionRepo := postgres.NewTransactionRepository(db)
	cloneStoryUseCase := story.NewCloneStoryUseCase(storyRepo, chapterRepo, sceneRepo, beatRepo, proseBlockRepo, auditLogRepo, transactionRepo, log)
	storyBody := `{"title": "Get Test Story"}`
	storyReq := httptest.NewRequest("POST", "/api/v1/stories", strings.NewReader(storyBody))
	storyReq.Header.Set("Content-Type", "application/json")
	storyReq.Header.Set("X-Tenant-ID", tenantID)
	storyW := httptest.NewRecorder()
	storyHandler := NewStoryHandler(createStoryUseCase, getStoryUseCase, updateStoryUseCase, listStoriesUseCase, cloneStoryUseCase, log)
	withTenantMiddleware(storyHandler.Create).ServeHTTP(storyW, storyReq)

	if storyW.Code != http.StatusCreated {
		t.Fatalf("failed to create story: status %d, body: %s", storyW.Code, storyW.Body.String())
	}

	var storyResp map[string]interface{}
	if err := json.NewDecoder(storyW.Body).Decode(&storyResp); err != nil {
		t.Fatalf("failed to decode story response: %v", err)
	}

	storyObj, ok := storyResp["story"].(map[string]interface{})
	if !ok {
		t.Fatalf("story response missing story object: %v", storyResp)
	}

	storyID, ok := storyObj["id"].(string)
	if !ok {
		t.Fatalf("story response missing id: %v", storyObj)
	}

	t.Run("existing story", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/stories/"+storyID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", storyID)
		w := httptest.NewRecorder()

		withTenantMiddleware(storyHandler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if story, ok := resp["story"].(map[string]interface{}); ok {
			if story["id"] != storyID {
				t.Errorf("expected ID %s, got %v", storyID, story["id"])
			}
		} else {
			t.Error("response missing story")
		}
	})

	t.Run("non-existing story", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/stories/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(storyHandler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

func TestStoryHandler_List(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantRepo := postgres.NewTenantRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	sceneRepo := postgres.NewSceneRepository(db)
	beatRepo := postgres.NewBeatRepository(db)
	proseBlockRepo := postgres.NewProseBlockRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	log := logger.New()

	// Create a tenant
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	tenantBody := `{"name": "Test Tenant"}`
	tenantReq := httptest.NewRequest("POST", "/api/v1/tenants", strings.NewReader(tenantBody))
	tenantReq.Header.Set("Content-Type", "application/json")
	tenantW := httptest.NewRecorder()
	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	tenantHandler.Create(tenantW, tenantReq)

	if tenantW.Code != http.StatusCreated {
		t.Fatalf("failed to create tenant: status %d, body: %s", tenantW.Code, tenantW.Body.String())
	}

	var tenantResp map[string]interface{}
	if err := json.NewDecoder(tenantW.Body).Decode(&tenantResp); err != nil {
		t.Fatalf("failed to decode tenant response: %v", err)
	}

	tenantObj, ok := tenantResp["tenant"].(map[string]interface{})
	if !ok {
		t.Fatalf("tenant response missing tenant object: %v", tenantResp)
	}

	tenantID, ok := tenantObj["id"].(string)
	if !ok {
		t.Fatalf("tenant response missing id: %v", tenantObj)
	}

	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	createStoryUseCase := story.NewCreateStoryUseCase(storyRepo, tenantRepo, worldRepo, createWorldUseCase, auditLogRepo, log)
	getStoryUseCase := story.NewGetStoryUseCase(storyRepo, log)
	updateStoryUseCase := story.NewUpdateStoryUseCase(storyRepo, log)
	listStoriesUseCase := story.NewListStoriesUseCase(storyRepo, log)
	transactionRepo := postgres.NewTransactionRepository(db)
	cloneStoryUseCase := story.NewCloneStoryUseCase(storyRepo, chapterRepo, sceneRepo, beatRepo, proseBlockRepo, auditLogRepo, transactionRepo, log)
	handler := NewStoryHandler(createStoryUseCase, getStoryUseCase, updateStoryUseCase, listStoriesUseCase, cloneStoryUseCase, log)

	// Create multiple stories
	for i := 1; i <= 3; i++ {
		storyBody := `{"title": "Story ` + strconv.Itoa(i) + `"}`
		storyReq := httptest.NewRequest("POST", "/api/v1/stories", strings.NewReader(storyBody))
		storyReq.Header.Set("Content-Type", "application/json")
		storyReq.Header.Set("X-Tenant-ID", tenantID)
		storyW := httptest.NewRecorder()
		withTenantMiddleware(handler.Create).ServeHTTP(storyW, storyReq)

		if storyW.Code != http.StatusCreated {
			t.Fatalf("failed to create story %d: status %d, body: %s", i, storyW.Code, storyW.Body.String())
		}
	}

	t.Run("list stories", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/stories", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.List).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if stories, ok := resp["stories"].([]interface{}); ok {
			if len(stories) != 3 {
				t.Errorf("expected 3 stories, got %d", len(stories))
			}
		} else {
			t.Error("response missing stories")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/stories", nil)
		// X-Tenant-ID header not set
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.List).ServeHTTP(w, req)

		// Middleware returns 400 (ValidationError) when X-Tenant-ID is missing
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}
