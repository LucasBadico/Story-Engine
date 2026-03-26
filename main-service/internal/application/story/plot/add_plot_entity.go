package plot

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddPlotEntityUseCase handles adding an entity to a plot
type AddPlotEntityUseCase struct {
	plotRepo repositories.PlotRepository
	logger   logger.Logger
}

// NewAddPlotEntityUseCase creates a new AddPlotEntityUseCase
func NewAddPlotEntityUseCase(
	plotRepo repositories.PlotRepository,
	logger logger.Logger,
) *AddPlotEntityUseCase {
	return &AddPlotEntityUseCase{
		plotRepo: plotRepo,
		logger:   logger,
	}
}

// AddPlotEntityInput represents the input for adding an entity to a plot
type AddPlotEntityInput struct {
	TenantID   uuid.UUID
	PlotID     uuid.UUID
	Role       string
	EntityType string
	EntityID   uuid.UUID
}

// AddPlotEntityOutput represents the output of adding an entity to a plot
type AddPlotEntityOutput struct {
	PlotEntity *story.PlotEntity
}

// Execute adds an entity to a plot
func (uc *AddPlotEntityUseCase) Execute(ctx context.Context, input AddPlotEntityInput) (*AddPlotEntityOutput, error) {
	// Validate plot exists
	_, err := uc.plotRepo.GetByID(ctx, input.TenantID, input.PlotID)
	if err != nil {
		return nil, err
	}

	entity := &story.PlotEntity{
		ID:         uuid.New(),
		PlotID:     input.PlotID,
		Role:       input.Role,
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		CreatedAt:  time.Now(),
	}

	if err := uc.plotRepo.AddEntity(ctx, entity); err != nil {
		uc.logger.Error("failed to add entity to plot", "error", err, "plot_id", input.PlotID)
		return nil, err
	}

	uc.logger.Info("entity added to plot", "plot_id", input.PlotID, "entity_id", input.EntityID)

	return &AddPlotEntityOutput{
		PlotEntity: entity,
	}, nil
}
