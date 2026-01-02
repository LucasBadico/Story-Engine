package character

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteCharacterUseCase handles character deletion
type DeleteCharacterUseCase struct {
	characterRepo      repositories.CharacterRepository
	characterTraitRepo repositories.CharacterTraitRepository
	worldRepo          repositories.WorldRepository
	auditLogRepo       repositories.AuditLogRepository
	logger             logger.Logger
}

// NewDeleteCharacterUseCase creates a new DeleteCharacterUseCase
func NewDeleteCharacterUseCase(
	characterRepo repositories.CharacterRepository,
	characterTraitRepo repositories.CharacterTraitRepository,
	worldRepo repositories.WorldRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *DeleteCharacterUseCase {
	return &DeleteCharacterUseCase{
		characterRepo:      characterRepo,
		characterTraitRepo: characterTraitRepo,
		worldRepo:          worldRepo,
		auditLogRepo:       auditLogRepo,
		logger:             logger,
	}
}

// DeleteCharacterInput represents the input for deleting a character
type DeleteCharacterInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes a character
func (uc *DeleteCharacterUseCase) Execute(ctx context.Context, input DeleteCharacterInput) error {
	c, err := uc.characterRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	// Delete all character traits first (CASCADE should handle this, but being explicit)
	if err := uc.characterTraitRepo.DeleteByCharacter(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Warn("failed to delete character traits", "error", err)
	}

	if err := uc.characterRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete character", "error", err, "character_id", input.ID)
		return err
	}

	w, _ := uc.worldRepo.GetByID(ctx, input.TenantID, c.WorldID)
	auditLog := audit.NewAuditLog(
		w.TenantID,
		nil,
		audit.ActionDelete,
		audit.EntityTypeCharacter,
		c.ID,
		map[string]interface{}{
			"name": c.Name,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("character deleted", "character_id", input.ID, "name", c.Name)

	return nil
}


