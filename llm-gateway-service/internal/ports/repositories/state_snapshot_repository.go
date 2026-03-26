package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
)

// StateSnapshotRepository defines the interface for state snapshot persistence
type StateSnapshotRepository interface {
	Create(ctx context.Context, snapshot *memory.StateSnapshot) error
	GetLatestByEntity(ctx context.Context, storyID, entityID uuid.UUID, asOfChapter int) (*memory.StateSnapshot, error)
	GetLatestByEntityAtDay(ctx context.Context, storyID, entityID uuid.UUID, asOfDay int) (*memory.StateSnapshot, error)
	ListByStory(ctx context.Context, storyID uuid.UUID, asOfChapter int) ([]*memory.StateSnapshot, error)
	ListByStoryAtDay(ctx context.Context, storyID uuid.UUID, asOfDay int) ([]*memory.StateSnapshot, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
