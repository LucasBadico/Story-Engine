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
	eventRepo            repositories.EventRepository
	eventCharacterRepo   repositories.EventCharacterRepository
	eventLocationRepo    repositories.EventLocationRepository
	eventArtifactRepo    repositories.EventArtifactRepository
	auditLogRepo         repositories.AuditLogRepository
	logger               logger.Logger
}

// NewDeleteEventUseCase creates a new DeleteEventUseCase
func NewDeleteEventUseCase(
	eventRepo repositories.EventRepository,
	eventCharacterRepo repositories.EventCharacterRepository,
	eventLocationRepo repositories.EventLocationRepository,
	eventArtifactRepo repositories.EventArtifactRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *DeleteEventUseCase {
	return &DeleteEventUseCase{
		eventRepo:          eventRepo,
		eventCharacterRepo: eventCharacterRepo,
		eventLocationRepo:  eventLocationRepo,
		eventArtifactRepo:  eventArtifactRepo,
		auditLogRepo:       auditLogRepo,
		logger:             logger,
	}
}

// DeleteEventInput represents the input for deleting an event
type DeleteEventInput struct {
	ID uuid.UUID
}

// Execute deletes an event
func (uc *DeleteEventUseCase) Execute(ctx context.Context, input DeleteEventInput) error {
	// Get event to get world_id for audit log
	event, err := uc.eventRepo.GetByID(ctx, input.ID)
	if err != nil {
		return err
	}

	// Delete junction table entries (will be handled by CASCADE, but explicit for clarity)
	if err := uc.eventCharacterRepo.DeleteByEvent(ctx, input.ID); err != nil {
		uc.logger.Error("failed to delete event characters", "error", err)
		// Continue anyway
	}
	if err := uc.eventLocationRepo.DeleteByEvent(ctx, input.ID); err != nil {
		uc.logger.Error("failed to delete event locations", "error", err)
		// Continue anyway
	}
	if err := uc.eventArtifactRepo.DeleteByEvent(ctx, input.ID); err != nil {
		uc.logger.Error("failed to delete event artifacts", "error", err)
		// Continue anyway
	}

	// Delete event
	if err := uc.eventRepo.Delete(ctx, input.ID); err != nil {
		uc.logger.Error("failed to delete event", "error", err, "event_id", input.ID)
		return err
	}

	// Log audit event
	auditLog := audit.NewAuditLog(
		event.WorldID, // Using WorldID as tenant context - TODO: get tenant from world
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


