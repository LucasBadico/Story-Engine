package faction

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListFactionsUseCase handles listing factions for a world
type ListFactionsUseCase struct {
	factionRepo repositories.FactionRepository
	logger      logger.Logger
}

// NewListFactionsUseCase creates a new ListFactionsUseCase
func NewListFactionsUseCase(
	factionRepo repositories.FactionRepository,
	logger logger.Logger,
) *ListFactionsUseCase {
	return &ListFactionsUseCase{
		factionRepo: factionRepo,
		logger:      logger,
	}
}

// ListFactionsInput represents the input for listing factions
type ListFactionsInput struct {
	TenantID uuid.UUID
	WorldID  uuid.UUID
}

// ListFactionsOutput represents the output of listing factions
type ListFactionsOutput struct {
	Factions []*world.Faction
}

// Execute lists factions for a world
func (uc *ListFactionsUseCase) Execute(ctx context.Context, input ListFactionsInput) (*ListFactionsOutput, error) {
	factions, err := uc.factionRepo.ListByWorld(ctx, input.TenantID, input.WorldID)
	if err != nil {
		uc.logger.Error("failed to list factions", "error", err, "world_id", input.WorldID)
		return nil, err
	}

	return &ListFactionsOutput{
		Factions: factions,
	}, nil
}

