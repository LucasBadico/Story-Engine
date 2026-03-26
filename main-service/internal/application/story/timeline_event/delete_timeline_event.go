package timeline_event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteTimelineEventUseCase handles timeline event deletion
type DeleteTimelineEventUseCase struct {
	timelineEventRepo repositories.TimelineEventRepository
	auditLogRepo      repositories.AuditLogRepository
	logger            logger.Logger
}

// NewDeleteTimelineEventUseCase creates a new DeleteTimelineEventUseCase
func NewDeleteTimelineEventUseCase(
	timelineEventRepo repositories.TimelineEventRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *DeleteTimelineEventUseCase {
	return &DeleteTimelineEventUseCase{
		timelineEventRepo: timelineEventRepo,
		auditLogRepo:      auditLogRepo,
		logger:            logger,
	}
}

// DeleteTimelineEventInput represents the input for deleting a timeline event
type DeleteTimelineEventInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes a timeline event
func (uc *DeleteTimelineEventUseCase) Execute(ctx context.Context, input DeleteTimelineEventInput) error {
	te, err := uc.timelineEventRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	if err := uc.timelineEventRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete timeline event", "error", err, "timeline_event_id", input.ID)
		return err
	}

	auditLog := audit.NewAuditLog(
		input.TenantID,
		nil,
		audit.ActionDelete,
		audit.EntityTypeTimelineEvent,
		te.ID,
		map[string]interface{}{
			"story_id":       te.StoryID.String(),
			"chapter_number": te.ChapterNumber,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("timeline event deleted", "timeline_event_id", input.ID)

	return nil
}
