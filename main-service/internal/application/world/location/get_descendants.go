package location

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetDescendantsUseCase handles getting descendants of a location
type GetDescendantsUseCase struct {
	locationRepo repositories.LocationRepository
	logger       logger.Logger
}

// NewGetDescendantsUseCase creates a new GetDescendantsUseCase
func NewGetDescendantsUseCase(
	locationRepo repositories.LocationRepository,
	logger logger.Logger,
) *GetDescendantsUseCase {
	return &GetDescendantsUseCase{
		locationRepo: locationRepo,
		logger:       logger,
	}
}

// GetDescendantsInput represents the input for getting descendants
type GetDescendantsInput struct {
	LocationID uuid.UUID
}

// GetDescendantsOutput represents the output of getting descendants
type GetDescendantsOutput struct {
	Descendants []*world.Location
}

// Execute retrieves all descendants of a location (recursive)
func (uc *GetDescendantsUseCase) Execute(ctx context.Context, input GetDescendantsInput) (*GetDescendantsOutput, error) {
	descendants, err := uc.locationRepo.GetDescendants(ctx, input.LocationID)
	if err != nil {
		uc.logger.Error("failed to get descendants", "error", err, "location_id", input.LocationID)
		return nil, err
	}

	return &GetDescendantsOutput{
		Descendants: descendants,
	}, nil
}

