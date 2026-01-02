package archetype

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListArchetypesUseCase handles listing archetypes
type ListArchetypesUseCase struct {
	archetypeRepo repositories.ArchetypeRepository
	logger        logger.Logger
}

// NewListArchetypesUseCase creates a new ListArchetypesUseCase
func NewListArchetypesUseCase(
	archetypeRepo repositories.ArchetypeRepository,
	logger logger.Logger,
) *ListArchetypesUseCase {
	return &ListArchetypesUseCase{
		archetypeRepo: archetypeRepo,
		logger:        logger,
	}
}

// ListArchetypesInput represents the input for listing archetypes
type ListArchetypesInput struct {
	TenantID uuid.UUID
	Limit    int
	Offset   int
}

// ListArchetypesOutput represents the output of listing archetypes
type ListArchetypesOutput struct {
	Archetypes []*world.Archetype
	Total      int
}

// Execute lists archetypes for a tenant
func (uc *ListArchetypesUseCase) Execute(ctx context.Context, input ListArchetypesInput) (*ListArchetypesOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	archetypes, err := uc.archetypeRepo.ListByTenant(ctx, input.TenantID, limit, input.Offset)
	if err != nil {
		uc.logger.Error("failed to list archetypes", "error", err, "tenant_id", input.TenantID)
		return nil, err
	}

	total, err := uc.archetypeRepo.CountByTenant(ctx, input.TenantID)
	if err != nil {
		uc.logger.Warn("failed to count archetypes", "error", err)
		total = len(archetypes)
	}

	return &ListArchetypesOutput{
		Archetypes: archetypes,
		Total:      total,
	}, nil
}


