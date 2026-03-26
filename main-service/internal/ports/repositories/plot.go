package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
)

// PlotRepository defines the interface for plot persistence
type PlotRepository interface {
	Create(ctx context.Context, plot *story.Plot) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.Plot, error)
	ListByStory(ctx context.Context, tenantID, storyID uuid.UUID) ([]*story.Plot, error)
	ListByStoryAndStatus(ctx context.Context, tenantID, storyID uuid.UUID, status string) ([]*story.Plot, error)
	Update(ctx context.Context, plot *story.Plot) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	// Entity links
	AddEntity(ctx context.Context, entity *story.PlotEntity) error
	ListEntities(ctx context.Context, plotID uuid.UUID) ([]*story.PlotEntity, error)
	DeleteEntities(ctx context.Context, plotID uuid.UUID) error
}

// PlotDevelopmentRepository defines the interface for plot development persistence
type PlotDevelopmentRepository interface {
	Create(ctx context.Context, dev *story.PlotDevelopment) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.PlotDevelopment, error)
	ListByPlot(ctx context.Context, tenantID, plotID uuid.UUID) ([]*story.PlotDevelopment, error)
	ListByPlotUpToChapter(ctx context.Context, tenantID, plotID uuid.UUID, asOfChapter int) ([]*story.PlotDevelopment, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
}
