package rpg_system

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListRPGSystemsUseCase handles listing RPG systems
type ListRPGSystemsUseCase struct {
	rpgSystemRepo repositories.RPGSystemRepository
	logger         logger.Logger
}

// NewListRPGSystemsUseCase creates a new ListRPGSystemsUseCase
func NewListRPGSystemsUseCase(
	rpgSystemRepo repositories.RPGSystemRepository,
	logger logger.Logger,
) *ListRPGSystemsUseCase {
	return &ListRPGSystemsUseCase{
		rpgSystemRepo: rpgSystemRepo,
		logger:         logger,
	}
}

// ListRPGSystemsInput represents the input for listing RPG systems
type ListRPGSystemsInput struct {
	TenantID *uuid.UUID // nil = builtin only
}

// ListRPGSystemsOutput represents the output of listing RPG systems
type ListRPGSystemsOutput struct {
	RPGSystems []*rpg.RPGSystem
}

// Execute lists RPG systems (builtin + tenant custom if tenantID provided)
func (uc *ListRPGSystemsUseCase) Execute(ctx context.Context, input ListRPGSystemsInput) (*ListRPGSystemsOutput, error) {
	systems, err := uc.rpgSystemRepo.List(ctx, input.TenantID)
	if err != nil {
		uc.logger.Error("failed to list RPG systems", "error", err)
		return nil, err
	}

	return &ListRPGSystemsOutput{
		RPGSystems: systems,
	}, nil
}

