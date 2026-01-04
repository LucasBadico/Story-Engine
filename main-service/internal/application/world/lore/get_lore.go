package lore

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetLoreUseCase handles retrieving a lore
type GetLoreUseCase struct {
	loreRepo repositories.LoreRepository
	logger   logger.Logger
}

// NewGetLoreUseCase creates a new GetLoreUseCase
func NewGetLoreUseCase(
	loreRepo repositories.LoreRepository,
	logger logger.Logger,
) *GetLoreUseCase {
	return &GetLoreUseCase{
		loreRepo: loreRepo,
		logger:   logger,
	}
}

// GetLoreInput represents the input for getting a lore
type GetLoreInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetLoreOutput represents the output of getting a lore
type GetLoreOutput struct {
	Lore *world.Lore
}

// Execute retrieves a lore by ID
func (uc *GetLoreUseCase) Execute(ctx context.Context, input GetLoreInput) (*GetLoreOutput, error) {
	lore, err := uc.loreRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get lore", "error", err, "lore_id", input.ID)
		return nil, err
	}

	return &GetLoreOutput{
		Lore: lore,
	}, nil
}

