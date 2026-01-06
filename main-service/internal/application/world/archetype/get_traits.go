package archetype

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetArchetypeTraitsUseCase handles getting traits for an archetype
type GetArchetypeTraitsUseCase struct {
	archetypeTraitRepo repositories.ArchetypeTraitRepository
	logger             logger.Logger
}

// NewGetArchetypeTraitsUseCase creates a new GetArchetypeTraitsUseCase
func NewGetArchetypeTraitsUseCase(
	archetypeTraitRepo repositories.ArchetypeTraitRepository,
	logger logger.Logger,
) *GetArchetypeTraitsUseCase {
	return &GetArchetypeTraitsUseCase{
		archetypeTraitRepo: archetypeTraitRepo,
		logger:             logger,
	}
}

// GetArchetypeTraitsInput represents the input for getting archetype traits
type GetArchetypeTraitsInput struct {
	TenantID    uuid.UUID
	ArchetypeID uuid.UUID
}

// GetArchetypeTraitsOutput represents the output of getting archetype traits
type GetArchetypeTraitsOutput struct {
	Traits []*world.ArchetypeTrait
}

// Execute retrieves all traits for an archetype
func (uc *GetArchetypeTraitsUseCase) Execute(ctx context.Context, input GetArchetypeTraitsInput) (*GetArchetypeTraitsOutput, error) {
	traits, err := uc.archetypeTraitRepo.GetByArchetype(ctx, input.TenantID, input.ArchetypeID)
	if err != nil {
		uc.logger.Error("failed to get archetype traits", "error", err, "archetype_id", input.ArchetypeID)
		return nil, err
	}

	return &GetArchetypeTraitsOutput{
		Traits: traits,
	}, nil
}

