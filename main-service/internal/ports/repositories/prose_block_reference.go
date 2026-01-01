package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
)

// ProseBlockReferenceRepository defines the interface for prose block reference persistence
type ProseBlockReferenceRepository interface {
	Create(ctx context.Context, ref *story.ProseBlockReference) error
	GetByID(ctx context.Context, id uuid.UUID) (*story.ProseBlockReference, error)
	ListByProseBlock(ctx context.Context, proseBlockID uuid.UUID) ([]*story.ProseBlockReference, error)
	ListByEntity(ctx context.Context, entityType story.EntityType, entityID uuid.UUID) ([]*story.ProseBlockReference, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByProseBlock(ctx context.Context, proseBlockID uuid.UUID) error
}

