package location

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetChildrenUseCase handles getting children of a location
type GetChildrenUseCase struct {
	locationRepo repositories.LocationRepository
	logger       logger.Logger
}

// NewGetChildrenUseCase creates a new GetChildrenUseCase
func NewGetChildrenUseCase(
	locationRepo repositories.LocationRepository,
	logger logger.Logger,
) *GetChildrenUseCase {
	return &GetChildrenUseCase{
		locationRepo: locationRepo,
		logger:       logger,
	}
}

// GetChildrenInput represents the input for getting children
type GetChildrenInput struct {
	TenantID   uuid.UUID
	LocationID uuid.UUID
}

// GetChildrenOutput represents the output of getting children
type GetChildrenOutput struct {
	Children []*world.Location
}

// Execute retrieves direct children of a location
func (uc *GetChildrenUseCase) Execute(ctx context.Context, input GetChildrenInput) (*GetChildrenOutput, error) {
	children, err := uc.locationRepo.GetChildren(ctx, input.TenantID, input.LocationID)
	if err != nil {
		uc.logger.Error("failed to get children", "error", err, "location_id", input.LocationID)
		return nil, err
	}

	return &GetChildrenOutput{
		Children: children,
	}, nil
}


