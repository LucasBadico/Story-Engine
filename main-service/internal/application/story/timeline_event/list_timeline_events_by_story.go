package timeline_event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListTimelineEventsByStoryUseCase handles listing timeline events by story
type ListTimelineEventsByStoryUseCase struct {
	timelineEventRepo repositories.TimelineEventRepository
	logger            logger.Logger
}

// NewListTimelineEventsByStoryUseCase creates a new ListTimelineEventsByStoryUseCase
func NewListTimelineEventsByStoryUseCase(
	timelineEventRepo repositories.TimelineEventRepository,
	logger logger.Logger,
) *ListTimelineEventsByStoryUseCase {
	return &ListTimelineEventsByStoryUseCase{
		timelineEventRepo: timelineEventRepo,
		logger:            logger,
	}
}

// ListTimelineEventsByStoryInput represents the input for listing timeline events by story
type ListTimelineEventsByStoryInput struct {
	TenantID    uuid.UUID
	StoryID     uuid.UUID
	AsOfChapter int
	Limit       int
}

// ListTimelineEventsByStoryOutput represents the output of listing timeline events by story
type ListTimelineEventsByStoryOutput struct {
	TimelineEvents []*story.TimelineEvent
}

// Execute lists timeline events for a story up to the given chapter
func (uc *ListTimelineEventsByStoryUseCase) Execute(ctx context.Context, input ListTimelineEventsByStoryInput) (*ListTimelineEventsByStoryOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	events, err := uc.timelineEventRepo.ListByStory(ctx, input.TenantID, input.StoryID, input.AsOfChapter, limit)
	if err != nil {
		uc.logger.Error("failed to list timeline events by story", "error", err, "story_id", input.StoryID)
		return nil, err
	}

	return &ListTimelineEventsByStoryOutput{
		TimelineEvents: events,
	}, nil
}
