package faction

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetFactionUseCase handles retrieving a faction
type GetFactionUseCase struct {
	factionRepo repositories.FactionRepository
	logger      logger.Logger
}

// NewGetFactionUseCase creates a new GetFactionUseCase
func NewGetFactionUseCase(
	factionRepo repositories.FactionRepository,
	logger logger.Logger,
) *GetFactionUseCase {
	return &GetFactionUseCase{
		factionRepo: factionRepo,
		logger:      logger,
	}
}

// GetFactionInput represents the input for getting a faction
type GetFactionInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetFactionOutput represents the output of getting a faction
type GetFactionOutput struct {
	Faction *world.Faction
}

// Execute retrieves a faction by ID
func (uc *GetFactionUseCase) Execute(ctx context.Context, input GetFactionInput) (*GetFactionOutput, error) {
	faction, err := uc.factionRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get faction", "error", err, "faction_id", input.ID)
		return nil, err
	}

	return &GetFactionOutput{
		Faction: faction,
	}, nil
}

