package timeline_event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListTimelineEventsByChapterUseCase handles listing timeline events by chapter
type ListTimelineEventsByChapterUseCase struct {
	timelineEventRepo repositories.TimelineEventRepository
	logger            logger.Logger
}

// NewListTimelineEventsByChapterUseCase creates a new ListTimelineEventsByChapterUseCase
func NewListTimelineEventsByChapterUseCase(
	timelineEventRepo repositories.TimelineEventRepository,
	logger logger.Logger,
) *ListTimelineEventsByChapterUseCase {
	return &ListTimelineEventsByChapterUseCase{
		timelineEventRepo: timelineEventRepo,
		logger:            logger,
	}
}

// ListTimelineEventsByChapterInput represents the input for listing timeline events by chapter
type ListTimelineEventsByChapterInput struct {
	TenantID  uuid.UUID
	ChapterID uuid.UUID
}

// ListTimelineEventsByChapterOutput represents the output of listing timeline events by chapter
type ListTimelineEventsByChapterOutput struct {
	TimelineEvents []*story.TimelineEvent
}

// Execute lists timeline events for a chapter
func (uc *ListTimelineEventsByChapterUseCase) Execute(ctx context.Context, input ListTimelineEventsByChapterInput) (*ListTimelineEventsByChapterOutput, error) {
	events, err := uc.timelineEventRepo.ListByChapter(ctx, input.TenantID, input.ChapterID)
	if err != nil {
		uc.logger.Error("failed to list timeline events by chapter", "error", err, "chapter_id", input.ChapterID)
		return nil, err
	}

	return &ListTimelineEventsByChapterOutput{
		TimelineEvents: events,
	}, nil
}
