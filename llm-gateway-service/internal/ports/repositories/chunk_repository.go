package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
)

// ChunkRepository defines the interface for chunk persistence
type ChunkRepository interface {
	Create(ctx context.Context, chunk *memory.Chunk) error
	CreateBatch(ctx context.Context, chunks []*memory.Chunk) error
	GetByID(ctx context.Context, id uuid.UUID) (*memory.Chunk, error)
	ListByDocument(ctx context.Context, documentID uuid.UUID) ([]*memory.Chunk, error)
	DeleteByDocument(ctx context.Context, documentID uuid.UUID) error
	SearchSimilar(ctx context.Context, tenantID uuid.UUID, embedding []float32, limit int, sourceTypes []memory.SourceType) ([]*memory.Chunk, error)
}

