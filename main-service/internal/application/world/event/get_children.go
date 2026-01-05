package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetChildrenUseCase handles retrieving children of an event
type GetChildrenUseCase struct {
	eventRepo repositories.EventRepository
	logger    logger.Logger
}

// NewGetChildrenUseCase creates a new GetChildrenUseCase
func NewGetChildrenUseCase(
	eventRepo repositories.EventRepository,
	logger logger.Logger,
) *GetChildrenUseCase {
	return &GetChildrenUseCase{
		eventRepo: eventRepo,
		logger:    logger,
	}
}

// GetChildrenInput represents the input for getting children
type GetChildrenInput struct {
	TenantID uuid.UUID
	ParentID uuid.UUID
}

// GetChildrenOutput represents the output of getting children
type GetChildrenOutput struct {
	Events []*world.Event
}

// Execute retrieves direct children of an event
func (uc *GetChildrenUseCase) Execute(ctx context.Context, input GetChildrenInput) (*GetChildrenOutput, error) {
	events, err := uc.eventRepo.GetChildren(ctx, input.TenantID, input.ParentID)
	if err != nil {
		uc.logger.Error("failed to get event children", "error", err, "parent_id", input.ParentID)
		return nil, err
	}

	return &GetChildrenOutput{
		Events: events,
	}, nil
}

