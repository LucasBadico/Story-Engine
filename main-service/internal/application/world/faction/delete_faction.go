package faction

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteFactionUseCase handles faction deletion
type DeleteFactionUseCase struct {
	factionRepo  repositories.FactionRepository
	relationRepo repositories.EntityRelationRepository
	auditLogRepo repositories.AuditLogRepository
	logger       logger.Logger
}

// NewDeleteFactionUseCase creates a new DeleteFactionUseCase
func NewDeleteFactionUseCase(
	factionRepo repositories.FactionRepository,
	relationRepo repositories.EntityRelationRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *DeleteFactionUseCase {
	return &DeleteFactionUseCase{
		factionRepo:  factionRepo,
		relationRepo: relationRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// DeleteFactionInput represents the input for deleting a faction
type DeleteFactionInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes a faction
func (uc *DeleteFactionUseCase) Execute(ctx context.Context, input DeleteFactionInput) error {
	// Get faction to get tenant_id for audit log
	faction, err := uc.factionRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	// Delete relations where faction is source or target
	if err := uc.relationRepo.DeleteByEntity(ctx, input.TenantID, "faction", input.ID); err != nil {
		uc.logger.Error("failed to delete faction relations", "error", err)
		// Continue anyway
	}

	// Delete faction
	if err := uc.factionRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete faction", "error", err, "faction_id", input.ID)
		return err
	}

	// Log audit event
	auditLog := audit.NewAuditLog(
		faction.TenantID,
		nil,
		audit.ActionDelete,
		audit.EntityTypeFaction,
		input.ID,
		map[string]interface{}{},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Error("failed to create audit log", "error", err)
	}

	uc.logger.Info("faction deleted", "faction_id", input.ID)

	return nil
}
