package world

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetWorldUseCase handles world retrieval
type GetWorldUseCase struct {
	worldRepo repositories.WorldRepository
	logger    logger.Logger
}

// NewGetWorldUseCase creates a new GetWorldUseCase
func NewGetWorldUseCase(
	worldRepo repositories.WorldRepository,
	logger logger.Logger,
) *GetWorldUseCase {
	return &GetWorldUseCase{
		worldRepo: worldRepo,
		logger:    logger,
	}
}

// GetWorldInput represents the input for getting a world
type GetWorldInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetWorldOutput represents the output of getting a world
type GetWorldOutput struct {
	World *world.World
}

// Execute retrieves a world by ID
func (uc *GetWorldUseCase) Execute(ctx context.Context, input GetWorldInput) (*GetWorldOutput, error) {
	w, err := uc.worldRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get world", "error", err, "world_id", input.ID)
		return nil, err
	}

	return &GetWorldOutput{
		World: w,
	}, nil
}


