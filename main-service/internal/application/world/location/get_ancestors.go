package location

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetAncestorsUseCase handles getting ancestors of a location
type GetAncestorsUseCase struct {
	locationRepo repositories.LocationRepository
	logger       logger.Logger
}

// NewGetAncestorsUseCase creates a new GetAncestorsUseCase
func NewGetAncestorsUseCase(
	locationRepo repositories.LocationRepository,
	logger logger.Logger,
) *GetAncestorsUseCase {
	return &GetAncestorsUseCase{
		locationRepo: locationRepo,
		logger:       logger,
	}
}

// GetAncestorsInput represents the input for getting ancestors
type GetAncestorsInput struct {
	LocationID uuid.UUID
}

// GetAncestorsOutput represents the output of getting ancestors
type GetAncestorsOutput struct {
	Ancestors []*world.Location
}

// Execute retrieves all ancestors of a location (path to root)
func (uc *GetAncestorsUseCase) Execute(ctx context.Context, input GetAncestorsInput) (*GetAncestorsOutput, error) {
	ancestors, err := uc.locationRepo.GetAncestors(ctx, input.LocationID)
	if err != nil {
		uc.logger.Error("failed to get ancestors", "error", err, "location_id", input.LocationID)
		return nil, err
	}

	return &GetAncestorsOutput{
		Ancestors: ancestors,
	}, nil
}

