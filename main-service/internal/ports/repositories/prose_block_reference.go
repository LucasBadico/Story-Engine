package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
)

// ProseBlockReferenceRepository defines the interface for prose block reference persistence
type ProseBlockReferenceRepository interface {
	Create(ctx context.Context, ref *story.ProseBlockReference) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.ProseBlockReference, error)
	ListByProseBlock(ctx context.Context, tenantID, proseBlockID uuid.UUID) ([]*story.ProseBlockReference, error)
	ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType story.EntityType, entityID uuid.UUID) ([]*story.ProseBlockReference, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	DeleteByProseBlock(ctx context.Context, tenantID, proseBlockID uuid.UUID) error
}


