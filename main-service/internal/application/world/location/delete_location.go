package location

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteLocationUseCase handles location deletion
type DeleteLocationUseCase struct {
	locationRepo repositories.LocationRepository
	auditLogRepo repositories.AuditLogRepository
	logger       logger.Logger
}

// NewDeleteLocationUseCase creates a new DeleteLocationUseCase
func NewDeleteLocationUseCase(
	locationRepo repositories.LocationRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *DeleteLocationUseCase {
	return &DeleteLocationUseCase{
		locationRepo: locationRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// DeleteLocationInput represents the input for deleting a location
type DeleteLocationInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes a location
func (uc *DeleteLocationUseCase) Execute(ctx context.Context, input DeleteLocationInput) error {
	l, err := uc.locationRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	if err := uc.locationRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete location", "error", err, "location_id", input.ID)
		return err
	}

	auditLog := audit.NewAuditLog(
		l.TenantID,
		nil,
		audit.ActionDelete,
		audit.EntityTypeLocation,
		l.ID,
		map[string]interface{}{
			"name": l.Name,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("location deleted", "location_id", input.ID, "name", l.Name)

	return nil
}


