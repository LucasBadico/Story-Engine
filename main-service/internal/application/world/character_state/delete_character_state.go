package character_state

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteCharacterStateUseCase handles character state deletion
type DeleteCharacterStateUseCase struct {
	characterStateRepo repositories.CharacterStateRepository
	auditLogRepo       repositories.AuditLogRepository
	logger             logger.Logger
}

// NewDeleteCharacterStateUseCase creates a new DeleteCharacterStateUseCase
func NewDeleteCharacterStateUseCase(
	characterStateRepo repositories.CharacterStateRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *DeleteCharacterStateUseCase {
	return &DeleteCharacterStateUseCase{
		characterStateRepo: characterStateRepo,
		auditLogRepo:       auditLogRepo,
		logger:             logger,
	}
}

// DeleteCharacterStateInput represents the input for deleting a character state
type DeleteCharacterStateInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes a character state
func (uc *DeleteCharacterStateUseCase) Execute(ctx context.Context, input DeleteCharacterStateInput) error {
	cs, err := uc.characterStateRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	if err := uc.characterStateRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete character state", "error", err, "character_state_id", input.ID)
		return err
	}

	auditLog := audit.NewAuditLog(
		input.TenantID,
		nil,
		audit.ActionDelete,
		audit.EntityTypeCharacterState,
		cs.ID,
		map[string]interface{}{
			"character_id":   cs.CharacterID.String(),
			"chapter_number": cs.ChapterNumber,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("character state deleted", "character_state_id", input.ID)

	return nil
}
