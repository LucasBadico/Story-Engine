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

// UpdateLocationUseCase handles location updates
type UpdateLocationUseCase struct {
	locationRepo repositories.LocationRepository
	auditLogRepo  repositories.AuditLogRepository
	logger        logger.Logger
}

// NewUpdateLocationUseCase creates a new UpdateLocationUseCase
func NewUpdateLocationUseCase(
	locationRepo repositories.LocationRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *UpdateLocationUseCase {
	return &UpdateLocationUseCase{
		locationRepo: locationRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// UpdateLocationInput represents the input for updating a location
type UpdateLocationInput struct {
	ID          uuid.UUID
	Name        *string
	Type        *string
	Description *string
}

// UpdateLocationOutput represents the output of updating a location
type UpdateLocationOutput struct {
	Location *world.Location
}

// Execute updates a location
func (uc *UpdateLocationUseCase) Execute(ctx context.Context, input UpdateLocationInput) (*UpdateLocationOutput, error) {
	l, err := uc.locationRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		if err := l.UpdateName(*input.Name); err != nil {
			return nil, &platformerrors.ValidationError{
				Field:   "name",
				Message: err.Error(),
			}
		}
	}
	if input.Type != nil {
		l.UpdateType(*input.Type)
	}
	if input.Description != nil {
		l.UpdateDescription(*input.Description)
	}

	if err := l.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "location",
			Message: err.Error(),
		}
	}

	if err := uc.locationRepo.Update(ctx, l); err != nil {
		uc.logger.Error("failed to update location", "error", err, "location_id", input.ID)
		return nil, err
	}

	auditLog := audit.NewAuditLog(
		l.WorldID,
		nil,
		audit.ActionUpdate,
		audit.EntityTypeLocation,
		l.ID,
		map[string]interface{}{
			"name": l.Name,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("location updated", "location_id", l.ID, "name", l.Name)

	return &UpdateLocationOutput{
		Location: l,
	}, nil
}


