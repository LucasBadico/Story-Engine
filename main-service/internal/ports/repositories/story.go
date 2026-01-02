package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
)

// StoryRepository defines the interface for story persistence
type StoryRepository interface {
	Create(ctx context.Context, s *story.Story) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.Story, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*story.Story, error)
	ListVersionsByRoot(ctx context.Context, tenantID, rootStoryID uuid.UUID) ([]*story.Story, error)
	Update(ctx context.Context, s *story.Story) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
	GetVersionGraph(ctx context.Context, tenantID, rootStoryID uuid.UUID) ([]*story.Story, error)
}

