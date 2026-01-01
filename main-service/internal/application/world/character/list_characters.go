package character

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListCharactersUseCase handles listing characters
type ListCharactersUseCase struct {
	characterRepo repositories.CharacterRepository
	logger        logger.Logger
}

// NewListCharactersUseCase creates a new ListCharactersUseCase
func NewListCharactersUseCase(
	characterRepo repositories.CharacterRepository,
	logger logger.Logger,
) *ListCharactersUseCase {
	return &ListCharactersUseCase{
		characterRepo: characterRepo,
		logger:        logger,
	}
}

// ListCharactersInput represents the input for listing characters
type ListCharactersInput struct {
	WorldID uuid.UUID
	Limit   int
	Offset  int
}

// ListCharactersOutput represents the output of listing characters
type ListCharactersOutput struct {
	Characters []*world.Character
	Total      int
}

// Execute lists characters for a world
func (uc *ListCharactersUseCase) Execute(ctx context.Context, input ListCharactersInput) (*ListCharactersOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	characters, err := uc.characterRepo.ListByWorld(ctx, input.WorldID, limit, input.Offset)
	if err != nil {
		uc.logger.Error("failed to list characters", "error", err, "world_id", input.WorldID)
		return nil, err
	}

	total, err := uc.characterRepo.CountByWorld(ctx, input.WorldID)
	if err != nil {
		uc.logger.Warn("failed to count characters", "error", err)
		total = len(characters)
	}

	return &ListCharactersOutput{
		Characters: characters,
		Total:      total,
	}, nil
}

