package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/ingestion-service/internal/core/memory"
)

// DocumentRepository defines the interface for document persistence
type DocumentRepository interface {
	Create(ctx context.Context, doc *memory.Document) error
	GetByID(ctx context.Context, id uuid.UUID) (*memory.Document, error)
	GetBySource(ctx context.Context, tenantID uuid.UUID, sourceType memory.SourceType, sourceID uuid.UUID) (*memory.Document, error)
	Update(ctx context.Context, doc *memory.Document) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*memory.Document, error)
}

