package archetype

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteArchetypeUseCase handles archetype deletion
type DeleteArchetypeUseCase struct {
	archetypeRepo        repositories.ArchetypeRepository
	archetypeTraitRepo   repositories.ArchetypeTraitRepository
	auditLogRepo         repositories.AuditLogRepository
	logger               logger.Logger
}

// NewDeleteArchetypeUseCase creates a new DeleteArchetypeUseCase
func NewDeleteArchetypeUseCase(
	archetypeRepo repositories.ArchetypeRepository,
	archetypeTraitRepo repositories.ArchetypeTraitRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *DeleteArchetypeUseCase {
	return &DeleteArchetypeUseCase{
		archetypeRepo:      archetypeRepo,
		archetypeTraitRepo: archetypeTraitRepo,
		auditLogRepo:       auditLogRepo,
		logger:             logger,
	}
}

// DeleteArchetypeInput represents the input for deleting an archetype
type DeleteArchetypeInput struct {
	ID uuid.UUID
}

// Execute deletes an archetype
func (uc *DeleteArchetypeUseCase) Execute(ctx context.Context, input DeleteArchetypeInput) error {
	a, err := uc.archetypeRepo.GetByID(ctx, input.ID)
	if err != nil {
		return err
	}

	// Delete all archetype-trait relationships first (CASCADE should handle this, but being explicit)
	if err := uc.archetypeTraitRepo.DeleteByArchetype(ctx, input.ID); err != nil {
		uc.logger.Warn("failed to delete archetype traits", "error", err)
	}

	if err := uc.archetypeRepo.Delete(ctx, input.ID); err != nil {
		uc.logger.Error("failed to delete archetype", "error", err, "archetype_id", input.ID)
		return err
	}

	auditLog := audit.NewAuditLog(
		a.TenantID,
		nil,
		audit.ActionDelete,
		audit.EntityTypeArchetype,
		a.ID,
		map[string]interface{}{
			"name": a.Name,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("archetype deleted", "archetype_id", input.ID, "name", a.Name)

	return nil
}

