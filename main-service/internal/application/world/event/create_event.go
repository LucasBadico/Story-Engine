package event

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateEventUseCase handles event creation
type CreateEventUseCase struct {
	eventRepo   repositories.EventRepository
	worldRepo   repositories.WorldRepository
	auditLogRepo repositories.AuditLogRepository
	logger      logger.Logger
}

// NewCreateEventUseCase creates a new CreateEventUseCase
func NewCreateEventUseCase(
	eventRepo repositories.EventRepository,
	worldRepo repositories.WorldRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *CreateEventUseCase {
	return &CreateEventUseCase{
		eventRepo:   eventRepo,
		worldRepo:   worldRepo,
		auditLogRepo: auditLogRepo,
		logger:      logger,
	}
}

// CreateEventInput represents the input for creating an event
type CreateEventInput struct {
	TenantID        uuid.UUID
	WorldID         uuid.UUID
	Name            string
	Type            *string
	Description     *string
	Timeline        *string
	Importance      int
	ParentID        *uuid.UUID
	TimelinePosition float64
	IsEpoch         bool
}

// CreateEventOutput represents the output of creating an event
type CreateEventOutput struct {
	Event *world.Event
}

// Execute creates a new event
func (uc *CreateEventUseCase) Execute(ctx context.Context, input CreateEventInput) (*CreateEventOutput, error) {
	// Validate world exists
	_, err := uc.worldRepo.GetByID(ctx, input.TenantID, input.WorldID)
	if err != nil {
		return nil, err
	}

	// Create event
	evt, err := world.NewEvent(input.TenantID, input.WorldID, input.Name)
	if err != nil {
		if errors.Is(err, world.ErrEventNameRequired) {
			return nil, &platformerrors.ValidationError{
				Field:   "name",
				Message: err.Error(),
			}
		}
		return nil, err
	}

	if input.Type != nil {
		evt.UpdateType(input.Type)
	}
	if input.Description != nil {
		evt.UpdateDescription(input.Description)
	}
	if input.Timeline != nil {
		evt.UpdateTimeline(input.Timeline)
	}
	if input.Importance > 0 {
		if err := evt.UpdateImportance(input.Importance); err != nil {
			return nil, &platformerrors.ValidationError{
				Field:   "importance",
				Message: err.Error(),
			}
		}
	}

	// Set hierarchy fields
	if input.ParentID != nil {
		parent, err := uc.eventRepo.GetByID(ctx, input.TenantID, *input.ParentID)
		if err != nil {
			return nil, &platformerrors.ValidationError{
				Field:   "parent_id",
				Message: "parent event not found",
			}
		}
		if parent.WorldID != evt.WorldID {
			return nil, &platformerrors.ValidationError{
				Field:   "parent_id",
				Message: "parent event must belong to the same world",
			}
		}
		evt.SetParent(input.ParentID, parent.HierarchyLevel)
	}

	// Set timeline position
	if input.TimelinePosition != 0 {
		evt.SetTimelinePosition(input.TimelinePosition)
	}

	// Set epoch
	if input.IsEpoch {
		// Validate that there's no existing epoch
		existingEpoch, err := uc.eventRepo.GetEpoch(ctx, input.TenantID, input.WorldID)
		if err == nil {
			// Remove epoch from existing event
			existingEpoch.SetAsEpoch(false)
			if err := uc.eventRepo.Update(ctx, existingEpoch); err != nil {
				uc.logger.Error("failed to remove epoch from existing event", "error", err)
			}
		}
		evt.SetAsEpoch(true)
	}

	if err := evt.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "event",
			Message: err.Error(),
		}
	}

	if err := uc.eventRepo.Create(ctx, evt); err != nil {
		uc.logger.Error("failed to create event", "error", err, "world_id", input.WorldID)
		return nil, err
	}

	// Log audit event
	auditLog := audit.NewAuditLog(
		evt.TenantID,
		nil,
		audit.ActionCreate,
		audit.EntityTypeEvent,
		evt.ID,
		map[string]interface{}{
			"name": evt.Name,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Error("failed to create audit log", "error", err)
	}

	uc.logger.Info("event created", "event_id", evt.ID, "world_id", input.WorldID)

	return &CreateEventOutput{
		Event: evt,
	}, nil
}


