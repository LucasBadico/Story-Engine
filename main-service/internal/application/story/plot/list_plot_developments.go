package plot

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListPlotDevelopmentsUseCase handles listing plot developments
type ListPlotDevelopmentsUseCase struct {
	plotDevelopmentRepo repositories.PlotDevelopmentRepository
	logger              logger.Logger
}

// NewListPlotDevelopmentsUseCase creates a new ListPlotDevelopmentsUseCase
func NewListPlotDevelopmentsUseCase(
	plotDevelopmentRepo repositories.PlotDevelopmentRepository,
	logger logger.Logger,
) *ListPlotDevelopmentsUseCase {
	return &ListPlotDevelopmentsUseCase{
		plotDevelopmentRepo: plotDevelopmentRepo,
		logger:              logger,
	}
}

// ListPlotDevelopmentsInput represents the input for listing plot developments
type ListPlotDevelopmentsInput struct {
	TenantID    uuid.UUID
	PlotID      uuid.UUID
	AsOfChapter *int // optional filter
}

// ListPlotDevelopmentsOutput represents the output of listing plot developments
type ListPlotDevelopmentsOutput struct {
	PlotDevelopments []*story.PlotDevelopment
}

// Execute lists plot developments for a plot
func (uc *ListPlotDevelopmentsUseCase) Execute(ctx context.Context, input ListPlotDevelopmentsInput) (*ListPlotDevelopmentsOutput, error) {
	var devs []*story.PlotDevelopment
	var err error

	if input.AsOfChapter != nil {
		devs, err = uc.plotDevelopmentRepo.ListByPlotUpToChapter(ctx, input.TenantID, input.PlotID, *input.AsOfChapter)
	} else {
		devs, err = uc.plotDevelopmentRepo.ListByPlot(ctx, input.TenantID, input.PlotID)
	}
	if err != nil {
		uc.logger.Error("failed to list plot developments", "error", err, "plot_id", input.PlotID)
		return nil, err
	}

	return &ListPlotDevelopmentsOutput{
		PlotDevelopments: devs,
	}, nil
}
