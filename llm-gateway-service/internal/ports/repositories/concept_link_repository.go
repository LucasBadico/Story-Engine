package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
)

// ConceptLinkRepository defines the interface for concept link persistence
type ConceptLinkRepository interface {
	Create(ctx context.Context, link *memory.ConceptLink) error
	GetByID(ctx context.Context, id uuid.UUID) (*memory.ConceptLink, error)
	FindByPair(ctx context.Context, sourceID, targetID uuid.UUID) (*memory.ConceptLink, error)
	List(ctx context.Context, storyID uuid.UUID, asOfChapter *int) ([]*memory.ConceptLink, error)
	ListByConceptID(ctx context.Context, conceptID uuid.UUID) ([]*memory.ConceptLink, error)
	CloseVersion(ctx context.Context, id uuid.UUID, validToChapter int) error
	Delete(ctx context.Context, id uuid.UUID) error
}
