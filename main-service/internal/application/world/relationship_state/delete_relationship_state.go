package relationship_state

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteRelationshipStateUseCase handles relationship state deletion
type DeleteRelationshipStateUseCase struct {
	relationshipStateRepo repositories.RelationshipStateRepository
	auditLogRepo          repositories.AuditLogRepository
	logger                logger.Logger
}

// NewDeleteRelationshipStateUseCase creates a new DeleteRelationshipStateUseCase
func NewDeleteRelationshipStateUseCase(
	relationshipStateRepo repositories.RelationshipStateRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *DeleteRelationshipStateUseCase {
	return &DeleteRelationshipStateUseCase{
		relationshipStateRepo: relationshipStateRepo,
		auditLogRepo:          auditLogRepo,
		logger:                logger,
	}
}

// DeleteRelationshipStateInput represents the input for deleting a relationship state
type DeleteRelationshipStateInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes a relationship state
func (uc *DeleteRelationshipStateUseCase) Execute(ctx context.Context, input DeleteRelationshipStateInput) error {
	rs, err := uc.relationshipStateRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	if err := uc.relationshipStateRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete relationship state", "error", err, "relationship_state_id", input.ID)
		return err
	}

	auditLog := audit.NewAuditLog(
		input.TenantID,
		nil,
		audit.ActionDelete,
		audit.EntityTypeRelationshipState,
		rs.ID,
		map[string]interface{}{
			"chapter_number": rs.ChapterNumber,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("relationship state deleted", "relationship_state_id", input.ID)

	return nil
}
