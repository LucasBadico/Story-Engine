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
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestStoryHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantRepo := postgres.NewTenantRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
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

	createStoryUseCase := story.NewCreateStoryUseCase(storyRepo, tenantRepo, auditLogRepo, log)
	handler := NewStoryHandler(createStoryUseCase, nil, storyRepo, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"tenant_id": "` + tenantID + `", "title": "Test Story"}`
		req := httptest.NewRequest("POST", "/api/v1/stories", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Create(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
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
		w := httptest.NewRecorder()

		handler.Create(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("empty title", func(t *testing.T) {
		body := `{"tenant_id": "` + tenantID + `", "title": ""}`
		req := httptest.NewRequest("POST", "/api/v1/stories", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Create(w, req)

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
	auditLogRepo := postgres.NewAuditLogRepository(db)
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

	createStoryUseCase := story.NewCreateStoryUseCase(storyRepo, tenantRepo, auditLogRepo, log)
	storyBody := `{"tenant_id": "` + tenantID + `", "title": "Get Test Story"}`
	storyReq := httptest.NewRequest("POST", "/api/v1/stories", strings.NewReader(storyBody))
	storyReq.Header.Set("Content-Type", "application/json")
	storyW := httptest.NewRecorder()
	storyHandler := NewStoryHandler(createStoryUseCase, nil, storyRepo, log)
	storyHandler.Create(storyW, storyReq)

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
		req.SetPathValue("id", storyID)
		w := httptest.NewRecorder()

		storyHandler.Get(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
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
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		storyHandler.Get(w, req)

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
	auditLogRepo := postgres.NewAuditLogRepository(db)
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

	createStoryUseCase := story.NewCreateStoryUseCase(storyRepo, tenantRepo, auditLogRepo, log)
	handler := NewStoryHandler(createStoryUseCase, nil, storyRepo, log)

	// Create multiple stories
	for i := 1; i <= 3; i++ {
		storyBody := `{"tenant_id": "` + tenantID + `", "title": "Story ` + strconv.Itoa(i) + `"}`
		storyReq := httptest.NewRequest("POST", "/api/v1/stories", strings.NewReader(storyBody))
		storyReq.Header.Set("Content-Type", "application/json")
		storyW := httptest.NewRecorder()
		handler.Create(storyW, storyReq)

		if storyW.Code != http.StatusCreated {
			t.Fatalf("failed to create story %d: status %d, body: %s", i, storyW.Code, storyW.Body.String())
		}
	}

	t.Run("list stories", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/stories?tenant_id="+tenantID, nil)
		w := httptest.NewRecorder()

		handler.List(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
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
		w := httptest.NewRecorder()

		handler.List(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}
