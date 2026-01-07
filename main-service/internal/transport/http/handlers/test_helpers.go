//go:build integration

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	chapterapp "github.com/story-engine/main-service/internal/application/story/chapter"
	"github.com/story-engine/main-service/internal/application/story"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/application/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// withTenantMiddleware wraps a handler function with the TenantMiddleware
func withTenantMiddleware(handlerFunc http.HandlerFunc) http.Handler {
	return middleware.TenantMiddleware(handlerFunc)
}

// setupTestTenant creates a tenant and returns its ID
func setupTestTenant(t *testing.T, db *postgres.DB) string {
	tenantRepo := postgres.NewTenantRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

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

	return tenantID
}

// setupTestWorld creates a world and returns its ID
func setupTestWorld(t *testing.T, db *postgres.DB, tenantID string) string {
	worldRepo := postgres.NewWorldRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	getWorldUseCase := world.NewGetWorldUseCase(worldRepo, log)
	listWorldsUseCase := world.NewListWorldsUseCase(worldRepo, log)
	updateWorldUseCase := world.NewUpdateWorldUseCase(worldRepo, auditLogRepo, log)
	deleteWorldUseCase := world.NewDeleteWorldUseCase(worldRepo, auditLogRepo, log)
	worldHandler := NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)

	worldBody := `{"name": "Test World"}`
	worldReq := httptest.NewRequest("POST", "/api/v1/worlds", strings.NewReader(worldBody))
	worldReq.Header.Set("Content-Type", "application/json")
	worldReq.Header.Set("X-Tenant-ID", tenantID)
	worldW := httptest.NewRecorder()
	withTenantMiddleware(worldHandler.Create).ServeHTTP(worldW, worldReq)

	if worldW.Code != http.StatusCreated {
		t.Fatalf("failed to create world: status %d, body: %s", worldW.Code, worldW.Body.String())
	}

	var worldResp map[string]interface{}
	if err := json.NewDecoder(worldW.Body).Decode(&worldResp); err != nil {
		t.Fatalf("failed to decode world response: %v", err)
	}

	worldObj, ok := worldResp["world"].(map[string]interface{})
	if !ok {
		t.Fatalf("world response missing world object: %v", worldResp)
	}

	worldID, ok := worldObj["id"].(string)
	if !ok {
		t.Fatalf("world response missing id: %v", worldObj)
	}

	return worldID
}

// setupTestStory creates a story and returns its ID
func setupTestStory(t *testing.T, db *postgres.DB, tenantID string) string {
	storyRepo := postgres.NewStoryRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	createStoryUseCase := story.NewCreateStoryUseCase(storyRepo, tenantRepo, worldRepo, createWorldUseCase, auditLogRepo, nil, log)
	getStoryUseCase := story.NewGetStoryUseCase(storyRepo, log)
	updateStoryUseCase := story.NewUpdateStoryUseCase(storyRepo, nil, log)
	listStoriesUseCase := story.NewListStoriesUseCase(storyRepo, log)
	chapterRepo := postgres.NewChapterRepository(db)
	sceneRepo := postgres.NewSceneRepository(db)
	beatRepo := postgres.NewBeatRepository(db)
	contentBlockRepo := postgres.NewContentBlockRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)
	cloneStoryUseCase := story.NewCloneStoryUseCase(storyRepo, chapterRepo, sceneRepo, beatRepo, contentBlockRepo, auditLogRepo, transactionRepo, log)
	storyHandler := NewStoryHandler(createStoryUseCase, getStoryUseCase, updateStoryUseCase, listStoriesUseCase, cloneStoryUseCase, log)

	storyBody := `{"title": "Test Story"}`
	storyReq := httptest.NewRequest("POST", "/api/v1/stories", strings.NewReader(storyBody))
	storyReq.Header.Set("Content-Type", "application/json")
	storyReq.Header.Set("X-Tenant-ID", tenantID)
	storyW := httptest.NewRecorder()
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

	return storyID
}

// setupTestChapter creates a chapter and returns its ID
func setupTestChapter(t *testing.T, db *postgres.DB, tenantID, storyID string) string {
	chapterRepo := postgres.NewChapterRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	log := logger.New()

	createChapterUseCase := chapterapp.NewCreateChapterUseCase(chapterRepo, storyRepo, nil, log)
	getChapterUseCase := chapterapp.NewGetChapterUseCase(chapterRepo, log)
	updateChapterUseCase := chapterapp.NewUpdateChapterUseCase(chapterRepo, nil, log)
	deleteChapterUseCase := chapterapp.NewDeleteChapterUseCase(chapterRepo, log)
	listChaptersUseCase := chapterapp.NewListChaptersUseCase(chapterRepo, log)
	chapterHandler := NewChapterHandler(createChapterUseCase, getChapterUseCase, updateChapterUseCase, deleteChapterUseCase, listChaptersUseCase, log)

	chapterBody := `{"story_id": "` + storyID + `", "number": 1, "title": "Test Chapter"}`
	chapterReq := httptest.NewRequest("POST", "/api/v1/chapters", strings.NewReader(chapterBody))
	chapterReq.Header.Set("Content-Type", "application/json")
	chapterReq.Header.Set("X-Tenant-ID", tenantID)
	chapterW := httptest.NewRecorder()
	withTenantMiddleware(chapterHandler.Create).ServeHTTP(chapterW, chapterReq)

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

	return chapterID
}

