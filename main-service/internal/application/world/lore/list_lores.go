package lore

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListLoresUseCase handles listing lores for a world
type ListLoresUseCase struct {
	loreRepo repositories.LoreRepository
	logger   logger.Logger
}

// NewListLoresUseCase creates a new ListLoresUseCase
func NewListLoresUseCase(
	loreRepo repositories.LoreRepository,
	logger logger.Logger,
) *ListLoresUseCase {
	return &ListLoresUseCase{
		loreRepo: loreRepo,
		logger:   logger,
	}
}

// ListLoresInput represents the input for listing lores
type ListLoresInput struct {
	TenantID uuid.UUID
	WorldID  uuid.UUID
}

// ListLoresOutput represents the output of listing lores
type ListLoresOutput struct {
	Lores []*world.Lore
}

// Execute lists lores for a world
func (uc *ListLoresUseCase) Execute(ctx context.Context, input ListLoresInput) (*ListLoresOutput, error) {
	lores, err := uc.loreRepo.ListByWorld(ctx, input.TenantID, input.WorldID)
	if err != nil {
		uc.logger.Error("failed to list lores", "error", err, "world_id", input.WorldID)
		return nil, err
	}

	return &ListLoresOutput{
		Lores: lores,
	}, nil
}

