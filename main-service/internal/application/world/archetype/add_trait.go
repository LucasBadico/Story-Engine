package archetype

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddTraitToArchetypeUseCase handles adding a trait to an archetype
type AddTraitToArchetypeUseCase struct {
	archetypeRepo      repositories.ArchetypeRepository
	traitRepo          repositories.TraitRepository
	archetypeTraitRepo repositories.ArchetypeTraitRepository
	logger             logger.Logger
}

// NewAddTraitToArchetypeUseCase creates a new AddTraitToArchetypeUseCase
func NewAddTraitToArchetypeUseCase(
	archetypeRepo repositories.ArchetypeRepository,
	traitRepo repositories.TraitRepository,
	archetypeTraitRepo repositories.ArchetypeTraitRepository,
	logger logger.Logger,
) *AddTraitToArchetypeUseCase {
	return &AddTraitToArchetypeUseCase{
		archetypeRepo:      archetypeRepo,
		traitRepo:          traitRepo,
		archetypeTraitRepo: archetypeTraitRepo,
		logger:             logger,
	}
}

// AddTraitToArchetypeInput represents the input for adding a trait to an archetype
type AddTraitToArchetypeInput struct {
	TenantID     uuid.UUID
	ArchetypeID  uuid.UUID
	TraitID      uuid.UUID
	DefaultValue string
}

// Execute adds a trait to an archetype
func (uc *AddTraitToArchetypeUseCase) Execute(ctx context.Context, input AddTraitToArchetypeInput) error {
	_, err := uc.archetypeRepo.GetByID(ctx, input.TenantID, input.ArchetypeID)
	if err != nil {
		return err
	}

	_, err = uc.traitRepo.GetByID(ctx, input.TenantID, input.TraitID)
	if err != nil {
		return err
	}

	archetypeTrait := world.NewArchetypeTrait(input.ArchetypeID, input.TraitID, input.DefaultValue)

	if err := uc.archetypeTraitRepo.Create(ctx, archetypeTrait); err != nil {
		uc.logger.Error("failed to add trait to archetype", "error", err, "archetype_id", input.ArchetypeID, "trait_id", input.TraitID)
		return err
	}

	uc.logger.Info("trait added to archetype", "archetype_id", input.ArchetypeID, "trait_id", input.TraitID)

	return nil
}


