package event

import (
	"context"

	"github.com/google/uuid"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// MoveEventUseCase handles moving an event to another parent
type MoveEventUseCase struct {
	eventRepo repositories.EventRepository
	logger    logger.Logger
}

// NewMoveEventUseCase creates a new MoveEventUseCase
func NewMoveEventUseCase(
	eventRepo repositories.EventRepository,
	logger logger.Logger,
) *MoveEventUseCase {
	return &MoveEventUseCase{
		eventRepo: eventRepo,
		logger:    logger,
	}
}

// MoveEventInput represents the input for moving an event
type MoveEventInput struct {
	TenantID uuid.UUID
	EventID  uuid.UUID
	ParentID *uuid.UUID // nil to make it a root event
}

// Execute moves an event to another parent and recalculates hierarchy levels
func (uc *MoveEventUseCase) Execute(ctx context.Context, input MoveEventInput) error {
	// Get event
	event, err := uc.eventRepo.GetByID(ctx, input.TenantID, input.EventID)
	if err != nil {
		return err
	}

	// Validate parent if provided
	var parentLevel int
	if input.ParentID != nil {
		parent, err := uc.eventRepo.GetByID(ctx, input.TenantID, *input.ParentID)
		if err != nil {
			return err
		}
		if parent.WorldID != event.WorldID {
			return &platformerrors.ValidationError{
				Field:   "parent_id",
				Message: "parent event must belong to the same world",
			}
		}
		// Prevent circular reference
		ancestors, err := uc.eventRepo.GetAncestors(ctx, input.TenantID, *input.ParentID)
		if err != nil {
			return err
		}
		for _, ancestor := range ancestors {
			if ancestor.ID == event.ID {
				return &platformerrors.ValidationError{
					Field:   "parent_id",
					Message: "cannot set parent to a descendant",
				}
			}
		}
		parentLevel = parent.HierarchyLevel
	}

	// Update event parent and hierarchy level
	event.SetParent(input.ParentID, parentLevel)
	if err := uc.eventRepo.Update(ctx, event); err != nil {
		uc.logger.Error("failed to move event", "error", err, "event_id", input.EventID)
		return err
	}

	// Recalculate hierarchy levels for all descendants
	descendants, err := uc.eventRepo.GetDescendants(ctx, input.TenantID, event.ID)
	if err != nil {
		uc.logger.Error("failed to get descendants for hierarchy update", "error", err)
		return err
	}

	for _, descendant := range descendants {
		// Get current parent to calculate level
		if descendant.ParentID != nil {
			parent, err := uc.eventRepo.GetByID(ctx, input.TenantID, *descendant.ParentID)
			if err != nil {
				uc.logger.Error("failed to get parent for descendant", "error", err, "descendant_id", descendant.ID)
				continue
			}
			descendant.SetHierarchyLevel(parent.HierarchyLevel + 1)
			if err := uc.eventRepo.Update(ctx, descendant); err != nil {
				uc.logger.Error("failed to update descendant hierarchy", "error", err, "descendant_id", descendant.ID)
				// Continue with other descendants
			}
		}
	}

	uc.logger.Info("event moved", "event_id", input.EventID, "parent_id", input.ParentID)
	return nil
}

