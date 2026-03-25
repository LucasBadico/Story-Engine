package timeline_event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/queue"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// EntityLinkInput represents an entity link to be created
type EntityLinkInput struct {
	Role       string
	EntityType string
	EntityID   uuid.UUID
}

// CreateTimelineEventUseCase handles timeline event creation
type CreateTimelineEventUseCase struct {
	timelineEventRepo repositories.TimelineEventRepository
	storyRepo         repositories.StoryRepository
	auditLogRepo      repositories.AuditLogRepository
	ingestionQueue    queue.IngestionQueue
	logger            logger.Logger
}

// NewCreateTimelineEventUseCase creates a new CreateTimelineEventUseCase
func NewCreateTimelineEventUseCase(
	timelineEventRepo repositories.TimelineEventRepository,
	storyRepo repositories.StoryRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *CreateTimelineEventUseCase {
	return &CreateTimelineEventUseCase{
		timelineEventRepo: timelineEventRepo,
		storyRepo:         storyRepo,
		auditLogRepo:      auditLogRepo,
		ingestionQueue:    nil,
		logger:            logger,
	}
}

func (uc *CreateTimelineEventUseCase) SetIngestionQueue(queue queue.IngestionQueue) {
	uc.ingestionQueue = queue
}

// CreateTimelineEventInput represents the input for creating a timeline event
type CreateTimelineEventInput struct {
	TenantID           uuid.UUID
	StoryID            uuid.UUID
	ChapterID          uuid.UUID
	ChapterNumber      int
	TimelineDay        *int
	TimelineConfidence *string
	Description        string
	Entities           []EntityLinkInput
}

// CreateTimelineEventOutput represents the output of creating a timeline event
type CreateTimelineEventOutput struct {
	TimelineEvent *story.TimelineEvent
}

// Execute creates a new timeline event
func (uc *CreateTimelineEventUseCase) Execute(ctx context.Context, input CreateTimelineEventInput) (*CreateTimelineEventOutput, error) {
	// Validate story exists
	_, err := uc.storyRepo.GetByID(ctx, input.TenantID, input.StoryID)
	if err != nil {
		return nil, err
	}

	newEvent, err := story.NewTimelineEvent(input.TenantID, input.StoryID, input.ChapterNumber, input.Description)
	if err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "timeline_event",
			Message: err.Error(),
		}
	}

	newEvent.SetChapterID(input.ChapterID)
	if input.TimelineDay != nil && input.TimelineConfidence != nil {
		newEvent.SetTimeline(*input.TimelineDay, *input.TimelineConfidence)
	}

	// Add entity links
	for _, e := range input.Entities {
		newEvent.AddEntity(e.Role, e.EntityType, e.EntityID)
	}

	if err := newEvent.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "timeline_event",
			Message: err.Error(),
		}
	}

	if err := uc.timelineEventRepo.Create(ctx, newEvent); err != nil {
		uc.logger.Error("failed to create timeline event", "error", err, "story_id", input.StoryID)
		return nil, err
	}

	// Create entity links in DB
	for _, entity := range newEvent.Entities {
		if err := uc.timelineEventRepo.AddEntity(ctx, &entity); err != nil {
			uc.logger.Error("failed to add entity to timeline event", "error", err, "entity_id", entity.EntityID)
		}
	}

	auditLog := audit.NewAuditLog(
		input.TenantID,
		nil,
		audit.ActionCreate,
		audit.EntityTypeTimelineEvent,
		newEvent.ID,
		map[string]interface{}{
			"story_id":       input.StoryID.String(),
			"chapter_number": input.ChapterNumber,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("timeline event created", "timeline_event_id", newEvent.ID, "story_id", input.StoryID)
	uc.enqueueIngestion(ctx, input.TenantID, newEvent.ID)

	return &CreateTimelineEventOutput{
		TimelineEvent: newEvent,
	}, nil
}

func (uc *CreateTimelineEventUseCase) enqueueIngestion(ctx context.Context, tenantID uuid.UUID, timelineEventID uuid.UUID) {
	if uc.ingestionQueue == nil {
		return
	}
	if err := uc.ingestionQueue.Push(ctx, tenantID, "timeline_event", timelineEventID); err != nil {
		uc.logger.Error("failed to enqueue timeline event ingestion", "error", err, "timeline_event_id", timelineEventID, "tenant_id", tenantID)
	}
}
