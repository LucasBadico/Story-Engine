package story

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListStoriesUseCase handles listing stories
type ListStoriesUseCase struct {
	storyRepo repositories.StoryRepository
	logger    logger.Logger
}

// NewListStoriesUseCase creates a new ListStoriesUseCase
func NewListStoriesUseCase(
	storyRepo repositories.StoryRepository,
	logger logger.Logger,
) *ListStoriesUseCase {
	return &ListStoriesUseCase{
		storyRepo: storyRepo,
		logger:    logger,
	}
}

// ListStoriesInput represents the input for listing stories
type ListStoriesInput struct {
	TenantID uuid.UUID
	Limit    int
	Offset   int
}

// ListStoriesOutput represents the output of listing stories
type ListStoriesOutput struct {
	Stories []*story.Story
	Total   int
}

// Execute lists stories for a tenant
func (uc *ListStoriesUseCase) Execute(ctx context.Context, input ListStoriesInput) (*ListStoriesOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 50 // default
	}
	if limit > 100 {
		limit = 100 // max
	}

	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	stories, err := uc.storyRepo.ListByTenant(ctx, input.TenantID, limit, offset)
	if err != nil {
		uc.logger.Error("failed to list stories", "error", err, "tenant_id", input.TenantID)
		return nil, err
	}

	return &ListStoriesOutput{
		Stories: stories,
		Total:   len(stories),
	}, nil
}

