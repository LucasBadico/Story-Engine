package trait

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetTraitUseCase handles trait retrieval
type GetTraitUseCase struct {
	traitRepo repositories.TraitRepository
	logger    logger.Logger
}

// NewGetTraitUseCase creates a new GetTraitUseCase
func NewGetTraitUseCase(
	traitRepo repositories.TraitRepository,
	logger logger.Logger,
) *GetTraitUseCase {
	return &GetTraitUseCase{
		traitRepo: traitRepo,
		logger:    logger,
	}
}

// GetTraitInput represents the input for getting a trait
type GetTraitInput struct {
	ID uuid.UUID
}

// GetTraitOutput represents the output of getting a trait
type GetTraitOutput struct {
	Trait *world.Trait
}

// Execute retrieves a trait by ID
func (uc *GetTraitUseCase) Execute(ctx context.Context, input GetTraitInput) (*GetTraitOutput, error) {
	t, err := uc.traitRepo.GetByID(ctx, input.ID)
	if err != nil {
		uc.logger.Error("failed to get trait", "error", err, "trait_id", input.ID)
		return nil, err
	}

	return &GetTraitOutput{
		Trait: t,
	}, nil
}


