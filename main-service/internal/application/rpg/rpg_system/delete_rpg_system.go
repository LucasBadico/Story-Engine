package rpg_system

import (
	"context"

	"github.com/google/uuid"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteRPGSystemUseCase handles RPG system deletion
type DeleteRPGSystemUseCase struct {
	rpgSystemRepo repositories.RPGSystemRepository
	logger         logger.Logger
}

// NewDeleteRPGSystemUseCase creates a new DeleteRPGSystemUseCase
func NewDeleteRPGSystemUseCase(
	rpgSystemRepo repositories.RPGSystemRepository,
	logger logger.Logger,
) *DeleteRPGSystemUseCase {
	return &DeleteRPGSystemUseCase{
		rpgSystemRepo: rpgSystemRepo,
		logger:         logger,
	}
}

// DeleteRPGSystemInput represents the input for deleting an RPG system
type DeleteRPGSystemInput struct {
	ID uuid.UUID
}

// Execute deletes an RPG system
func (uc *DeleteRPGSystemUseCase) Execute(ctx context.Context, input DeleteRPGSystemInput) error {
	// Get existing system to check if it's builtin
	system, err := uc.rpgSystemRepo.GetByID(ctx, input.ID)
	if err != nil {
		return err
	}

	// Cannot delete builtin systems
	if system.IsBuiltin {
		return &platformerrors.ValidationError{
			Field:   "id",
			Message: "cannot delete builtin RPG systems",
		}
	}

	if err := uc.rpgSystemRepo.Delete(ctx, input.ID); err != nil {
		uc.logger.Error("failed to delete RPG system", "error", err, "rpg_system_id", input.ID)
		return err
	}

	uc.logger.Info("RPG system deleted", "rpg_system_id", input.ID)

	return nil
}

