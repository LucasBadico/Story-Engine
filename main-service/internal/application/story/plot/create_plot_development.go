package plot

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreatePlotDevelopmentUseCase handles plot development creation
type CreatePlotDevelopmentUseCase struct {
	plotDevelopmentRepo repositories.PlotDevelopmentRepository
	plotRepo            repositories.PlotRepository
	auditLogRepo        repositories.AuditLogRepository
	logger              logger.Logger
}

// NewCreatePlotDevelopmentUseCase creates a new CreatePlotDevelopmentUseCase
func NewCreatePlotDevelopmentUseCase(
	plotDevelopmentRepo repositories.PlotDevelopmentRepository,
	plotRepo repositories.PlotRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *CreatePlotDevelopmentUseCase {
	return &CreatePlotDevelopmentUseCase{
		plotDevelopmentRepo: plotDevelopmentRepo,
		plotRepo:            plotRepo,
		auditLogRepo:        auditLogRepo,
		logger:              logger,
	}
}

// CreatePlotDevelopmentInput represents the input for creating a plot development
type CreatePlotDevelopmentInput struct {
	TenantID           uuid.UUID
	PlotID             uuid.UUID
	ChapterNumber      int
	Action             string
	Description        string
	Notes              string
	TimelineDay        *int
	TimelineConfidence *string
	TimelineEventID    *uuid.UUID
}

// CreatePlotDevelopmentOutput represents the output of creating a plot development
type CreatePlotDevelopmentOutput struct {
	PlotDevelopment *story.PlotDevelopment
}

// Execute creates a new plot development
func (uc *CreatePlotDevelopmentUseCase) Execute(ctx context.Context, input CreatePlotDevelopmentInput) (*CreatePlotDevelopmentOutput, error) {
	// Validate plot exists
	_, err := uc.plotRepo.GetByID(ctx, input.TenantID, input.PlotID)
	if err != nil {
		return nil, err
	}

	newDev, err := story.NewPlotDevelopment(input.TenantID, input.PlotID, input.ChapterNumber, input.Action, input.Description)
	if err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "plot_development",
			Message: err.Error(),
		}
	}

	if input.TimelineDay != nil && input.TimelineConfidence != nil {
		newDev.SetTimeline(*input.TimelineDay, *input.TimelineConfidence)
	}

	newDev.Notes = input.Notes

	if input.TimelineEventID != nil {
		newDev.SetTimelineEvent(*input.TimelineEventID)
	}

	if err := uc.plotDevelopmentRepo.Create(ctx, newDev); err != nil {
		uc.logger.Error("failed to create plot development", "error", err, "plot_id", input.PlotID)
		return nil, err
	}

	auditLog := audit.NewAuditLog(
		input.TenantID,
		nil,
		audit.ActionCreate,
		audit.EntityTypePlotDevelopment,
		newDev.ID,
		map[string]interface{}{
			"plot_id":        input.PlotID.String(),
			"chapter_number": input.ChapterNumber,
			"action":         input.Action,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("plot development created", "plot_development_id", newDev.ID, "plot_id", input.PlotID)

	return &CreatePlotDevelopmentOutput{
		PlotDevelopment: newDev,
	}, nil
}
