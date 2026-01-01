package artifact

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteArtifactUseCase handles artifact deletion
type DeleteArtifactUseCase struct {
	artifactRepo repositories.ArtifactRepository
	worldRepo repositories.WorldRepository
	auditLogRepo repositories.AuditLogRepository
	logger       logger.Logger
}

// NewDeleteArtifactUseCase creates a new DeleteArtifactUseCase
func NewDeleteArtifactUseCase(
	artifactRepo repositories.ArtifactRepository,
	worldRepo repositories.WorldRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *DeleteArtifactUseCase {
	return &DeleteArtifactUseCase{
		artifactRepo: artifactRepo,
		worldRepo: worldRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// DeleteArtifactInput represents the input for deleting an artifact
type DeleteArtifactInput struct {
	ID uuid.UUID
}

// Execute deletes an artifact
func (uc *DeleteArtifactUseCase) Execute(ctx context.Context, input DeleteArtifactInput) error {
	a, err := uc.artifactRepo.GetByID(ctx, input.ID)
	if err != nil {
		return err
	}

	if err := uc.artifactRepo.Delete(ctx, input.ID); err != nil {
		uc.logger.Error("failed to delete artifact", "error", err, "artifact_id", input.ID)
		return err
	}

	w, _ := uc.worldRepo.GetByID(ctx, a.WorldID)
	auditLog := audit.NewAuditLog(
		w.TenantID,
		nil,
		audit.ActionDelete,
		audit.EntityTypeArtifact,
		a.ID,
		map[string]interface{}{
			"name": a.Name,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("artifact deleted", "artifact_id", input.ID, "name", a.Name)

	return nil
}

