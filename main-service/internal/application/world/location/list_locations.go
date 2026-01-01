package location

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListLocationsUseCase handles listing locations
type ListLocationsUseCase struct {
	locationRepo repositories.LocationRepository
	logger       logger.Logger
}

// NewListLocationsUseCase creates a new ListLocationsUseCase
func NewListLocationsUseCase(
	locationRepo repositories.LocationRepository,
	logger logger.Logger,
) *ListLocationsUseCase {
	return &ListLocationsUseCase{
		locationRepo: locationRepo,
		logger:       logger,
	}
}

// ListLocationsInput represents the input for listing locations
type ListLocationsInput struct {
	WorldID uuid.UUID
	Format  string // "flat" or "tree"
	Limit   int
	Offset  int
}

// ListLocationsOutput represents the output of listing locations
type ListLocationsOutput struct {
	Locations []*world.Location
	Total     int
}

// Execute lists locations for a world
func (uc *ListLocationsUseCase) Execute(ctx context.Context, input ListLocationsInput) (*ListLocationsOutput, error) {
	var locations []*world.Location
	var err error

	if input.Format == "tree" {
		locations, err = uc.locationRepo.ListByWorldTree(ctx, input.WorldID)
	} else {
		limit := input.Limit
		if limit <= 0 {
			limit = 50
		}
		if limit > 100 {
			limit = 100
		}
		locations, err = uc.locationRepo.ListByWorld(ctx, input.WorldID, limit, input.Offset)
	}

	if err != nil {
		uc.logger.Error("failed to list locations", "error", err, "world_id", input.WorldID)
		return nil, err
	}

	total, err := uc.locationRepo.CountByWorld(ctx, input.WorldID)
	if err != nil {
		uc.logger.Warn("failed to count locations", "error", err)
		total = len(locations)
	}

	return &ListLocationsOutput{
		Locations: locations,
		Total:     total,
	}, nil
}

