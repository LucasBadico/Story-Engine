package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateEventUseCase handles event updates
type UpdateEventUseCase struct {
	eventRepo   repositories.EventRepository
	auditLogRepo repositories.AuditLogRepository
	logger      logger.Logger
}

// NewUpdateEventUseCase creates a new UpdateEventUseCase
func NewUpdateEventUseCase(
	eventRepo repositories.EventRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *UpdateEventUseCase {
	return &UpdateEventUseCase{
		eventRepo:   eventRepo,
		auditLogRepo: auditLogRepo,
		logger:      logger,
	}
}

// UpdateEventInput represents the input for updating an event
type UpdateEventInput struct {
	TenantID    uuid.UUID
	ID          uuid.UUID
	Name        *string
	Type        *string
	Description *string
	Timeline    *string
	Importance  *int
}

// UpdateEventOutput represents the output of updating an event
type UpdateEventOutput struct {
	Event *world.Event
}

// Execute updates an event
func (uc *UpdateEventUseCase) Execute(ctx context.Context, input UpdateEventInput) (*UpdateEventOutput, error) {
	// Get existing event
	event, err := uc.eventRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Name != nil {
		if err := event.UpdateName(*input.Name); err != nil {
			return nil, err
		}
	}
	if input.Type != nil {
		event.UpdateType(input.Type)
	}
	if input.Description != nil {
		event.UpdateDescription(input.Description)
	}
	if input.Timeline != nil {
		event.UpdateTimeline(input.Timeline)
	}
	if input.Importance != nil {
		if err := event.UpdateImportance(*input.Importance); err != nil {
			return nil, err
		}
	}

	if err := event.Validate(); err != nil {
		return nil, err
	}

	if err := uc.eventRepo.Update(ctx, event); err != nil {
		uc.logger.Error("failed to update event", "error", err, "event_id", input.ID)
		return nil, err
	}

	// Log audit event
	auditLog := audit.NewAuditLog(
		event.TenantID,
		nil,
		audit.ActionUpdate,
		audit.EntityTypeEvent,
		event.ID,
		map[string]interface{}{},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Error("failed to create audit log", "error", err)
	}

	uc.logger.Info("event updated", "event_id", input.ID)

	return &UpdateEventOutput{
		Event: event,
	}, nil
}


