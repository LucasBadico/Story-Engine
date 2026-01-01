package trait

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteTraitUseCase handles trait deletion
type DeleteTraitUseCase struct {
	traitRepo    repositories.TraitRepository
	auditLogRepo repositories.AuditLogRepository
	logger       logger.Logger
}

// NewDeleteTraitUseCase creates a new DeleteTraitUseCase
func NewDeleteTraitUseCase(
	traitRepo repositories.TraitRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *DeleteTraitUseCase {
	return &DeleteTraitUseCase{
		traitRepo:    traitRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// DeleteTraitInput represents the input for deleting a trait
type DeleteTraitInput struct {
	ID uuid.UUID
}

// Execute deletes a trait
func (uc *DeleteTraitUseCase) Execute(ctx context.Context, input DeleteTraitInput) error {
	t, err := uc.traitRepo.GetByID(ctx, input.ID)
	if err != nil {
		return err
	}

	if err := uc.traitRepo.Delete(ctx, input.ID); err != nil {
		uc.logger.Error("failed to delete trait", "error", err, "trait_id", input.ID)
		return err
	}

	auditLog := audit.NewAuditLog(
		t.TenantID,
		nil,
		audit.ActionDelete,
		audit.EntityTypeTrait,
		t.ID,
		map[string]interface{}{
			"name": t.Name,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("trait deleted", "trait_id", input.ID, "name", t.Name)

	return nil
}

