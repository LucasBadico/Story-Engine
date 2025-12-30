package story

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetStoryVersionGraphUseCase handles retrieving story version graphs
type GetStoryVersionGraphUseCase struct {
	storyRepo repositories.StoryRepository
	logger    logger.Logger
}

// NewGetStoryVersionGraphUseCase creates a new GetStoryVersionGraphUseCase
func NewGetStoryVersionGraphUseCase(
	storyRepo repositories.StoryRepository,
	logger logger.Logger,
) *GetStoryVersionGraphUseCase {
	return &GetStoryVersionGraphUseCase{
		storyRepo: storyRepo,
		logger:    logger,
	}
}

// GetStoryVersionGraphInput represents the input for getting a version graph
type GetStoryVersionGraphInput struct {
	RootStoryID uuid.UUID
}

// GetStoryVersionGraphOutput represents the output of getting a version graph
type GetStoryVersionGraphOutput struct {
	Versions []*story.Story
}

// Execute retrieves all versions for a root story ID
func (uc *GetStoryVersionGraphUseCase) Execute(ctx context.Context, input GetStoryVersionGraphInput) (*GetStoryVersionGraphOutput, error) {
	versions, err := uc.storyRepo.GetVersionGraph(ctx, input.RootStoryID)
	if err != nil {
		uc.logger.Error("failed to get version graph", "error", err, "root_story_id", input.RootStoryID)
		return nil, err
	}

	return &GetStoryVersionGraphOutput{
		Versions: versions,
	}, nil
}

