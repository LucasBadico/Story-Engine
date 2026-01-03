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
	imageblockapp "github.com/story-engine/main-service/internal/application/story/image_block"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestImageBlockHandler_Create(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	imageBlockRepo := postgres.NewImageBlockRepository(db)
	imageBlockReferenceRepo := postgres.NewImageBlockReferenceRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	log := logger.New()

	createImageBlockUseCase := imageblockapp.NewCreateImageBlockUseCase(imageBlockRepo, chapterRepo, log)
	getImageBlockUseCase := imageblockapp.NewGetImageBlockUseCase(imageBlockRepo, log)
	listImageBlocksUseCase := imageblockapp.NewListImageBlocksUseCase(imageBlockRepo, log)
	updateImageBlockUseCase := imageblockapp.NewUpdateImageBlockUseCase(imageBlockRepo, log)
	deleteImageBlockUseCase := imageblockapp.NewDeleteImageBlockUseCase(imageBlockRepo, imageBlockReferenceRepo, log)
	addReferenceUseCase := imageblockapp.NewAddImageBlockReferenceUseCase(imageBlockRepo, imageBlockReferenceRepo, log)
	removeReferenceUseCase := imageblockapp.NewRemoveImageBlockReferenceUseCase(imageBlockReferenceRepo, log)
	getReferencesUseCase := imageblockapp.NewGetImageBlockReferencesUseCase(imageBlockReferenceRepo, log)
	handler := NewImageBlockHandler(createImageBlockUseCase, getImageBlockUseCase, listImageBlocksUseCase, updateImageBlockUseCase, deleteImageBlockUseCase, addReferenceUseCase, removeReferenceUseCase, getReferencesUseCase, log)

	t.Run("successful creation", func(t *testing.T) {
		body := `{"kind": "draft", "image_url": "https://example.com/image.jpg", "alt_text": "Test image"}`
		req := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/image-blocks", strings.NewReader(body))
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

		if imageBlock, ok := resp["image_block"].(map[string]interface{}); ok {
			if imageBlock["image_url"] != "https://example.com/image.jpg" {
				t.Errorf("expected image_url 'https://example.com/image.jpg', got %v", imageBlock["image_url"])
			}
		} else {
			t.Error("response missing image_block")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		body := `{"kind": "draft", "image_url": "https://example.com/image.jpg"}`
		req := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/image-blocks", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("id", chapterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Create).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestImageBlockHandler_Get(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	imageBlockRepo := postgres.NewImageBlockRepository(db)
	imageBlockReferenceRepo := postgres.NewImageBlockReferenceRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	log := logger.New()

	createImageBlockUseCase := imageblockapp.NewCreateImageBlockUseCase(imageBlockRepo, chapterRepo, log)
	getImageBlockUseCase := imageblockapp.NewGetImageBlockUseCase(imageBlockRepo, log)
	listImageBlocksUseCase := imageblockapp.NewListImageBlocksUseCase(imageBlockRepo, log)
	updateImageBlockUseCase := imageblockapp.NewUpdateImageBlockUseCase(imageBlockRepo, log)
	deleteImageBlockUseCase := imageblockapp.NewDeleteImageBlockUseCase(imageBlockRepo, imageBlockReferenceRepo, log)
	addReferenceUseCase := imageblockapp.NewAddImageBlockReferenceUseCase(imageBlockRepo, imageBlockReferenceRepo, log)
	removeReferenceUseCase := imageblockapp.NewRemoveImageBlockReferenceUseCase(imageBlockReferenceRepo, log)
	getReferencesUseCase := imageblockapp.NewGetImageBlockReferencesUseCase(imageBlockReferenceRepo, log)
	handler := NewImageBlockHandler(createImageBlockUseCase, getImageBlockUseCase, listImageBlocksUseCase, updateImageBlockUseCase, deleteImageBlockUseCase, addReferenceUseCase, removeReferenceUseCase, getReferencesUseCase, log)

	// Create image block
	imageBlockBody := `{"kind": "draft", "image_url": "https://example.com/get-test.jpg"}`
	imageBlockReq := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/image-blocks", strings.NewReader(imageBlockBody))
	imageBlockReq.Header.Set("Content-Type", "application/json")
	imageBlockReq.Header.Set("X-Tenant-ID", tenantID)
	imageBlockReq.SetPathValue("id", chapterID)
	imageBlockW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(imageBlockW, imageBlockReq)

	if imageBlockW.Code != http.StatusCreated {
		t.Fatalf("failed to create image block: status %d, body: %s", imageBlockW.Code, imageBlockW.Body.String())
	}

	var imageBlockResp map[string]interface{}
	if err := json.NewDecoder(imageBlockW.Body).Decode(&imageBlockResp); err != nil {
		t.Fatalf("failed to decode image block response: %v", err)
	}

	imageBlockObj, ok := imageBlockResp["image_block"].(map[string]interface{})
	if !ok {
		t.Fatalf("image block response missing image_block object: %v", imageBlockResp)
	}

	imageBlockID, ok := imageBlockObj["id"].(string)
	if !ok {
		t.Fatalf("image block response missing id: %v", imageBlockObj)
	}

	t.Run("existing image_block", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/image-blocks/"+imageBlockID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", imageBlockID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if imageBlock, ok := resp["image_block"].(map[string]interface{}); ok {
			if imageBlock["id"] != imageBlockID {
				t.Errorf("expected ID %s, got %v", imageBlockID, imageBlock["id"])
			}
		} else {
			t.Error("response missing image_block")
		}
	})

	t.Run("non-existing image_block", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/image-blocks/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

func TestImageBlockHandler_List(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	imageBlockRepo := postgres.NewImageBlockRepository(db)
	imageBlockReferenceRepo := postgres.NewImageBlockReferenceRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	log := logger.New()

	createImageBlockUseCase := imageblockapp.NewCreateImageBlockUseCase(imageBlockRepo, chapterRepo, log)
	getImageBlockUseCase := imageblockapp.NewGetImageBlockUseCase(imageBlockRepo, log)
	listImageBlocksUseCase := imageblockapp.NewListImageBlocksUseCase(imageBlockRepo, log)
	updateImageBlockUseCase := imageblockapp.NewUpdateImageBlockUseCase(imageBlockRepo, log)
	deleteImageBlockUseCase := imageblockapp.NewDeleteImageBlockUseCase(imageBlockRepo, imageBlockReferenceRepo, log)
	addReferenceUseCase := imageblockapp.NewAddImageBlockReferenceUseCase(imageBlockRepo, imageBlockReferenceRepo, log)
	removeReferenceUseCase := imageblockapp.NewRemoveImageBlockReferenceUseCase(imageBlockReferenceRepo, log)
	getReferencesUseCase := imageblockapp.NewGetImageBlockReferencesUseCase(imageBlockReferenceRepo, log)
	handler := NewImageBlockHandler(createImageBlockUseCase, getImageBlockUseCase, listImageBlocksUseCase, updateImageBlockUseCase, deleteImageBlockUseCase, addReferenceUseCase, removeReferenceUseCase, getReferencesUseCase, log)

	// Create multiple image blocks
	for i := 1; i <= 3; i++ {
		imageBlockBody := `{"kind": "draft", "image_url": "https://example.com/image` + strconv.Itoa(i) + `.jpg"}`
		imageBlockReq := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/image-blocks", strings.NewReader(imageBlockBody))
		imageBlockReq.Header.Set("Content-Type", "application/json")
		imageBlockReq.Header.Set("X-Tenant-ID", tenantID)
		imageBlockReq.SetPathValue("id", chapterID)
		imageBlockW := httptest.NewRecorder()
		withTenantMiddleware(handler.Create).ServeHTTP(imageBlockW, imageBlockReq)

		if imageBlockW.Code != http.StatusCreated {
			t.Fatalf("failed to create image block %d: status %d, body: %s", i, imageBlockW.Code, imageBlockW.Body.String())
		}
	}

	t.Run("list image_blocks", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/chapters/"+chapterID+"/image-blocks", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", chapterID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.List).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if imageBlocks, ok := resp["image_blocks"].([]interface{}); ok {
			if len(imageBlocks) < 3 {
				t.Errorf("expected at least 3 image blocks, got %d", len(imageBlocks))
			}
		} else {
			t.Error("response missing image_blocks")
		}
	})
}

func TestImageBlockHandler_Update(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	imageBlockRepo := postgres.NewImageBlockRepository(db)
	imageBlockReferenceRepo := postgres.NewImageBlockReferenceRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	log := logger.New()

	createImageBlockUseCase := imageblockapp.NewCreateImageBlockUseCase(imageBlockRepo, chapterRepo, log)
	getImageBlockUseCase := imageblockapp.NewGetImageBlockUseCase(imageBlockRepo, log)
	listImageBlocksUseCase := imageblockapp.NewListImageBlocksUseCase(imageBlockRepo, log)
	updateImageBlockUseCase := imageblockapp.NewUpdateImageBlockUseCase(imageBlockRepo, log)
	deleteImageBlockUseCase := imageblockapp.NewDeleteImageBlockUseCase(imageBlockRepo, imageBlockReferenceRepo, log)
	addReferenceUseCase := imageblockapp.NewAddImageBlockReferenceUseCase(imageBlockRepo, imageBlockReferenceRepo, log)
	removeReferenceUseCase := imageblockapp.NewRemoveImageBlockReferenceUseCase(imageBlockReferenceRepo, log)
	getReferencesUseCase := imageblockapp.NewGetImageBlockReferencesUseCase(imageBlockReferenceRepo, log)
	handler := NewImageBlockHandler(createImageBlockUseCase, getImageBlockUseCase, listImageBlocksUseCase, updateImageBlockUseCase, deleteImageBlockUseCase, addReferenceUseCase, removeReferenceUseCase, getReferencesUseCase, log)

	// Create image block
	imageBlockBody := `{"kind": "draft", "image_url": "https://example.com/original.jpg"}`
	imageBlockReq := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/image-blocks", strings.NewReader(imageBlockBody))
	imageBlockReq.Header.Set("Content-Type", "application/json")
	imageBlockReq.Header.Set("X-Tenant-ID", tenantID)
	imageBlockReq.SetPathValue("id", chapterID)
	imageBlockW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(imageBlockW, imageBlockReq)

	if imageBlockW.Code != http.StatusCreated {
		t.Fatalf("failed to create image block: status %d, body: %s", imageBlockW.Code, imageBlockW.Body.String())
	}

	var imageBlockResp map[string]interface{}
	if err := json.NewDecoder(imageBlockW.Body).Decode(&imageBlockResp); err != nil {
		t.Fatalf("failed to decode image block response: %v", err)
	}

	imageBlockObj, ok := imageBlockResp["image_block"].(map[string]interface{})
	if !ok {
		t.Fatalf("image block response missing image_block object: %v", imageBlockResp)
	}

	imageBlockID, ok := imageBlockObj["id"].(string)
	if !ok {
		t.Fatalf("image block response missing id: %v", imageBlockObj)
	}

	t.Run("update image_block url", func(t *testing.T) {
		body := `{"image_url": "https://example.com/updated.jpg"}`
		req := httptest.NewRequest("PUT", "/api/v1/image-blocks/"+imageBlockID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", imageBlockID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Update).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if imageBlock, ok := resp["image_block"].(map[string]interface{}); ok {
			if imageBlock["image_url"] != "https://example.com/updated.jpg" {
				t.Errorf("expected image_url 'https://example.com/updated.jpg', got %v", imageBlock["image_url"])
			}
		} else {
			t.Error("response missing image_block")
		}
	})

	t.Run("non-existing image_block", func(t *testing.T) {
		body := `{"image_url": "https://example.com/non-existent.jpg"}`
		req := httptest.NewRequest("PUT", "/api/v1/image-blocks/00000000-0000-0000-0000-000000000000", strings.NewReader(body))
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

func TestImageBlockHandler_Delete(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	tenantID := setupTestTenant(t, db)
	storyID := setupTestStory(t, db, tenantID)
	chapterID := setupTestChapter(t, db, tenantID, storyID)
	imageBlockRepo := postgres.NewImageBlockRepository(db)
	imageBlockReferenceRepo := postgres.NewImageBlockReferenceRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	log := logger.New()

	createImageBlockUseCase := imageblockapp.NewCreateImageBlockUseCase(imageBlockRepo, chapterRepo, log)
	getImageBlockUseCase := imageblockapp.NewGetImageBlockUseCase(imageBlockRepo, log)
	listImageBlocksUseCase := imageblockapp.NewListImageBlocksUseCase(imageBlockRepo, log)
	updateImageBlockUseCase := imageblockapp.NewUpdateImageBlockUseCase(imageBlockRepo, log)
	deleteImageBlockUseCase := imageblockapp.NewDeleteImageBlockUseCase(imageBlockRepo, imageBlockReferenceRepo, log)
	addReferenceUseCase := imageblockapp.NewAddImageBlockReferenceUseCase(imageBlockRepo, imageBlockReferenceRepo, log)
	removeReferenceUseCase := imageblockapp.NewRemoveImageBlockReferenceUseCase(imageBlockReferenceRepo, log)
	getReferencesUseCase := imageblockapp.NewGetImageBlockReferencesUseCase(imageBlockReferenceRepo, log)
	handler := NewImageBlockHandler(createImageBlockUseCase, getImageBlockUseCase, listImageBlocksUseCase, updateImageBlockUseCase, deleteImageBlockUseCase, addReferenceUseCase, removeReferenceUseCase, getReferencesUseCase, log)

	// Create image block
	imageBlockBody := `{"kind": "draft", "image_url": "https://example.com/delete.jpg"}`
	imageBlockReq := httptest.NewRequest("POST", "/api/v1/chapters/"+chapterID+"/image-blocks", strings.NewReader(imageBlockBody))
	imageBlockReq.Header.Set("Content-Type", "application/json")
	imageBlockReq.Header.Set("X-Tenant-ID", tenantID)
	imageBlockReq.SetPathValue("id", chapterID)
	imageBlockW := httptest.NewRecorder()
	withTenantMiddleware(handler.Create).ServeHTTP(imageBlockW, imageBlockReq)

	if imageBlockW.Code != http.StatusCreated {
		t.Fatalf("failed to create image block: status %d, body: %s", imageBlockW.Code, imageBlockW.Body.String())
	}

	var imageBlockResp map[string]interface{}
	if err := json.NewDecoder(imageBlockW.Body).Decode(&imageBlockResp); err != nil {
		t.Fatalf("failed to decode image block response: %v", err)
	}

	imageBlockObj, ok := imageBlockResp["image_block"].(map[string]interface{})
	if !ok {
		t.Fatalf("image block response missing image_block object: %v", imageBlockResp)
	}

	imageBlockID, ok := imageBlockObj["id"].(string)
	if !ok {
		t.Fatalf("image block response missing id: %v", imageBlockObj)
	}

	t.Run("delete existing image_block", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/image-blocks/"+imageBlockID, nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", imageBlockID)
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}

		// Verify image block is deleted
		getReq := httptest.NewRequest("GET", "/api/v1/image-blocks/"+imageBlockID, nil)
		getReq.Header.Set("X-Tenant-ID", tenantID)
		getReq.SetPathValue("id", imageBlockID)
		getW := httptest.NewRecorder()

		withTenantMiddleware(handler.Get).ServeHTTP(getW, getReq)

		if getW.Code != http.StatusNotFound {
			t.Errorf("expected status 404 when getting deleted image block, got %d", getW.Code)
		}
	})

	t.Run("delete non-existing image_block", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/image-blocks/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("X-Tenant-ID", tenantID)
		req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
		w := httptest.NewRecorder()

		withTenantMiddleware(handler.Delete).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}
