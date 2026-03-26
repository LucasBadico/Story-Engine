package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
)

// ConceptRepository defines the interface for concept persistence
type ConceptRepository interface {
	Create(ctx context.Context, concept *memory.Concept) error
	GetByID(ctx context.Context, id uuid.UUID) (*memory.Concept, error)
	FindByName(ctx context.Context, storyID uuid.UUID, name string, conceptType memory.ConceptType) (*memory.Concept, error)
	List(ctx context.Context, storyID uuid.UUID, conceptType memory.ConceptType) ([]*memory.Concept, error)
	Update(ctx context.Context, concept *memory.Concept) error
	Delete(ctx context.Context, id uuid.UUID) error
}
