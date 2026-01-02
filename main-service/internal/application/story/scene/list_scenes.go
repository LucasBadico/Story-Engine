package scene

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListScenesUseCase handles listing scenes
type ListScenesUseCase struct {
	sceneRepo repositories.SceneRepository
	logger    logger.Logger
}

// NewListScenesUseCase creates a new ListScenesUseCase
func NewListScenesUseCase(
	sceneRepo repositories.SceneRepository,
	logger logger.Logger,
) *ListScenesUseCase {
	return &ListScenesUseCase{
		sceneRepo: sceneRepo,
		logger:    logger,
	}
}

// ListScenesInput represents the input for listing scenes
type ListScenesInput struct {
	TenantID  uuid.UUID
	StoryID   uuid.UUID
	ChapterID *uuid.UUID // optional: filter by chapter
}

// ListScenesOutput represents the output of listing scenes
type ListScenesOutput struct {
	Scenes []*story.Scene
	Total  int
}

// Execute lists scenes
func (uc *ListScenesUseCase) Execute(ctx context.Context, input ListScenesInput) (*ListScenesOutput, error) {
	var scenes []*story.Scene
	var err error

	if input.ChapterID != nil {
		scenes, err = uc.sceneRepo.ListByChapter(ctx, input.TenantID, *input.ChapterID)
	} else {
		scenes, err = uc.sceneRepo.ListByStory(ctx, input.TenantID, input.StoryID)
	}

	if err != nil {
		uc.logger.Error("failed to list scenes", "error", err, "story_id", input.StoryID, "tenant_id", input.TenantID)
		return nil, err
	}

	return &ListScenesOutput{
		Scenes: scenes,
		Total:  len(scenes),
	}, nil
}

