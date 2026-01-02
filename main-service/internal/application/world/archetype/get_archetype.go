package archetype

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetArchetypeUseCase handles archetype retrieval
type GetArchetypeUseCase struct {
	archetypeRepo repositories.ArchetypeRepository
	logger        logger.Logger
}

// NewGetArchetypeUseCase creates a new GetArchetypeUseCase
func NewGetArchetypeUseCase(
	archetypeRepo repositories.ArchetypeRepository,
	logger logger.Logger,
) *GetArchetypeUseCase {
	return &GetArchetypeUseCase{
		archetypeRepo: archetypeRepo,
		logger:        logger,
	}
}

// GetArchetypeInput represents the input for getting an archetype
type GetArchetypeInput struct {
	ID uuid.UUID
}

// GetArchetypeOutput represents the output of getting an archetype
type GetArchetypeOutput struct {
	Archetype *world.Archetype
}

// Execute retrieves an archetype by ID
func (uc *GetArchetypeUseCase) Execute(ctx context.Context, input GetArchetypeInput) (*GetArchetypeOutput, error) {
	a, err := uc.archetypeRepo.GetByID(ctx, input.ID)
	if err != nil {
		uc.logger.Error("failed to get archetype", "error", err, "archetype_id", input.ID)
		return nil, err
	}

	return &GetArchetypeOutput{
		Archetype: a,
	}, nil
}


