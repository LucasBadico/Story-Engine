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
	relationRepo repositories.EntityRelationRepository
	worldRepo    repositories.WorldRepository
	auditLogRepo repositories.AuditLogRepository
	logger       logger.Logger
}

// NewDeleteArtifactUseCase creates a new DeleteArtifactUseCase
func NewDeleteArtifactUseCase(
	artifactRepo repositories.ArtifactRepository,
	relationRepo repositories.EntityRelationRepository,
	worldRepo repositories.WorldRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *DeleteArtifactUseCase {
	return &DeleteArtifactUseCase{
		artifactRepo: artifactRepo,
		relationRepo: relationRepo,
		worldRepo:    worldRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// DeleteArtifactInput represents the input for deleting an artifact
type DeleteArtifactInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes an artifact
func (uc *DeleteArtifactUseCase) Execute(ctx context.Context, input DeleteArtifactInput) error {
	a, err := uc.artifactRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	// Delete relations where artifact is source or target
	if err := uc.relationRepo.DeleteByEntity(ctx, input.TenantID, "artifact", input.ID); err != nil {
		uc.logger.Warn("failed to delete artifact relations", "error", err)
	}

	if err := uc.artifactRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete artifact", "error", err, "artifact_id", input.ID)
		return err
	}

	w, _ := uc.worldRepo.GetByID(ctx, input.TenantID, a.WorldID)
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
