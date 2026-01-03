package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
)

// ContentBlockRepository defines the interface for content block persistence
type ContentBlockRepository interface {
	Create(ctx context.Context, c *story.ContentBlock) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.ContentBlock, error)
	ListByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) ([]*story.ContentBlock, error)
	GetByChapterAndKind(ctx context.Context, tenantID, chapterID uuid.UUID, kind story.ContentKind) (*story.ContentBlock, error)
	Update(ctx context.Context, c *story.ContentBlock) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	DeleteByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) error
}

