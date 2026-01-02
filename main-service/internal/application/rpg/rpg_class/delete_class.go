package rpg_class

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteRPGClassUseCase handles RPG class deletion
type DeleteRPGClassUseCase struct {
	classRepo repositories.RPGClassRepository
	logger    logger.Logger
}

// NewDeleteRPGClassUseCase creates a new DeleteRPGClassUseCase
func NewDeleteRPGClassUseCase(
	classRepo repositories.RPGClassRepository,
	logger logger.Logger,
) *DeleteRPGClassUseCase {
	return &DeleteRPGClassUseCase{
		classRepo: classRepo,
		logger:    logger,
	}
}

// DeleteRPGClassInput represents the input for deleting an RPG class
type DeleteRPGClassInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes an RPG class
func (uc *DeleteRPGClassUseCase) Execute(ctx context.Context, input DeleteRPGClassInput) error {
	if err := uc.classRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete RPG class", "error", err, "class_id", input.ID)
		return err
	}

	uc.logger.Info("RPG class deleted", "class_id", input.ID)

	return nil
}


