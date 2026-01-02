package rpg_system

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetRPGSystemUseCase handles retrieving an RPG system
type GetRPGSystemUseCase struct {
	rpgSystemRepo repositories.RPGSystemRepository
	logger         logger.Logger
}

// NewGetRPGSystemUseCase creates a new GetRPGSystemUseCase
func NewGetRPGSystemUseCase(
	rpgSystemRepo repositories.RPGSystemRepository,
	logger logger.Logger,
) *GetRPGSystemUseCase {
	return &GetRPGSystemUseCase{
		rpgSystemRepo: rpgSystemRepo,
		logger:         logger,
	}
}

// GetRPGSystemInput represents the input for getting an RPG system
type GetRPGSystemInput struct {
	ID uuid.UUID
}

// GetRPGSystemOutput represents the output of getting an RPG system
type GetRPGSystemOutput struct {
	RPGSystem *rpg.RPGSystem
}

// Execute retrieves an RPG system by ID
func (uc *GetRPGSystemUseCase) Execute(ctx context.Context, input GetRPGSystemInput) (*GetRPGSystemOutput, error) {
	system, err := uc.rpgSystemRepo.GetByID(ctx, input.ID)
	if err != nil {
		uc.logger.Error("failed to get RPG system", "error", err, "rpg_system_id", input.ID)
		return nil, err
	}

	return &GetRPGSystemOutput{
		RPGSystem: system,
	}, nil
}


