package location

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetLocationUseCase handles location retrieval
type GetLocationUseCase struct {
	locationRepo repositories.LocationRepository
	logger       logger.Logger
}

// NewGetLocationUseCase creates a new GetLocationUseCase
func NewGetLocationUseCase(
	locationRepo repositories.LocationRepository,
	logger logger.Logger,
) *GetLocationUseCase {
	return &GetLocationUseCase{
		locationRepo: locationRepo,
		logger:       logger,
	}
}

// GetLocationInput represents the input for getting a location
type GetLocationInput struct {
	ID uuid.UUID
}

// GetLocationOutput represents the output of getting a location
type GetLocationOutput struct {
	Location *world.Location
}

// Execute retrieves a location by ID
func (uc *GetLocationUseCase) Execute(ctx context.Context, input GetLocationInput) (*GetLocationOutput, error) {
	l, err := uc.locationRepo.GetByID(ctx, input.ID)
	if err != nil {
		uc.logger.Error("failed to get location", "error", err, "location_id", input.ID)
		return nil, err
	}

	return &GetLocationOutput{
		Location: l,
	}, nil
}


