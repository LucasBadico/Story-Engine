package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
)

// ProseBlockRepository defines the interface for prose block persistence
type ProseBlockRepository interface {
	Create(ctx context.Context, p *story.ProseBlock) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.ProseBlock, error)
	ListByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) ([]*story.ProseBlock, error)
	GetByChapterAndKind(ctx context.Context, tenantID, chapterID uuid.UUID, kind story.ProseKind) (*story.ProseBlock, error)
	Update(ctx context.Context, p *story.ProseBlock) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	DeleteByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) error
}

