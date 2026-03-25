package plot

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeletePlotUseCase handles plot deletion
type DeletePlotUseCase struct {
	plotRepo     repositories.PlotRepository
	auditLogRepo repositories.AuditLogRepository
	logger       logger.Logger
}

// NewDeletePlotUseCase creates a new DeletePlotUseCase
func NewDeletePlotUseCase(
	plotRepo repositories.PlotRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *DeletePlotUseCase {
	return &DeletePlotUseCase{
		plotRepo:     plotRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// DeletePlotInput represents the input for deleting a plot
type DeletePlotInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes a plot
func (uc *DeletePlotUseCase) Execute(ctx context.Context, input DeletePlotInput) error {
	p, err := uc.plotRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	if err := uc.plotRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete plot", "error", err, "plot_id", input.ID)
		return err
	}

	auditLog := audit.NewAuditLog(
		input.TenantID,
		nil,
		audit.ActionDelete,
		audit.EntityTypePlot,
		p.ID,
		map[string]interface{}{
			"story_id": p.StoryID.String(),
			"title":    p.Title,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("plot deleted", "plot_id", input.ID)

	return nil
}
