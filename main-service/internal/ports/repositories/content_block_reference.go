package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
)

// ContentBlockReferenceRepository defines the interface for content block reference persistence
type ContentBlockReferenceRepository interface {
	Create(ctx context.Context, ref *story.ContentBlockReference) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.ContentBlockReference, error)
	ListByContentBlock(ctx context.Context, tenantID, contentBlockID uuid.UUID) ([]*story.ContentBlockReference, error)
	ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType story.EntityType, entityID uuid.UUID) ([]*story.ContentBlockReference, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	DeleteByContentBlock(ctx context.Context, tenantID, contentBlockID uuid.UUID) error
}

