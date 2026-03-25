package plot

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/queue"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// EntityLinkInput represents an entity link to be created
type EntityLinkInput struct {
	Role       string
	EntityType string
	EntityID   uuid.UUID
}

// CreatePlotUseCase handles plot creation
type CreatePlotUseCase struct {
	plotRepo       repositories.PlotRepository
	storyRepo      repositories.StoryRepository
	auditLogRepo   repositories.AuditLogRepository
	ingestionQueue queue.IngestionQueue
	logger         logger.Logger
}

// NewCreatePlotUseCase creates a new CreatePlotUseCase
func NewCreatePlotUseCase(
	plotRepo repositories.PlotRepository,
	storyRepo repositories.StoryRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *CreatePlotUseCase {
	return &CreatePlotUseCase{
		plotRepo:       plotRepo,
		storyRepo:      storyRepo,
		auditLogRepo:   auditLogRepo,
		ingestionQueue: nil,
		logger:         logger,
	}
}

func (uc *CreatePlotUseCase) SetIngestionQueue(queue queue.IngestionQueue) {
	uc.ingestionQueue = queue
}

// CreatePlotInput represents the input for creating a plot
type CreatePlotInput struct {
	TenantID    uuid.UUID
	StoryID     uuid.UUID
	Title       string
	Type        string
	Description string
	Entities    []EntityLinkInput
}

// CreatePlotOutput represents the output of creating a plot
type CreatePlotOutput struct {
	Plot *story.Plot
}

// Execute creates a new plot
func (uc *CreatePlotUseCase) Execute(ctx context.Context, input CreatePlotInput) (*CreatePlotOutput, error) {
	// Validate story exists
	_, err := uc.storyRepo.GetByID(ctx, input.TenantID, input.StoryID)
	if err != nil {
		return nil, err
	}

	newPlot, err := story.NewPlot(input.TenantID, input.StoryID, input.Title, input.Type)
	if err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "plot",
			Message: err.Error(),
		}
	}

	if input.Description != "" {
		newPlot.Description = input.Description
	}

	// Add entity links
	for _, e := range input.Entities {
		newPlot.AddEntity(e.Role, e.EntityType, e.EntityID)
	}

	if err := uc.plotRepo.Create(ctx, newPlot); err != nil {
		uc.logger.Error("failed to create plot", "error", err, "story_id", input.StoryID)
		return nil, err
	}

	// Create entity links in DB
	for _, entity := range newPlot.Entities {
		if err := uc.plotRepo.AddEntity(ctx, &entity); err != nil {
			uc.logger.Error("failed to add entity to plot", "error", err, "entity_id", entity.EntityID)
		}
	}

	auditLog := audit.NewAuditLog(
		input.TenantID,
		nil,
		audit.ActionCreate,
		audit.EntityTypePlot,
		newPlot.ID,
		map[string]interface{}{
			"story_id": input.StoryID.String(),
			"title":    input.Title,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("plot created", "plot_id", newPlot.ID, "story_id", input.StoryID)
	uc.enqueueIngestion(ctx, input.TenantID, newPlot.ID)

	return &CreatePlotOutput{
		Plot: newPlot,
	}, nil
}

func (uc *CreatePlotUseCase) enqueueIngestion(ctx context.Context, tenantID uuid.UUID, plotID uuid.UUID) {
	if uc.ingestionQueue == nil {
		return
	}
	if err := uc.ingestionQueue.Push(ctx, tenantID, "plot", plotID); err != nil {
		uc.logger.Error("failed to enqueue plot ingestion", "error", err, "plot_id", plotID, "tenant_id", tenantID)
	}
}
