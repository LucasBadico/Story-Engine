package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
)

// EventRepository defines the interface for event persistence
type EventRepository interface {
	Create(ctx context.Context, event *memory.Event) error
	ListByStory(ctx context.Context, storyID uuid.UUID, asOfChapter int, limit int) ([]*memory.Event, error)
	ListByStoryAtDay(ctx context.Context, storyID uuid.UUID, asOfDay int, limit int) ([]*memory.Event, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
