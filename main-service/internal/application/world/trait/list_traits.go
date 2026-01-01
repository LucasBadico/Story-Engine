package trait

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListTraitsUseCase handles listing traits
type ListTraitsUseCase struct {
	traitRepo repositories.TraitRepository
	logger    logger.Logger
}

// NewListTraitsUseCase creates a new ListTraitsUseCase
func NewListTraitsUseCase(
	traitRepo repositories.TraitRepository,
	logger logger.Logger,
) *ListTraitsUseCase {
	return &ListTraitsUseCase{
		traitRepo: traitRepo,
		logger:    logger,
	}
}

// ListTraitsInput represents the input for listing traits
type ListTraitsInput struct {
	TenantID uuid.UUID
	Limit    int
	Offset   int
}

// ListTraitsOutput represents the output of listing traits
type ListTraitsOutput struct {
	Traits []*world.Trait
	Total  int
}

// Execute lists traits for a tenant
func (uc *ListTraitsUseCase) Execute(ctx context.Context, input ListTraitsInput) (*ListTraitsOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	traits, err := uc.traitRepo.ListByTenant(ctx, input.TenantID, limit, input.Offset)
	if err != nil {
		uc.logger.Error("failed to list traits", "error", err, "tenant_id", input.TenantID)
		return nil, err
	}

	total, err := uc.traitRepo.CountByTenant(ctx, input.TenantID)
	if err != nil {
		uc.logger.Warn("failed to count traits", "error", err)
		total = len(traits)
	}

	return &ListTraitsOutput{
		Traits: traits,
		Total:  total,
	}, nil
}

