//go:build integration

package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/adapters/llm/gemini"
	"github.com/story-engine/llm-gateway-service/internal/application/extract"
	"github.com/story-engine/llm-gateway-service/internal/application/extract/entities"
	"github.com/story-engine/llm-gateway-service/internal/application/extract/relations"
	"github.com/story-engine/llm-gateway-service/internal/application/search"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

func TestEntityExtractHandler_RelationsIntegration(t *testing.T) {
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

	contentBlockID := uuid.New()
	doc := memory.NewDocument(tenantID, memory.SourceTypeContentBlock, contentBlockID, "Scene 1", "Ari swore loyalty to the Order of the Sun.")
	if err := docRepo.Create(ctx, doc); err != nil {
		t.Fatalf("create doc: %v", err)
	}

	chunk := memory.NewChunk(doc.ID, 0, doc.Content, []float32{0.1, 0.1, 0.1}, 10)
	if err := chunkRepo.Create(ctx, chunk); err != nil {
		t.Fatalf("create chunk: %v", err)
	}

	searchUseCase := search.NewSearchMemoryUseCase(chunkRepo, docRepo, embedder, log)

	model := gemini.NewRouterModel(apiKey, modelName)
	router := entities.NewPhase1EntityTypeRouterUseCase(model, log)
	extractor := entities.NewPhase2EntryUseCase(model, log, nil)
	matcher := entities.NewPhase3MatchUseCase(chunkRepo, docRepo, embedder, model, log)
	payload := entities.NewPhase4EntitiesPayloadUseCase()
	relationDiscovery := relations.NewPhase5RelationDiscoveryUseCase(model, log)
	relationNormalize := relations.NewPhase6RelationNormalizeUseCase(log)
	relationNormalize.SetSummaryModel(model)
	relationMatcher := relations.NewPhase7RelationMatchUseCase(searchUseCase, log)
	useCase := extract.NewExtractOrchestrator(
		router,
		extractor,
		matcher,
		payload,
		relationDiscovery,
		relationNormalize,
		relationMatcher,
		log,
	)

	relationTypes := map[string]relations.Phase6RelationTypeDefinition{
		"member_of": {
			Mirror:             "has_member",
			PreferredDirection: "source_to_target",
			Semantics:          "Source is a member of target.",
		},
	}
	suggestedRelations := map[string]relations.Phase5PerEntityRelationMap{
		"character": {
			EntityType: "character",
			Version:    1,
			Relations: map[string]relations.Phase5RelationConstraintSpec{
				"member_of": {
					PairCandidates: []string{"faction"},
					Description:    "Character belongs to a group or organization.",
					Constraints: &relations.Phase5RelationConstraints{
						MinConfidence:    0.5,
						AllowImplicit:    true,
						RequiresEvidence: true,
					},
				},
			},
		},
	}
	handler := NewEntityExtractHandler(useCase, relationTypes, suggestedRelations, log)

	timeoutCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	body, _ := json.Marshal(map[string]interface{}{
		"text":         "Ari swore loyalty to the Order of the Sun.",
		"world_id":     worldID.String(),
		"entity_types": []string{"character", "faction"},
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

	raw := rr.Body.Bytes()
	t.Logf("response: %s", string(raw))

	var resp struct {
		Relations []struct {
			Source struct {
				Ref  string `json:"ref"`
				ID   string `json:"id"`
				Type string `json:"type"`
				Name string `json:"name"`
			} `json:"source"`
			Target struct {
				Ref  string `json:"ref"`
				ID   string `json:"id"`
				Type string `json:"type"`
				Name string `json:"name"`
			} `json:"target"`
			RelationType string `json:"relation_type"`
			Summary      string `json:"summary"`
			Status       string `json:"status"`
			Matches      []struct {
				ChunkID uuid.UUID `json:"chunk_id"`
			} `json:"matches"`
		} `json:"relations"`
	}
	if err := json.NewDecoder(bytes.NewReader(raw)).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(resp.Relations) == 0 {
		t.Fatalf("expected relations, got none")
	}

	var hasMemberOf bool
	for _, rel := range resp.Relations {
		if rel.RelationType == "member_of" {
			hasMemberOf = true
			if strings.TrimSpace(rel.Source.Type) == "" || strings.TrimSpace(rel.Target.Type) == "" {
				t.Fatalf("expected source/target types for member_of relation")
			}
			if strings.TrimSpace(rel.Status) == "" {
				t.Fatalf("expected status for member_of relation")
			}
			if len(rel.Matches) == 0 {
				t.Fatalf("expected relation matches for member_of")
			}
		}
	}
	if !hasMemberOf {
		t.Fatalf("expected member_of relation")
	}
}
