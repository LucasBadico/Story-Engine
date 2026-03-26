package character_state

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListCharacterStatesUseCase handles listing character states
type ListCharacterStatesUseCase struct {
	characterStateRepo repositories.CharacterStateRepository
	logger             logger.Logger
}

// NewListCharacterStatesUseCase creates a new ListCharacterStatesUseCase
func NewListCharacterStatesUseCase(
	characterStateRepo repositories.CharacterStateRepository,
	logger logger.Logger,
) *ListCharacterStatesUseCase {
	return &ListCharacterStatesUseCase{
		characterStateRepo: characterStateRepo,
		logger:             logger,
	}
}

// ListCharacterStatesInput represents the input for listing character states
type ListCharacterStatesInput struct {
	TenantID  uuid.UUID
	ChapterID uuid.UUID
}

// ListCharacterStatesOutput represents the output of listing character states
type ListCharacterStatesOutput struct {
	CharacterStates []*world.CharacterState
}

// Execute lists character states for a chapter
func (uc *ListCharacterStatesUseCase) Execute(ctx context.Context, input ListCharacterStatesInput) (*ListCharacterStatesOutput, error) {
	states, err := uc.characterStateRepo.ListByChapter(ctx, input.TenantID, input.ChapterID)
	if err != nil {
		uc.logger.Error("failed to list character states", "error", err, "chapter_id", input.ChapterID)
		return nil, err
	}

	return &ListCharacterStatesOutput{
		CharacterStates: states,
	}, nil
}
