package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
)

// ProseBlockRepository defines the interface for prose block persistence
type ProseBlockRepository interface {
	Create(ctx context.Context, p *story.ProseBlock) error
	GetByID(ctx context.Context, id uuid.UUID) (*story.ProseBlock, error)
	ListByScene(ctx context.Context, sceneID uuid.UUID) ([]*story.ProseBlock, error)
	GetBySceneAndKind(ctx context.Context, sceneID uuid.UUID, kind story.ProseKind) (*story.ProseBlock, error)
	Update(ctx context.Context, p *story.ProseBlock) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByScene(ctx context.Context, sceneID uuid.UUID) error
}

