package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
)

// SearchFilters contains filters for semantic search
type SearchFilters struct {
	SourceTypes []memory.SourceType
	ChunkTypes  []string
	BeatTypes   []string
	SceneIDs    []uuid.UUID
	Characters  []string
	LocationIDs []uuid.UUID
	StoryID     *uuid.UUID
}

// SearchCursor defines the position in a similarity-ordered result set.
type SearchCursor struct {
	Distance float64
	ChunkID  uuid.UUID
}

// ScoredChunk contains a chunk and its vector distance to the query.
type ScoredChunk struct {
	Chunk    *memory.Chunk
	Distance float64
}

// ChunkRepository defines the interface for chunk persistence
type ChunkRepository interface {
	Create(ctx context.Context, chunk *memory.Chunk) error
	CreateBatch(ctx context.Context, chunks []*memory.Chunk) error
	GetByID(ctx context.Context, id uuid.UUID) (*memory.Chunk, error)
	ListByDocument(ctx context.Context, documentID uuid.UUID) ([]*memory.Chunk, error)
	DeleteByDocument(ctx context.Context, documentID uuid.UUID) error
	SearchSimilar(ctx context.Context, tenantID uuid.UUID, embedding []float32, limit int, cursor *SearchCursor, filters *SearchFilters) ([]*ScoredChunk, error)
}
