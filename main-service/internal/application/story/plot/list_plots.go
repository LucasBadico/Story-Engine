package plot

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListPlotsUseCase handles listing plots
type ListPlotsUseCase struct {
	plotRepo repositories.PlotRepository
	logger   logger.Logger
}

// NewListPlotsUseCase creates a new ListPlotsUseCase
func NewListPlotsUseCase(
	plotRepo repositories.PlotRepository,
	logger logger.Logger,
) *ListPlotsUseCase {
	return &ListPlotsUseCase{
		plotRepo: plotRepo,
		logger:   logger,
	}
}

// ListPlotsInput represents the input for listing plots
type ListPlotsInput struct {
	TenantID uuid.UUID
	StoryID  uuid.UUID
	Status   string // optional filter
}

// ListPlotsOutput represents the output of listing plots
type ListPlotsOutput struct {
	Plots []*story.Plot
}

// Execute lists plots for a story, optionally filtered by status
func (uc *ListPlotsUseCase) Execute(ctx context.Context, input ListPlotsInput) (*ListPlotsOutput, error) {
	var plots []*story.Plot
	var err error

	if input.Status != "" {
		plots, err = uc.plotRepo.ListByStoryAndStatus(ctx, input.TenantID, input.StoryID, input.Status)
	} else {
		plots, err = uc.plotRepo.ListByStory(ctx, input.TenantID, input.StoryID)
	}
	if err != nil {
		uc.logger.Error("failed to list plots", "error", err, "story_id", input.StoryID)
		return nil, err
	}

	return &ListPlotsOutput{
		Plots: plots,
	}, nil
}
