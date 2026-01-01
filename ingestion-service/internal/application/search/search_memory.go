package search

import (
	"context"
	"math"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

// SearchMemoryInput is the input for searching memory
type SearchMemoryInput struct {
	TenantID    uuid.UUID
	Query       string
	Limit       int
	SourceTypes []memory.SourceType // Optional filter by source types
}

// SearchMemoryOutput is the output containing relevant chunks
type SearchMemoryOutput struct {
	Chunks []*SearchResult
}

// SearchResult represents a search result chunk
type SearchResult struct {
	ChunkID     uuid.UUID
	DocumentID  uuid.UUID
	SourceType  memory.SourceType
	SourceID    uuid.UUID
	Content     string
	Similarity  float64 // Cosine similarity score
}

// SearchMemoryUseCase handles semantic memory search
type SearchMemoryUseCase struct {
	chunkRepo repositories.ChunkRepository
	docRepo   repositories.DocumentRepository
	embedder  embeddings.Embedder
	logger    logger.Logger
}

// NewSearchMemoryUseCase creates a new SearchMemoryUseCase
func NewSearchMemoryUseCase(
	chunkRepo repositories.ChunkRepository,
	docRepo repositories.DocumentRepository,
	embedder embeddings.Embedder,
	logger logger.Logger,
) *SearchMemoryUseCase {
	return &SearchMemoryUseCase{
		chunkRepo: chunkRepo,
		docRepo:   docRepo,
		embedder:  embedder,
		logger:    logger,
	}
}

// Execute searches for relevant memory chunks based on semantic similarity
func (uc *SearchMemoryUseCase) Execute(ctx context.Context, input SearchMemoryInput) (*SearchMemoryOutput, error) {
	// Generate embedding for query
	queryEmbedding, err := uc.embedder.EmbedText(input.Query)
	if err != nil {
		return nil, err
	}

	// Set default limit
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}

	// Search for similar chunks
	chunks, err := uc.chunkRepo.SearchSimilar(ctx, input.TenantID, queryEmbedding, limit, input.SourceTypes)
	if err != nil {
		return nil, err
	}

	// Convert chunks to search results
	results := make([]*SearchResult, 0, len(chunks))
	for _, chunk := range chunks {
		// Get document for source info
		doc, err := uc.docRepo.GetByID(ctx, chunk.DocumentID)
		if err != nil {
			uc.logger.Error("failed to get document", "document_id", chunk.DocumentID, "error", err)
			continue
		}

		// Calculate cosine similarity
		similarity := uc.calculateSimilarity(queryEmbedding, chunk.Embedding)

		results = append(results, &SearchResult{
			ChunkID:    chunk.ID,
			DocumentID: chunk.DocumentID,
			SourceType: doc.SourceType,
			SourceID:   doc.SourceID,
			Content:    chunk.Content,
			Similarity: similarity,
		})
	}

	return &SearchMemoryOutput{
		Chunks: results,
	}, nil
}

// calculateSimilarity calculates cosine similarity between two vectors
func (uc *SearchMemoryUseCase) calculateSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	// Cosine similarity: dot product / (sqrt(normA) * sqrt(normB))
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

