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
	proseblockapp "github.com/story-engine/main-service/internal/application/story/prose_block"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestProseBlockHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	proseBlockRepo := postgres.NewProseBlockRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	log := logger.New()

	createProseBlockUseCase := proseblockapp.NewCreateProseBlockUseCase(proseBlockRepo, chapterRepo, log)
	getProseBlockUseCase := proseblockapp.NewGetProseBlockUseCase(proseBlockRepo, log)
	updateProseBlockUseCase := proseblockapp.NewUpdateProseBlockUseCase(proseBlockRepo, log)
	deleteProseBlockUseCase := proseblockapp.NewDeleteProseBlockUseCase(proseBlockRepo, log)
	listProseBlocksUseCase := proseblockapp.NewListProseBlocksUseCase(proseBlockRepo, log)
	handler := NewProseBlockHandler(createProseBlockUseCase, getProseBlockUseCase, updateProseBlockUseCase, deleteProseBlockUseCase, listProseBlocksUseCase, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"kind": "draft", "content": "Test content", "order_num": 1}`
		req := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/prose-blocks", strings.NewReader(body))
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

		if proseBlock, ok := resp["prose_block"].(map[string]interface{}); ok {
			if proseBlock["content"] != "Test content" {
				t.Errorf("expected content 'Test content', got %v", proseBlock["content"])
			}
		} else {
			t.Error("response missing prose_block")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"kind": "draft", "content": "Test content", "order_num": 1}`
		req := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/prose-blocks", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("id", chapterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestProseBlockHandler_Get(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	proseBlockRepo := postgres.NewProseBlockRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	log := logger.New()

	createProseBlockUseCase := proseblockapp.NewCreateProseBlockUseCase(proseBlockRepo, chapterRepo, log)
	getProseBlockUseCase := proseblockapp.NewGetProseBlockUseCase(proseBlockRepo, log)
	updateProseBlockUseCase := proseblockapp.NewUpdateProseBlockUseCase(proseBlockRepo, log)
	deleteProseBlockUseCase := proseblockapp.NewDeleteProseBlockUseCase(proseBlockRepo, log)
	listProseBlocksUseCase := proseblockapp.NewListProseBlocksUseCase(proseBlockRepo, log)
	handler := NewProseBlockHandler(createProseBlockUseCase, getProseBlockUseCase, updateProseBlockUseCase, deleteProseBlockUseCase, listProseBlocksUseCase, log)

	// Create prose block
	proseBlockBody := `{"kind": "draft", "content": "Get Test Content", "order_num": 1}`
	proseBlockReq := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/prose-blocks", strings.NewReader(proseBlockBody))
	proseBlockReq.Header.Set("Content-Type", "application/json")
	proseBlockReq.Header.Set("X-Tenant-ID", tenantID)
	proseBlockReq.SetPathValue("id", chapterID)
	proseBlockW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(proseBlockW, proseBlockReq)

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

	t.Run("existing prose_block", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/prose-blocks/"+proseBlockID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", proseBlockID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if proseBlock, ok := resp["prose_block"].(map[string]interface{}); ok {
			if proseBlock["id"] != proseBlockID {
				t.Errorf("expected ID %s, got %v", proseBlockID, proseBlock["id"])
			}
		} else {
			t.Error("response missing prose_block")
		}
	})

	t.Run("non-existing prose_block", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/prose-blocks/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

func TestProseBlockHandler_List(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	proseBlockRepo := postgres.NewProseBlockRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	log := logger.New()

	createProseBlockUseCase := proseblockapp.NewCreateProseBlockUseCase(proseBlockRepo, chapterRepo, log)
	getProseBlockUseCase := proseblockapp.NewGetProseBlockUseCase(proseBlockRepo, log)
	updateProseBlockUseCase := proseblockapp.NewUpdateProseBlockUseCase(proseBlockRepo, log)
	deleteProseBlockUseCase := proseblockapp.NewDeleteProseBlockUseCase(proseBlockRepo, log)
	listProseBlocksUseCase := proseblockapp.NewListProseBlocksUseCase(proseBlockRepo, log)
	handler := NewProseBlockHandler(createProseBlockUseCase, getProseBlockUseCase, updateProseBlockUseCase, deleteProseBlockUseCase, listProseBlocksUseCase, log)

	// Create multiple prose blocks
	for i := 1; i <= 3; i++ {
		proseBlockBody := `{"kind": "draft", "content": "Content ` + strconv.Itoa(i) + `", "order_num": ` + strconv.Itoa(i) + `}`
		proseBlockReq := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/prose-blocks", strings.NewReader(proseBlockBody))
		proseBlockReq.Header.Set("Content-Type", "application/json")
		proseBlockReq.Header.Set("X-Tenant-ID", tenantID)
		proseBlockReq.SetPathValue("id", chapterID)
		proseBlockW := httptest.NewRecorder()
		withTenantMiddleware(handler.Create).ServeHTTP(proseBlockW, proseBlockReq)

		if proseBlockW.Code != http.StatusCreated {
			t.Fatalf("failed to create prose block %d: status %d, body: %s", i, proseBlockW.Code, proseBlockW.Body.String())
		}
	}

	t.Run("list prose_blocks", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/chapters/"+chapterID+"/prose-blocks", nil)
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

		if proseBlocks, ok := resp["prose_blocks"].([]interface{}); ok {
			if len(proseBlocks) < 3 {
				t.Errorf("expected at least 3 prose blocks, got %d", len(proseBlocks))
			}
		} else {
			t.Error("response missing prose_blocks")
		}
	})
}

func TestProseBlockHandler_Update(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	proseBlockRepo := postgres.NewProseBlockRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	log := logger.New()

	createProseBlockUseCase := proseblockapp.NewCreateProseBlockUseCase(proseBlockRepo, chapterRepo, log)
	getProseBlockUseCase := proseblockapp.NewGetProseBlockUseCase(proseBlockRepo, log)
	updateProseBlockUseCase := proseblockapp.NewUpdateProseBlockUseCase(proseBlockRepo, log)
	deleteProseBlockUseCase := proseblockapp.NewDeleteProseBlockUseCase(proseBlockRepo, log)
	listProseBlocksUseCase := proseblockapp.NewListProseBlocksUseCase(proseBlockRepo, log)
	handler := NewProseBlockHandler(createProseBlockUseCase, getProseBlockUseCase, updateProseBlockUseCase, deleteProseBlockUseCase, listProseBlocksUseCase, log)

	// Create prose block
	proseBlockBody := `{"kind": "draft", "content": "Original Content", "order_num": 1}`
	proseBlockReq := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/prose-blocks", strings.NewReader(proseBlockBody))
	proseBlockReq.Header.Set("Content-Type", "application/json")
	proseBlockReq.Header.Set("X-Tenant-ID", tenantID)
	proseBlockReq.SetPathValue("id", chapterID)
	proseBlockW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(proseBlockW, proseBlockReq)

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

	t.Run("update prose_block content", func(t *testing.T) {
		body := `{"content": "Updated Content"}`
		req := httptest.NewRequest("PUT", "/api/v1/prose-blocks/"+proseBlockID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", proseBlockID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if proseBlock, ok := resp["prose_block"].(map[string]interface{}); ok {
			if proseBlock["content"] != "Updated Content" {
				t.Errorf("expected content 'Updated Content', got %v", proseBlock["content"])
			}
		} else {
			t.Error("response missing prose_block")
		}
	})

	t.Run("non-existing prose_block", func(t *testing.T) {
		body := `{"content": "Non-existent"}`
		req := httptest.NewRequest("PUT", "/api/v1/prose-blocks/00000000-0000-0000-0000-000000000000", strings.NewReader(body))
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

func TestProseBlockHandler_Delete(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	proseBlockRepo := postgres.NewProseBlockRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	log := logger.New()

	createProseBlockUseCase := proseblockapp.NewCreateProseBlockUseCase(proseBlockRepo, chapterRepo, log)
	getProseBlockUseCase := proseblockapp.NewGetProseBlockUseCase(proseBlockRepo, log)
	updateProseBlockUseCase := proseblockapp.NewUpdateProseBlockUseCase(proseBlockRepo, log)
	deleteProseBlockUseCase := proseblockapp.NewDeleteProseBlockUseCase(proseBlockRepo, log)
	listProseBlocksUseCase := proseblockapp.NewListProseBlocksUseCase(proseBlockRepo, log)
	handler := NewProseBlockHandler(createProseBlockUseCase, getProseBlockUseCase, updateProseBlockUseCase, deleteProseBlockUseCase, listProseBlocksUseCase, log)

	// Create prose block
	proseBlockBody := `{"kind": "draft", "content": "Content to Delete", "order_num": 1}`
	proseBlockReq := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/prose-blocks", strings.NewReader(proseBlockBody))
	proseBlockReq.Header.Set("Content-Type", "application/json")
	proseBlockReq.Header.Set("X-Tenant-ID", tenantID)
	proseBlockReq.SetPathValue("id", chapterID)
	proseBlockW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(proseBlockW, proseBlockReq)

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

	t.Run("delete existing prose_block", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/prose-blocks/"+proseBlockID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", proseBlockID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}

		// Verify prose block is deleted
		getReq := httptest.NewRequest("GET", "/api/v1/prose-blocks/"+proseBlockID, nil)
		getReq.Header.Set("X-Tenant-ID", tenantID)
		getReq.SetPathValue("id", proseBlockID)
		getW := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(getW, getReq)

		if getW.Code != http.StatusNotFound {
			t.Errorf("expected status 404 when getting deleted prose block, got %d", getW.Code)
		}
	})

	t.Run("delete non-existing prose_block", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/prose-blocks/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}
