package archetype

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// RemoveTraitFromArchetypeUseCase handles removing a trait from an archetype
type RemoveTraitFromArchetypeUseCase struct {
	archetypeTraitRepo repositories.ArchetypeTraitRepository
	logger             logger.Logger
}

// NewRemoveTraitFromArchetypeUseCase creates a new RemoveTraitFromArchetypeUseCase
func NewRemoveTraitFromArchetypeUseCase(
	archetypeTraitRepo repositories.ArchetypeTraitRepository,
	logger logger.Logger,
) *RemoveTraitFromArchetypeUseCase {
	return &RemoveTraitFromArchetypeUseCase{
		archetypeTraitRepo: archetypeTraitRepo,
		logger:             logger,
	}
}

// RemoveTraitFromArchetypeInput represents the input for removing a trait from an archetype
type RemoveTraitFromArchetypeInput struct {
	ArchetypeID uuid.UUID
	TraitID     uuid.UUID
}

// Execute removes a trait from an archetype
func (uc *RemoveTraitFromArchetypeUseCase) Execute(ctx context.Context, input RemoveTraitFromArchetypeInput) error {
	if err := uc.archetypeTraitRepo.Delete(ctx, input.ArchetypeID, input.TraitID); err != nil {
		uc.logger.Error("failed to remove trait from archetype", "error", err, "archetype_id", input.ArchetypeID, "trait_id", input.TraitID)
		return err
	}

	uc.logger.Info("trait removed from archetype", "archetype_id", input.ArchetypeID, "trait_id", input.TraitID)

	return nil
}

