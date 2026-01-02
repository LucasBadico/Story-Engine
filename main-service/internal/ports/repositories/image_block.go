package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
)

// ImageBlockRepository defines the interface for image block persistence
type ImageBlockRepository interface {
	Create(ctx context.Context, ib *story.ImageBlock) error
	GetByID(ctx context.Context, id uuid.UUID) (*story.ImageBlock, error)
	ListByChapter(ctx context.Context, chapterID uuid.UUID) ([]*story.ImageBlock, error)
	Update(ctx context.Context, ib *story.ImageBlock) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByChapter(ctx context.Context, chapterID uuid.UUID) error
}

// ImageBlockReferenceRepository defines the interface for image block reference persistence
type ImageBlockReferenceRepository interface {
	Create(ctx context.Context, ref *story.ImageBlockReference) error
	GetByID(ctx context.Context, id uuid.UUID) (*story.ImageBlockReference, error)
	ListByImageBlock(ctx context.Context, imageBlockID uuid.UUID) ([]*story.ImageBlockReference, error)
	ListByEntity(ctx context.Context, entityType story.ImageBlockReferenceEntityType, entityID uuid.UUID) ([]*story.ImageBlockReference, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByImageBlock(ctx context.Context, imageBlockID uuid.UUID) error
	DeleteByImageBlockAndEntity(ctx context.Context, imageBlockID uuid.UUID, entityType story.ImageBlockReferenceEntityType, entityID uuid.UUID) error
}


