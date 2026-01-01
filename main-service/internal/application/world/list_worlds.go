package world

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListWorldsUseCase handles listing worlds
type ListWorldsUseCase struct {
	worldRepo repositories.WorldRepository
	logger    logger.Logger
}

// NewListWorldsUseCase creates a new ListWorldsUseCase
func NewListWorldsUseCase(
	worldRepo repositories.WorldRepository,
	logger logger.Logger,
) *ListWorldsUseCase {
	return &ListWorldsUseCase{
		worldRepo: worldRepo,
		logger:    logger,
	}
}

// ListWorldsInput represents the input for listing worlds
type ListWorldsInput struct {
	TenantID uuid.UUID
	Limit    int
	Offset   int
}

// ListWorldsOutput represents the output of listing worlds
type ListWorldsOutput struct {
	Worlds []*world.World
	Total  int
}

// Execute lists worlds for a tenant
func (uc *ListWorldsUseCase) Execute(ctx context.Context, input ListWorldsInput) (*ListWorldsOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	worlds, err := uc.worldRepo.ListByTenant(ctx, input.TenantID, limit, input.Offset)
	if err != nil {
		uc.logger.Error("failed to list worlds", "error", err, "tenant_id", input.TenantID)
		return nil, err
	}

	total, err := uc.worldRepo.CountByTenant(ctx, input.TenantID)
	if err != nil {
		uc.logger.Warn("failed to count worlds", "error", err)
		total = len(worlds)
	}

	return &ListWorldsOutput{
		Worlds: worlds,
		Total:  total,
	}, nil
}

