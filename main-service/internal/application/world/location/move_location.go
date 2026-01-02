package location

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// MoveLocationUseCase handles moving a location to a different parent
type MoveLocationUseCase struct {
	locationRepo repositories.LocationRepository
	auditLogRepo repositories.AuditLogRepository
	logger       logger.Logger
}

// NewMoveLocationUseCase creates a new MoveLocationUseCase
func NewMoveLocationUseCase(
	locationRepo repositories.LocationRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *MoveLocationUseCase {
	return &MoveLocationUseCase{
		locationRepo: locationRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// MoveLocationInput represents the input for moving a location
type MoveLocationInput struct {
	LocationID uuid.UUID
	NewParentID *uuid.UUID // nil = move to root
}

// MoveLocationOutput represents the output of moving a location
type MoveLocationOutput struct {
	Location *world.Location
}

// Execute moves a location to a different parent
func (uc *MoveLocationUseCase) Execute(ctx context.Context, input MoveLocationInput) (*MoveLocationOutput, error) {
	l, err := uc.locationRepo.GetByID(ctx, input.LocationID)
	if err != nil {
		return nil, err
	}

	// Prevent circular reference: new parent cannot be a descendant of this location
	if input.NewParentID != nil {
		descendants, err := uc.locationRepo.GetDescendants(ctx, input.LocationID)
		if err != nil {
			return nil, err
		}
		for _, desc := range descendants {
			if desc.ID == *input.NewParentID {
				return nil, &platformerrors.ValidationError{
					Field:   "parent_id",
					Message: "cannot move location to its own descendant",
				}
			}
		}

		// Validate new parent exists and belongs to same world
		newParent, err := uc.locationRepo.GetByID(ctx, *input.NewParentID)
		if err != nil {
			return nil, err
		}
		if newParent.WorldID != l.WorldID {
			return nil, &platformerrors.ValidationError{
				Field:   "parent_id",
				Message: "parent location must belong to the same world",
			}
		}
		l.SetParent(input.NewParentID, newParent.HierarchyLevel)
	} else {
		l.SetParent(nil, 0)
	}

	if err := uc.locationRepo.Update(ctx, l); err != nil {
		uc.logger.Error("failed to move location", "error", err, "location_id", input.LocationID)
		return nil, err
	}

	auditLog := audit.NewAuditLog(
		l.WorldID,
		nil,
		audit.ActionUpdate,
		audit.EntityTypeLocation,
		l.ID,
		map[string]interface{}{
			"action": "move",
			"new_parent_id": input.NewParentID,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("location moved", "location_id", l.ID, "new_parent_id", input.NewParentID)

	return &MoveLocationOutput{
		Location: l,
	}, nil
}


