package faction

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetChildrenUseCase handles getting children of a faction
type GetChildrenUseCase struct {
	factionRepo repositories.FactionRepository
	logger      logger.Logger
}

// NewGetChildrenUseCase creates a new GetChildrenUseCase
func NewGetChildrenUseCase(
	factionRepo repositories.FactionRepository,
	logger logger.Logger,
) *GetChildrenUseCase {
	return &GetChildrenUseCase{
		factionRepo: factionRepo,
		logger:      logger,
	}
}

// GetChildrenInput represents the input for getting children
type GetChildrenInput struct {
	TenantID  uuid.UUID
	FactionID uuid.UUID
}

// GetChildrenOutput represents the output of getting children
type GetChildrenOutput struct {
	Children []*world.Faction
}

// Execute retrieves direct children of a faction
func (uc *GetChildrenUseCase) Execute(ctx context.Context, input GetChildrenInput) (*GetChildrenOutput, error) {
	children, err := uc.factionRepo.GetChildren(ctx, input.TenantID, input.FactionID)
	if err != nil {
		uc.logger.Error("failed to get children", "error", err, "faction_id", input.FactionID)
		return nil, err
	}

	return &GetChildrenOutput{
		Children: children,
	}, nil
}

