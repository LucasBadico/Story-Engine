package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteEventUseCase handles event deletion
type DeleteEventUseCase struct {
	eventRepo          repositories.EventRepository
	eventReferenceRepo repositories.EventReferenceRepository
	auditLogRepo       repositories.AuditLogRepository
	logger             logger.Logger
}

// NewDeleteEventUseCase creates a new DeleteEventUseCase
func NewDeleteEventUseCase(
	eventRepo repositories.EventRepository,
	eventReferenceRepo repositories.EventReferenceRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *DeleteEventUseCase {
	return &DeleteEventUseCase{
		eventRepo:          eventRepo,
		eventReferenceRepo: eventReferenceRepo,
		auditLogRepo:       auditLogRepo,
		logger:             logger,
	}
}

// DeleteEventInput represents the input for deleting an event
type DeleteEventInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes an event
func (uc *DeleteEventUseCase) Execute(ctx context.Context, input DeleteEventInput) error {
	// Get event to get tenant_id for audit log
	event, err := uc.eventRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	// Delete event references (will be handled by CASCADE, but explicit for clarity)
	if err := uc.eventReferenceRepo.DeleteByEvent(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete event references", "error", err)
		// Continue anyway
	}

	// Delete event
	if err := uc.eventRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete event", "error", err, "event_id", input.ID)
		return err
	}

	// Log audit event
	auditLog := audit.NewAuditLog(
		event.TenantID,
		nil,
		audit.ActionDelete,
		audit.EntityTypeEvent,
		input.ID,
		map[string]interface{}{},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Error("failed to create audit log", "error", err)
	}

	uc.logger.Info("event deleted", "event_id", input.ID)

	return nil
}


