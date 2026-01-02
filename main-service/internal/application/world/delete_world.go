package world

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteWorldUseCase handles world deletion
type DeleteWorldUseCase struct {
	worldRepo    repositories.WorldRepository
	auditLogRepo repositories.AuditLogRepository
	logger       logger.Logger
}

// NewDeleteWorldUseCase creates a new DeleteWorldUseCase
func NewDeleteWorldUseCase(
	worldRepo repositories.WorldRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *DeleteWorldUseCase {
	return &DeleteWorldUseCase{
		worldRepo:    worldRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// DeleteWorldInput represents the input for deleting a world
type DeleteWorldInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes a world
func (uc *DeleteWorldUseCase) Execute(ctx context.Context, input DeleteWorldInput) error {
	w, err := uc.worldRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	if err := uc.worldRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete world", "error", err, "world_id", input.ID)
		return err
	}

	// Log audit event
	auditLog := audit.NewAuditLog(
		w.TenantID,
		nil,
		audit.ActionDelete,
		audit.EntityTypeWorld,
		w.ID,
		map[string]interface{}{
			"name": w.Name,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("world deleted", "world_id", input.ID, "name", w.Name)

	return nil
}


