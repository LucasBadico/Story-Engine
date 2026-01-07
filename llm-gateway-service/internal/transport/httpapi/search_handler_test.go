package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/application/search"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

func TestSearchHandler_Search_Success(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	documentID := uuid.New()

	mockChunkRepo := repositories.NewMockChunkRepository()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	doc := memory.NewDocument(tenantID, memory.SourceTypeStory, uuid.New(), "Test", "Content")
	doc.ID = documentID
	if err := mockDocRepo.Create(ctx, doc); err != nil {
		t.Fatalf("failed to create document: %v", err)
	}

	chunk := memory.NewChunk(documentID, 0, "Test content", []float32{0.1, 0.2, 0.3}, 10)
	if err := mockChunkRepo.Create(ctx, chunk); err != nil {
		t.Fatalf("failed to create chunk: %v", err)
	}

	useCase := search.NewSearchMemoryUseCase(mockChunkRepo, mockDocRepo, mockEmbedder, log)
	handler := NewSearchHandler(useCase, log)

	body, _ := json.Marshal(map[string]interface{}{
		"query": "test query",
		"limit": 5,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	rr := httptest.NewRecorder()

	handler.Search(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Chunks     []map[string]interface{} `json:"chunks"`
		NextCursor string                   `json:"next_cursor"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(resp.Chunks))
	}
	if resp.NextCursor == "" {
		t.Error("expected next_cursor to be set")
	}
}

func TestSearchHandler_Search_MissingTenant(t *testing.T) {
	useCase := search.NewSearchMemoryUseCase(
		repositories.NewMockChunkRepository(),
		repositories.NewMockDocumentRepository(),
		embeddings.NewMockEmbedder(768),
		logger.New(),
	)
	handler := NewSearchHandler(useCase, logger.New())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", bytes.NewBufferString(`{"query":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Search(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestSearchHandler_Search_InvalidCursor(t *testing.T) {
	useCase := search.NewSearchMemoryUseCase(
		repositories.NewMockChunkRepository(),
		repositories.NewMockDocumentRepository(),
		embeddings.NewMockEmbedder(768),
		logger.New(),
	)
	handler := NewSearchHandler(useCase, logger.New())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", bytes.NewBufferString(`{"query":"test","cursor":"invalid"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", uuid.New().String())
	rr := httptest.NewRecorder()

	handler.Search(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}
