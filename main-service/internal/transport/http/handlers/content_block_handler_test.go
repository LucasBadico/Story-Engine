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
	contentblockapp "github.com/story-engine/main-service/internal/application/story/content_block"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestContentBlockHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	contentBlockRepo := postgres.NewContentBlockRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	log := logger.New()

	createContentBlockUseCase := contentblockapp.NewCreateContentBlockUseCase(contentBlockRepo, chapterRepo, log)
	getContentBlockUseCase := contentblockapp.NewGetContentBlockUseCase(contentBlockRepo, log)
	updateContentBlockUseCase := contentblockapp.NewUpdateContentBlockUseCase(contentBlockRepo, log)
	deleteContentBlockUseCase := contentblockapp.NewDeleteContentBlockUseCase(contentBlockRepo, log)
	listContentBlocksUseCase := contentblockapp.NewListContentBlocksUseCase(contentBlockRepo, log)
	handler := NewContentBlockHandler(createContentBlockUseCase, getContentBlockUseCase, updateContentBlockUseCase, deleteContentBlockUseCase, listContentBlocksUseCase, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"type": "text", "kind": "draft", "content": "Test content", "order_num": 1}`
		req := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/content-blocks", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", chapterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if contentBlock, ok := resp["content_block"].(map[string]interface{}); ok {
			if contentBlock["content"] != "Test content" {
				t.Errorf("expected content 'Test content', got %v", contentBlock["content"])
			}
		} else {
			t.Error("response missing content_block")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"type": "text", "kind": "draft", "content": "Test content", "order_num": 1}`
		req := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/content-blocks", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("id", chapterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestContentBlockHandler_Get(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	contentBlockRepo := postgres.NewContentBlockRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	log := logger.New()

	createContentBlockUseCase := contentblockapp.NewCreateContentBlockUseCase(contentBlockRepo, chapterRepo, log)
	getContentBlockUseCase := contentblockapp.NewGetContentBlockUseCase(contentBlockRepo, log)
	updateContentBlockUseCase := contentblockapp.NewUpdateContentBlockUseCase(contentBlockRepo, log)
	deleteContentBlockUseCase := contentblockapp.NewDeleteContentBlockUseCase(contentBlockRepo, log)
	listContentBlocksUseCase := contentblockapp.NewListContentBlocksUseCase(contentBlockRepo, log)
	handler := NewContentBlockHandler(createContentBlockUseCase, getContentBlockUseCase, updateContentBlockUseCase, deleteContentBlockUseCase, listContentBlocksUseCase, log)

	// Create content block
	contentBlockBody := `{"type": "text", "kind": "draft", "content": "Get Test Content", "order_num": 1}`
	contentBlockReq := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/content-blocks", strings.NewReader(contentBlockBody))
	contentBlockReq.Header.Set("Content-Type", "application/json")
	contentBlockReq.Header.Set("X-Tenant-ID", tenantID)
	contentBlockReq.SetPathValue("id", chapterID)
	contentBlockW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(contentBlockW, contentBlockReq)

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

	t.Run("existing content_block", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/content-blocks/"+contentBlockID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", contentBlockID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if contentBlock, ok := resp["content_block"].(map[string]interface{}); ok {
			if contentBlock["id"] != contentBlockID {
				t.Errorf("expected ID %s, got %v", contentBlockID, contentBlock["id"])
			}
		} else {
			t.Error("response missing content_block")
		}
	})

	t.Run("non-existing content_block", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/content-blocks/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

func TestContentBlockHandler_List(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	contentBlockRepo := postgres.NewContentBlockRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	log := logger.New()

	createContentBlockUseCase := contentblockapp.NewCreateContentBlockUseCase(contentBlockRepo, chapterRepo, log)
	getContentBlockUseCase := contentblockapp.NewGetContentBlockUseCase(contentBlockRepo, log)
	updateContentBlockUseCase := contentblockapp.NewUpdateContentBlockUseCase(contentBlockRepo, log)
	deleteContentBlockUseCase := contentblockapp.NewDeleteContentBlockUseCase(contentBlockRepo, log)
	listContentBlocksUseCase := contentblockapp.NewListContentBlocksUseCase(contentBlockRepo, log)
	handler := NewContentBlockHandler(createContentBlockUseCase, getContentBlockUseCase, updateContentBlockUseCase, deleteContentBlockUseCase, listContentBlocksUseCase, log)

	// Create multiple content blocks
	for i := 1; i <= 3; i++ {
		contentBlockBody := `{"type": "text", "kind": "draft", "content": "Content ` + strconv.Itoa(i) + `", "order_num": ` + strconv.Itoa(i) + `}`
		contentBlockReq := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/content-blocks", strings.NewReader(contentBlockBody))
		contentBlockReq.Header.Set("Content-Type", "application/json")
		contentBlockReq.Header.Set("X-Tenant-ID", tenantID)
		contentBlockReq.SetPathValue("id", chapterID)
		contentBlockW := httptest.NewRecorder()
		withTenantMiddleware(handler.Create).ServeHTTP(contentBlockW, contentBlockReq)

		if contentBlockW.Code != http.StatusCreated {
			t.Fatalf("failed to create content block %d: status %d, body: %s", i, contentBlockW.Code, contentBlockW.Body.String())
		}
	}

	t.Run("list content_blocks", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/chapters/"+chapterID+"/content-blocks", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", chapterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.ListByChapter).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if contentBlocks, ok := resp["content_blocks"].([]interface{}); ok {
			if len(contentBlocks) < 3 {
				t.Errorf("expected at least 3 content blocks, got %d", len(contentBlocks))
			}
		} else {
			t.Error("response missing content_blocks")
		}
	})
}

func TestContentBlockHandler_Update(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	contentBlockRepo := postgres.NewContentBlockRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	log := logger.New()

	createContentBlockUseCase := contentblockapp.NewCreateContentBlockUseCase(contentBlockRepo, chapterRepo, log)
	getContentBlockUseCase := contentblockapp.NewGetContentBlockUseCase(contentBlockRepo, log)
	updateContentBlockUseCase := contentblockapp.NewUpdateContentBlockUseCase(contentBlockRepo, log)
	deleteContentBlockUseCase := contentblockapp.NewDeleteContentBlockUseCase(contentBlockRepo, log)
	listContentBlocksUseCase := contentblockapp.NewListContentBlocksUseCase(contentBlockRepo, log)
	handler := NewContentBlockHandler(createContentBlockUseCase, getContentBlockUseCase, updateContentBlockUseCase, deleteContentBlockUseCase, listContentBlocksUseCase, log)

	// Create content block
	contentBlockBody := `{"type": "text", "kind": "draft", "content": "Original Content", "order_num": 1}`
	contentBlockReq := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/content-blocks", strings.NewReader(contentBlockBody))
	contentBlockReq.Header.Set("Content-Type", "application/json")
	contentBlockReq.Header.Set("X-Tenant-ID", tenantID)
	contentBlockReq.SetPathValue("id", chapterID)
	contentBlockW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(contentBlockW, contentBlockReq)

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

	t.Run("update content_block content", func(t *testing.T) {
		body := `{"content": "Updated Content"}`
		req := httptest.NewRequest("PUT", "/api/v1/content-blocks/"+contentBlockID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", contentBlockID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if contentBlock, ok := resp["content_block"].(map[string]interface{}); ok {
			if contentBlock["content"] != "Updated Content" {
				t.Errorf("expected content 'Updated Content', got %v", contentBlock["content"])
			}
		} else {
			t.Error("response missing content_block")
		}
	})

	t.Run("non-existing content_block", func(t *testing.T) {
		body := `{"content": "Non-existent"}`
		req := httptest.NewRequest("PUT", "/api/v1/content-blocks/00000000-0000-0000-0000-000000000000", strings.NewReader(body))
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

func TestContentBlockHandler_Delete(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	contentBlockRepo := postgres.NewContentBlockRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	log := logger.New()

	createContentBlockUseCase := contentblockapp.NewCreateContentBlockUseCase(contentBlockRepo, chapterRepo, log)
	getContentBlockUseCase := contentblockapp.NewGetContentBlockUseCase(contentBlockRepo, log)
	updateContentBlockUseCase := contentblockapp.NewUpdateContentBlockUseCase(contentBlockRepo, log)
	deleteContentBlockUseCase := contentblockapp.NewDeleteContentBlockUseCase(contentBlockRepo, log)
	listContentBlocksUseCase := contentblockapp.NewListContentBlocksUseCase(contentBlockRepo, log)
	handler := NewContentBlockHandler(createContentBlockUseCase, getContentBlockUseCase, updateContentBlockUseCase, deleteContentBlockUseCase, listContentBlocksUseCase, log)

	// Create content block
	contentBlockBody := `{"type": "text", "kind": "draft", "content": "Content to Delete", "order_num": 1}`
	contentBlockReq := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/content-blocks", strings.NewReader(contentBlockBody))
	contentBlockReq.Header.Set("Content-Type", "application/json")
	contentBlockReq.Header.Set("X-Tenant-ID", tenantID)
	contentBlockReq.SetPathValue("id", chapterID)
	contentBlockW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(contentBlockW, contentBlockReq)

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

	t.Run("delete existing content_block", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/content-blocks/"+contentBlockID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", contentBlockID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}

		// Verify content block is deleted
		getReq := httptest.NewRequest("GET", "/api/v1/content-blocks/"+contentBlockID, nil)
		getReq.Header.Set("X-Tenant-ID", tenantID)
		getReq.SetPathValue("id", contentBlockID)
		getW := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(getW, getReq)

		if getW.Code != http.StatusNotFound {
			t.Errorf("expected status 404 when getting deleted content block, got %d", getW.Code)
		}
	})

	t.Run("delete non-existing content_block", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/content-blocks/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}
