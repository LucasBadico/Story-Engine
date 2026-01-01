package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
)

// SearchFilters contains filters for semantic search
type SearchFilters struct {
	SourceTypes []memory.SourceType
	BeatTypes   []string
	SceneIDs    []uuid.UUID
	Characters  []string
	LocationIDs []uuid.UUID
	StoryID     *uuid.UUID
}

// ChunkRepository defines the interface for chunk persistence
type ChunkRepository interface {
	Create(ctx context.Context, chunk *memory.Chunk) error
	CreateBatch(ctx context.Context, chunks []*memory.Chunk) error
	GetByID(ctx context.Context, id uuid.UUID) (*memory.Chunk, error)
	ListByDocument(ctx context.Context, documentID uuid.UUID) ([]*memory.Chunk, error)
	DeleteByDocument(ctx context.Context, documentID uuid.UUID) error
	SearchSimilar(ctx context.Context, tenantID uuid.UUID, embedding []float32, limit int, filters *SearchFilters) ([]*memory.Chunk, error)
}

