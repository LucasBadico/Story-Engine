package entities

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

type phase3MockRouterModel struct {
	response string
	err      error
	prompts  []string
}

func (m *phase3MockRouterModel) Generate(ctx context.Context, prompt string) (string, error) {
	m.prompts = append(m.prompts, prompt)
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func TestPhase3MatchUseCase_SelectsCandidate(t *testing.T) {
	tenantID := uuid.New()
	sourceID := uuid.New()
	docID := uuid.New()

	docRepo := repositories.NewMockDocumentRepository()
	if err := docRepo.Create(context.Background(), &memory.Document{
		ID:         docID,
		TenantID:   tenantID,
		SourceType: memory.SourceTypeCharacter,
		SourceID:   sourceID,
		Title:      "Aria",
		Content:    "Aria is a mage.",
	}); err != nil {
		t.Fatalf("create doc: %v", err)
	}

	chunkRepo := repositories.NewMockChunkRepository()
	chunkType := "summary"
	entityName := "Aria"
	chunk := &memory.Chunk{
		ID:         uuid.New(),
		DocumentID: docID,
		ChunkIndex: 0,
		Content:    "Aria is a mage.",
		ChunkType:  &chunkType,
		EntityName: &entityName,
	}
	if err := chunkRepo.Create(context.Background(), chunk); err != nil {
		t.Fatalf("create chunk: %v", err)
	}

	model := &phase3MockRouterModel{
		response: `{"match":{"index":0,"reason":"same name and summary"}}`,
	}
	embedder := embeddings.NewMockEmbedder(3)
	uc := NewPhase3MatchUseCase(chunkRepo, docRepo, embedder, model, logger.New())

	output, err := uc.Execute(context.Background(), Phase3MatchInput{
		TenantID: tenantID,
		Findings: []Phase2EntityFinding{
			{
				EntityType: "character",
				Name:       "Aria",
				Summary:    "Aria is a mage.",
			},
		},
		MinSimilarity: 0.8,
		MaxCandidates: 3,
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if len(output.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(output.Results))
	}
	result := output.Results[0]
	if result.Match == nil {
		t.Fatalf("expected a match, got nil")
	}
	if result.Match.Candidate.SourceID != sourceID {
		t.Fatalf("expected source id %s, got %s", sourceID, result.Match.Candidate.SourceID)
	}
}

func TestPhase3MatchUseCase_NoCandidatesAboveThreshold(t *testing.T) {
	tenantID := uuid.New()
	docID := uuid.New()

	docRepo := repositories.NewMockDocumentRepository()
	if err := docRepo.Create(context.Background(), &memory.Document{
		ID:         docID,
		TenantID:   tenantID,
		SourceType: memory.SourceTypeCharacter,
		SourceID:   uuid.New(),
		Title:      "Aria",
		Content:    "Aria is a mage.",
	}); err != nil {
		t.Fatalf("create doc: %v", err)
	}

	chunkRepo := repositories.NewMockChunkRepository()
	chunkType := "summary"
	chunk := &memory.Chunk{
		ID:         uuid.New(),
		DocumentID: docID,
		ChunkIndex: 0,
		Content:    "Aria is a mage.",
		ChunkType:  &chunkType,
	}
	if err := chunkRepo.Create(context.Background(), chunk); err != nil {
		t.Fatalf("create chunk: %v", err)
	}

	embedder := embeddings.NewMockEmbedder(3)
	uc := NewPhase3MatchUseCase(chunkRepo, docRepo, embedder, &phase3MockRouterModel{
		response: `{"match":{"index":0,"reason":"default"}}`,
	}, logger.New())

	output, err := uc.Execute(context.Background(), Phase3MatchInput{
		TenantID: tenantID,
		Findings: []Phase2EntityFinding{
			{
				EntityType: "character",
				Name:       "Aria",
				Summary:    "Aria is a mage.",
			},
		},
		MinSimilarity: 1.1,
		MaxCandidates: 3,
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if len(output.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(output.Results))
	}
	if output.Results[0].Match != nil {
		t.Fatalf("expected no match, got %+v", output.Results[0].Match)
	}
}
