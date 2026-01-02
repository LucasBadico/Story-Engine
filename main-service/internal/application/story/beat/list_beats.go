package beat

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListBeatsUseCase handles listing beats
type ListBeatsUseCase struct {
	beatRepo repositories.BeatRepository
	logger   logger.Logger
}

// NewListBeatsUseCase creates a new ListBeatsUseCase
func NewListBeatsUseCase(
	beatRepo repositories.BeatRepository,
	logger logger.Logger,
) *ListBeatsUseCase {
	return &ListBeatsUseCase{
		beatRepo: beatRepo,
		logger:   logger,
	}
}

// ListBeatsInput represents the input for listing beats
type ListBeatsInput struct {
	TenantID uuid.UUID
	SceneID  uuid.UUID
	StoryID  *uuid.UUID // optional: filter by story
}

// ListBeatsOutput represents the output of listing beats
type ListBeatsOutput struct {
	Beats []*story.Beat
	Total int
}

// Execute lists beats
func (uc *ListBeatsUseCase) Execute(ctx context.Context, input ListBeatsInput) (*ListBeatsOutput, error) {
	var beats []*story.Beat
	var err error

	if input.StoryID != nil {
		beats, err = uc.beatRepo.ListByStory(ctx, input.TenantID, *input.StoryID)
	} else {
		beats, err = uc.beatRepo.ListByScene(ctx, input.TenantID, input.SceneID)
	}

	if err != nil {
		uc.logger.Error("failed to list beats", "error", err, "scene_id", input.SceneID, "tenant_id", input.TenantID)
		return nil, err
	}

	return &ListBeatsOutput{
		Beats: beats,
		Total: len(beats),
	}, nil
}

