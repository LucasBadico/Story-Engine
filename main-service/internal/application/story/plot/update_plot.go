package plot

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdatePlotUseCase handles plot updates
type UpdatePlotUseCase struct {
	plotRepo     repositories.PlotRepository
	auditLogRepo repositories.AuditLogRepository
	logger       logger.Logger
}

// NewUpdatePlotUseCase creates a new UpdatePlotUseCase
func NewUpdatePlotUseCase(
	plotRepo repositories.PlotRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *UpdatePlotUseCase {
	return &UpdatePlotUseCase{
		plotRepo:     plotRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// UpdatePlotInput represents the input for updating a plot
type UpdatePlotInput struct {
	TenantID    uuid.UUID
	ID          uuid.UUID
	Title       *string
	Type        *string
	Description *string
	Status      *string
}

// UpdatePlotOutput represents the output of updating a plot
type UpdatePlotOutput struct {
	Plot *story.Plot
}

// Execute updates a plot
func (uc *UpdatePlotUseCase) Execute(ctx context.Context, input UpdatePlotInput) (*UpdatePlotOutput, error) {
	p, err := uc.plotRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		p.Title = *input.Title
	}
	if input.Type != nil {
		p.Type = *input.Type
	}
	if input.Description != nil {
		p.Description = *input.Description
	}
	if input.Status != nil {
		p.Status = *input.Status
	}
	p.UpdatedAt = time.Now()

	if err := uc.plotRepo.Update(ctx, p); err != nil {
		uc.logger.Error("failed to update plot", "error", err, "plot_id", input.ID)
		return nil, err
	}

	auditLog := audit.NewAuditLog(
		input.TenantID,
		nil,
		audit.ActionUpdate,
		audit.EntityTypePlot,
		p.ID,
		map[string]interface{}{
			"story_id": p.StoryID.String(),
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("plot updated", "plot_id", input.ID)

	return &UpdatePlotOutput{
		Plot: p,
	}, nil
}
