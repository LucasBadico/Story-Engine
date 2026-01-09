package lore

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteLoreUseCase handles lore deletion
type DeleteLoreUseCase struct {
	loreRepo     repositories.LoreRepository
	relationRepo repositories.EntityRelationRepository
	auditLogRepo repositories.AuditLogRepository
	logger       logger.Logger
}

// NewDeleteLoreUseCase creates a new DeleteLoreUseCase
func NewDeleteLoreUseCase(
	loreRepo repositories.LoreRepository,
	relationRepo repositories.EntityRelationRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *DeleteLoreUseCase {
	return &DeleteLoreUseCase{
		loreRepo:     loreRepo,
		relationRepo: relationRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// DeleteLoreInput represents the input for deleting a lore
type DeleteLoreInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes a lore
func (uc *DeleteLoreUseCase) Execute(ctx context.Context, input DeleteLoreInput) error {
	// Get lore to get tenant_id for audit log
	lore, err := uc.loreRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	// Delete relations where lore is source or target
	if err := uc.relationRepo.DeleteByEntity(ctx, input.TenantID, "lore", input.ID); err != nil {
		uc.logger.Error("failed to delete lore relations", "error", err)
		// Continue anyway
	}

	// Delete lore
	if err := uc.loreRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete lore", "error", err, "lore_id", input.ID)
		return err
	}

	// Log audit event
	auditLog := audit.NewAuditLog(
		lore.TenantID,
		nil,
		audit.ActionDelete,
		audit.EntityTypeLore,
		input.ID,
		map[string]interface{}{},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Error("failed to create audit log", "error", err)
	}

	uc.logger.Info("lore deleted", "lore_id", input.ID)

	return nil
}
