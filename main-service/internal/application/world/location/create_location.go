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

// CreateLocationUseCase handles location creation
type CreateLocationUseCase struct {
	locationRepo repositories.LocationRepository
	worldRepo    repositories.WorldRepository
	auditLogRepo repositories.AuditLogRepository
	ingestionQueue queue.IngestionQueue
	logger       logger.Logger
}

// NewCreateLocationUseCase creates a new CreateLocationUseCase
func NewCreateLocationUseCase(
	locationRepo repositories.LocationRepository,
	worldRepo repositories.WorldRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *CreateLocationUseCase {
	return &CreateLocationUseCase{
		locationRepo: locationRepo,
		worldRepo:    worldRepo,
		auditLogRepo: auditLogRepo,
		ingestionQueue: nil,
		logger:       logger,
	}
}

func (uc *CreateLocationUseCase) SetIngestionQueue(queue queue.IngestionQueue) {
	uc.ingestionQueue = queue
}

// CreateLocationInput represents the input for creating a location
type CreateLocationInput struct {
	TenantID    uuid.UUID
	WorldID     uuid.UUID
	ParentID    *uuid.UUID
	Name        string
	Type        string
	Description string
}

// CreateLocationOutput represents the output of creating a location
type CreateLocationOutput struct {
	Location *world.Location
}

// Execute creates a new location
func (uc *CreateLocationUseCase) Execute(ctx context.Context, input CreateLocationInput) (*CreateLocationOutput, error) {
	// Validate world exists
	_, err := uc.worldRepo.GetByID(ctx, input.TenantID, input.WorldID)
	if err != nil {
		return nil, err
	}

	// Validate parent exists if provided
	if input.ParentID != nil {
		parent, err := uc.locationRepo.GetByID(ctx, input.TenantID, *input.ParentID)
		if err != nil {
			return nil, err
		}
		if parent.WorldID != input.WorldID {
			return nil, &platformerrors.ValidationError{
				Field:   "parent_id",
				Message: "parent location must belong to the same world",
			}
		}
	}

	newLocation, err := world.NewLocation(input.TenantID, input.WorldID, input.Name, input.ParentID)
	if err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "name",
			Message: err.Error(),
		}
	}

	if input.Type != "" {
		newLocation.UpdateType(input.Type)
	}
	if input.Description != "" {
		newLocation.UpdateDescription(input.Description)
	}

	if err := newLocation.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "location",
			Message: err.Error(),
		}
	}

	if err := uc.locationRepo.Create(ctx, newLocation); err != nil {
		uc.logger.Error("failed to create location", "error", err, "name", input.Name)
		return nil, err
	}

	auditLog := audit.NewAuditLog(
		input.TenantID,
		nil,
		audit.ActionCreate,
		audit.EntityTypeLocation,
		newLocation.ID,
		map[string]interface{}{
			"name":         newLocation.Name,
			"world_id":     newLocation.WorldID.String(),
			"hierarchy_level": newLocation.HierarchyLevel,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("location created", "location_id", newLocation.ID, "name", newLocation.Name)
	uc.enqueueIngestion(ctx, input.TenantID, newLocation.ID)

	return &CreateLocationOutput{
		Location: newLocation,
	}, nil
}

func (uc *CreateLocationUseCase) enqueueIngestion(ctx context.Context, tenantID uuid.UUID, locationID uuid.UUID) {
	if uc.ingestionQueue == nil {
		return
	}
	if err := uc.ingestionQueue.Push(ctx, tenantID, "location", locationID); err != nil {
		uc.logger.Error("failed to enqueue location ingestion", "error", err, "location_id", locationID, "tenant_id", tenantID)
	}
}

