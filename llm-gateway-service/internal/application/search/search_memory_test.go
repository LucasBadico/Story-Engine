package search

import (
	"context"
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

func TestSearchMemoryUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	documentID := uuid.New()

	mockChunkRepo := repositories.NewMockChunkRepository()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	// Create a document
	doc := memory.NewDocument(tenantID, memory.SourceTypeStory, uuid.New(), "Test Story", "Content")
	doc.ID = documentID
	mockDocRepo.Create(ctx, doc)

	// Create a chunk
	chunk := memory.NewChunk(documentID, 0, "Test content", []float32{0.1, 0.2, 0.3}, 10)
	mockChunkRepo.Create(ctx, chunk)

	useCase := NewSearchMemoryUseCase(mockChunkRepo, mockDocRepo, mockEmbedder, log)

	input := SearchMemoryInput{
		TenantID: tenantID,
		Query:    "test query",
		Limit:    10,
	}

	output, err := useCase.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(output.Chunks) == 0 {
		t.Error("Expected at least one result")
	}
}

func TestSearchMemoryUseCase_Execute_WithFilters(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	documentID := uuid.New()
	sceneID := uuid.New()

	mockChunkRepo := repositories.NewMockChunkRepository()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	// Create a document
	doc := memory.NewDocument(tenantID, memory.SourceTypeProseBlock, uuid.New(), "Test", "Content")
	doc.ID = documentID
	mockDocRepo.Create(ctx, doc)

	// Create a chunk with metadata
	beatType := "setup"
	chunk := memory.NewChunk(documentID, 0, "Test content", []float32{0.1, 0.2, 0.3}, 10)
	chunk.SceneID = &sceneID
	chunk.BeatType = &beatType
	chunk.Characters = []string{"John"}
	mockChunkRepo.Create(ctx, chunk)

	useCase := NewSearchMemoryUseCase(mockChunkRepo, mockDocRepo, mockEmbedder, log)

	input := SearchMemoryInput{
		TenantID:    tenantID,
		Query:       "test query",
		Limit:       10,
		BeatTypes:   []string{"setup"},
		SceneIDs:    []uuid.UUID{sceneID},
		Characters: []string{"John"},
	}

	output, err := useCase.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(output.Chunks) == 0 {
		t.Error("Expected at least one result")
	}

	result := output.Chunks[0]
	if result.BeatType == nil || *result.BeatType != "setup" {
		t.Error("Expected BeatType to be set in result")
	}
	if len(result.Characters) == 0 {
		t.Error("Expected Characters to be set in result")
	}
}

func TestSearchMemoryUseCase_Execute_EmptyResults(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	mockChunkRepo := repositories.NewMockChunkRepository()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	useCase := NewSearchMemoryUseCase(mockChunkRepo, mockDocRepo, mockEmbedder, log)

	input := SearchMemoryInput{
		TenantID: tenantID,
		Query:    "nonexistent query",
		Limit:    10,
	}

	output, err := useCase.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(output.Chunks) != 0 {
		t.Errorf("Expected 0 results, got %d", len(output.Chunks))
	}
}

func TestSearchMemoryUseCase_calculateSimilarity(t *testing.T) {
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	useCase := NewSearchMemoryUseCase(mockChunkRepo, mockDocRepo, mockEmbedder, log)

	// Test with identical vectors
	a := []float32{1.0, 0.0, 0.0}
	b := []float32{1.0, 0.0, 0.0}
	similarity := useCase.calculateSimilarity(a, b)
	if math.Abs(similarity-1.0) > 0.0001 {
		t.Errorf("Expected similarity 1.0 for identical vectors, got %f", similarity)
	}

	// Test with orthogonal vectors
	a = []float32{1.0, 0.0}
	b = []float32{0.0, 1.0}
	similarity = useCase.calculateSimilarity(a, b)
	if math.Abs(similarity-0.0) > 0.0001 {
		t.Errorf("Expected similarity 0.0 for orthogonal vectors, got %f", similarity)
	}

	// Test with different length vectors
	a = []float32{1.0, 0.0}
	b = []float32{1.0, 0.0, 0.0}
	similarity = useCase.calculateSimilarity(a, b)
	if similarity != 0.0 {
		t.Errorf("Expected similarity 0.0 for different length vectors, got %f", similarity)
	}

	// Test with zero vectors
	a = []float32{0.0, 0.0}
	b = []float32{0.0, 0.0}
	similarity = useCase.calculateSimilarity(a, b)
	if similarity != 0.0 {
		t.Errorf("Expected similarity 0.0 for zero vectors, got %f", similarity)
	}
}

