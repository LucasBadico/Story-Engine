package rpg_system

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateRPGSystemUseCase handles RPG system updates
type UpdateRPGSystemUseCase struct {
	rpgSystemRepo repositories.RPGSystemRepository
	logger         logger.Logger
}

// NewUpdateRPGSystemUseCase creates a new UpdateRPGSystemUseCase
func NewUpdateRPGSystemUseCase(
	rpgSystemRepo repositories.RPGSystemRepository,
	logger logger.Logger,
) *UpdateRPGSystemUseCase {
	return &UpdateRPGSystemUseCase{
		rpgSystemRepo: rpgSystemRepo,
		logger:         logger,
	}
}

// UpdateRPGSystemInput represents the input for updating an RPG system
type UpdateRPGSystemInput struct {
	ID                uuid.UUID
	Name              *string
	Description       *string
	BaseStatsSchema   *json.RawMessage
	DerivedStatsSchema *json.RawMessage
	ProgressionSchema *json.RawMessage
}

// UpdateRPGSystemOutput represents the output of updating an RPG system
type UpdateRPGSystemOutput struct {
	RPGSystem *rpg.RPGSystem
}

// Execute updates an RPG system
func (uc *UpdateRPGSystemUseCase) Execute(ctx context.Context, input UpdateRPGSystemInput) (*UpdateRPGSystemOutput, error) {
	// Get existing system
	system, err := uc.rpgSystemRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// Cannot update builtin systems
	if system.IsBuiltin {
		return nil, &platformerrors.ValidationError{
			Field:   "id",
			Message: "cannot update builtin RPG systems",
		}
	}

	// Update fields if provided
	if input.Name != nil {
		if err := system.UpdateName(*input.Name); err != nil {
			return nil, err
		}
	}
	if input.Description != nil {
		system.UpdateDescription(input.Description)
	}
	if input.BaseStatsSchema != nil || input.DerivedStatsSchema != nil || input.ProgressionSchema != nil {
		baseSchema := system.BaseStatsSchema
		if input.BaseStatsSchema != nil {
			baseSchema = *input.BaseStatsSchema
		}
		if err := system.UpdateSchemas(baseSchema, input.DerivedStatsSchema, input.ProgressionSchema); err != nil {
			return nil, err
		}
	}

	if err := system.Validate(); err != nil {
		return nil, err
	}

	if err := uc.rpgSystemRepo.Update(ctx, system); err != nil {
		uc.logger.Error("failed to update RPG system", "error", err, "rpg_system_id", input.ID)
		return nil, err
	}

	uc.logger.Info("RPG system updated", "rpg_system_id", input.ID)

	return &UpdateRPGSystemOutput{
		RPGSystem: system,
	}, nil
}

