package story

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetStoryUseCase handles retrieving a story
type GetStoryUseCase struct {
	storyRepo repositories.StoryRepository
	logger    logger.Logger
}

// NewGetStoryUseCase creates a new GetStoryUseCase
func NewGetStoryUseCase(
	storyRepo repositories.StoryRepository,
	logger logger.Logger,
) *GetStoryUseCase {
	return &GetStoryUseCase{
		storyRepo: storyRepo,
		logger:    logger,
	}
}

// GetStoryInput represents the input for getting a story
type GetStoryInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetStoryOutput represents the output of getting a story
type GetStoryOutput struct {
	Story *story.Story
}

// Execute retrieves a story by ID
func (uc *GetStoryUseCase) Execute(ctx context.Context, input GetStoryInput) (*GetStoryOutput, error) {
	s, err := uc.storyRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get story", "error", err, "story_id", input.ID, "tenant_id", input.TenantID)
		return nil, err
	}

	return &GetStoryOutput{
		Story: s,
	}, nil
}

