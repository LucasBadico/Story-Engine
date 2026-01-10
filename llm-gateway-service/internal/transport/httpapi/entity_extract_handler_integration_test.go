//go:build integration

package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/adapters/llm/gemini"
	"github.com/story-engine/llm-gateway-service/internal/application/entity_extraction"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

func TestEntityExtractHandler_Integration(t *testing.T) {
	if strings.TrimSpace(os.Getenv("LLM_TESTS_ENABLED")) == "" {
		t.Skip("LLM_TESTS_ENABLED not set; skipping LLM integration tests")
	}

	apiKey := loadGeminiAPIKey(t)
	if apiKey == "" {
		t.Fatalf("gemini api key not configured")
	}

	modelName := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))

	ctx := context.Background()
	tenantID := uuid.New()
	worldID := uuid.New()

	docRepo := repositories.NewMockDocumentRepository()
	chunkRepo := repositories.NewMockChunkRepository()
	embedder := embeddings.NewMockEmbedder(3)
	log := logger.New()

	knownSourceID := uuid.New()
	knownDocID := uuid.New()
	createFixtureEntity(t, ctx, docRepo, chunkRepo, tenantID, worldID, knownSourceID, knownDocID, "Aria", "Aria is a mage.")

	model := gemini.NewRouterModel(apiKey, modelName)
	router := entity_extraction.NewPhase1EntityTypeRouterUseCase(model, log)
	extractor := entity_extraction.NewPhase2EntryUseCase(model, log, nil)
	matcher := entity_extraction.NewPhase3MatchUseCase(chunkRepo, docRepo, embedder, model, log)
	payload := entity_extraction.NewPhase4EntitiesPayloadUseCase()
	useCase := entity_extraction.NewEntityAndRelationshipsExtractor(router, extractor, matcher, payload, log)
	handler := NewEntityExtractHandler(useCase, log)

	timeoutCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	body, _ := json.Marshal(map[string]interface{}{
		"text":     "Aria stepped into the tower. \n\n Luna arrived later.",
		"world_id": worldID.String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/entity-extract", bytes.NewBuffer(body))
	req = req.WithContext(timeoutCtx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	rr := httptest.NewRecorder()

	handler.Extract(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Entities []struct {
			Type  string `json:"type"`
			Name  string `json:"name"`
			Found bool   `json:"found"`
			Match *struct {
				SourceID uuid.UUID `json:"source_id"`
			} `json:"match"`
		} `json:"entities"`
		Relations []interface{} `json:"relations"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(resp.Entities) == 0 {
		t.Fatalf("expected entities, got 0")
	}
	if len(resp.Entities) < 2 {
		t.Fatalf("expected at least 2 entities, got %d", len(resp.Entities))
	}

	var ariaFound bool
	for _, entity := range resp.Entities {
		if entity.Name == "Aria" {
			ariaFound = true
			if !entity.Found || entity.Match == nil {
				t.Fatalf("expected match for Aria")
			}
			if entity.Match.SourceID != knownSourceID {
				t.Fatalf("expected source id %s, got %s", knownSourceID, entity.Match.SourceID)
			}
		}
	}
	if !ariaFound {
		t.Fatalf("expected Aria in response")
	}
}

func createFixtureEntity(
	t *testing.T,
	ctx context.Context,
	docRepo repositories.DocumentRepository,
	chunkRepo repositories.ChunkRepository,
	tenantID uuid.UUID,
	worldID uuid.UUID,
	sourceID uuid.UUID,
	docID uuid.UUID,
	name string,
	summary string,
) {
	t.Helper()

	doc := memory.NewDocument(tenantID, memory.SourceTypeCharacter, sourceID, name, summary)
	doc.ID = docID
	if err := docRepo.Create(ctx, doc); err != nil {
		t.Fatalf("create doc: %v", err)
	}

	chunkType := "summary"
	entityName := name
	chunk := &memory.Chunk{
		ID:         uuid.New(),
		DocumentID: docID,
		ChunkIndex: 0,
		Content:    summary,
		ChunkType:  &chunkType,
		EntityName: &entityName,
		WorldID:    &worldID,
	}
	if err := chunkRepo.Create(ctx, chunk); err != nil {
		t.Fatalf("create chunk: %v", err)
	}
}

func loadGeminiAPIKey(t *testing.T) string {
	if value := strings.TrimSpace(os.Getenv("GEMINI_API_KEY")); value != "" {
		return value
	}

	candidates := []string{
		"gemini.keys",
		filepath.Join("llm-gateway-service", "gemini.keys"),
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			if strings.Contains(trimmed, "=") {
				parts := strings.SplitN(trimmed, "=", 2)
				trimmed = strings.TrimSpace(parts[1])
			}
			if trimmed != "" {
				return trimmed
			}
		}
	}

	t.Log("gemini.keys not found or empty; set GEMINI_API_KEY")
	return ""
}
