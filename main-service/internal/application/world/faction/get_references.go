package faction

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetReferencesUseCase handles retrieving references for a faction
type GetReferencesUseCase struct {
	factionReferenceRepo repositories.FactionReferenceRepository
	logger               logger.Logger
}

// NewGetReferencesUseCase creates a new GetReferencesUseCase
func NewGetReferencesUseCase(
	factionReferenceRepo repositories.FactionReferenceRepository,
	logger logger.Logger,
) *GetReferencesUseCase {
	return &GetReferencesUseCase{
		factionReferenceRepo: factionReferenceRepo,
		logger:               logger,
	}
}

// GetReferencesInput represents the input for getting references
type GetReferencesInput struct {
	TenantID  uuid.UUID
	FactionID uuid.UUID
}

// GetReferencesOutput represents the output of getting references
type GetReferencesOutput struct {
	References []*world.FactionReference
}

// Execute retrieves references for a faction
func (uc *GetReferencesUseCase) Execute(ctx context.Context, input GetReferencesInput) (*GetReferencesOutput, error) {
	references, err := uc.factionReferenceRepo.ListByFaction(ctx, input.TenantID, input.FactionID)
	if err != nil {
		uc.logger.Error("failed to get faction references", "error", err, "faction_id", input.FactionID)
		return nil, err
	}

	return &GetReferencesOutput{
		References: references,
	}, nil
}

