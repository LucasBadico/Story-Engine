package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
)

// ContentAnchorRepository defines the interface for content anchor persistence
type ContentAnchorRepository interface {
	Create(ctx context.Context, anchor *story.ContentAnchor) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.ContentAnchor, error)
	ListByContentBlock(ctx context.Context, tenantID, contentBlockID uuid.UUID) ([]*story.ContentAnchor, error)
	ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType story.EntityType, entityID uuid.UUID) ([]*story.ContentAnchor, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	DeleteByContentBlock(ctx context.Context, tenantID, contentBlockID uuid.UUID) error
}


