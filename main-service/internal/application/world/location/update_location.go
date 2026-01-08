package location

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/queue"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateLocationUseCase handles location updates
type UpdateLocationUseCase struct {
	locationRepo repositories.LocationRepository
	auditLogRepo  repositories.AuditLogRepository
	ingestionQueue queue.IngestionQueue
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
		ingestionQueue: nil,
		logger:       logger,
	}
}

func (uc *UpdateLocationUseCase) SetIngestionQueue(queue queue.IngestionQueue) {
	uc.ingestionQueue = queue
}

// UpdateLocationInput represents the input for updating a location
type UpdateLocationInput struct {
	TenantID    uuid.UUID
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
	l, err := uc.locationRepo.GetByID(ctx, input.TenantID, input.ID)
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
		l.TenantID,
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
	uc.enqueueIngestion(ctx, input.TenantID, l.ID)

	return &UpdateLocationOutput{
		Location: l,
	}, nil
}

func (uc *UpdateLocationUseCase) enqueueIngestion(ctx context.Context, tenantID uuid.UUID, locationID uuid.UUID) {
	if uc.ingestionQueue == nil {
		return
	}
	if err := uc.ingestionQueue.Push(ctx, tenantID, "location", locationID); err != nil {
		uc.logger.Error("failed to enqueue location ingestion", "error", err, "location_id", locationID, "tenant_id", tenantID)
	}
}

