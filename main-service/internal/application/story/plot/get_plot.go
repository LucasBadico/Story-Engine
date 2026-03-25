package plot

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetPlotUseCase handles plot retrieval
type GetPlotUseCase struct {
	plotRepo repositories.PlotRepository
	logger   logger.Logger
}

// NewGetPlotUseCase creates a new GetPlotUseCase
func NewGetPlotUseCase(
	plotRepo repositories.PlotRepository,
	logger logger.Logger,
) *GetPlotUseCase {
	return &GetPlotUseCase{
		plotRepo: plotRepo,
		logger:   logger,
	}
}

// GetPlotInput represents the input for getting a plot
type GetPlotInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetPlotOutput represents the output of getting a plot
type GetPlotOutput struct {
	Plot *story.Plot
}

// Execute retrieves a plot by ID
func (uc *GetPlotUseCase) Execute(ctx context.Context, input GetPlotInput) (*GetPlotOutput, error) {
	p, err := uc.plotRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get plot", "error", err, "plot_id", input.ID)
		return nil, err
	}

	return &GetPlotOutput{
		Plot: p,
	}, nil
}
