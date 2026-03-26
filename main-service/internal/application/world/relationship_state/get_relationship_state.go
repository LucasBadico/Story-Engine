package relationship_state

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetRelationshipStateUseCase handles relationship state retrieval
type GetRelationshipStateUseCase struct {
	relationshipStateRepo repositories.RelationshipStateRepository
	logger                logger.Logger
}

// NewGetRelationshipStateUseCase creates a new GetRelationshipStateUseCase
func NewGetRelationshipStateUseCase(
	relationshipStateRepo repositories.RelationshipStateRepository,
	logger logger.Logger,
) *GetRelationshipStateUseCase {
	return &GetRelationshipStateUseCase{
		relationshipStateRepo: relationshipStateRepo,
		logger:                logger,
	}
}

// GetRelationshipStateInput represents the input for getting a relationship state
type GetRelationshipStateInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetRelationshipStateOutput represents the output of getting a relationship state
type GetRelationshipStateOutput struct {
	RelationshipState *world.RelationshipState
}

// Execute retrieves a relationship state by ID
func (uc *GetRelationshipStateUseCase) Execute(ctx context.Context, input GetRelationshipStateInput) (*GetRelationshipStateOutput, error) {
	rs, err := uc.relationshipStateRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get relationship state", "error", err, "relationship_state_id", input.ID)
		return nil, err
	}

	return &GetRelationshipStateOutput{
		RelationshipState: rs,
	}, nil
}
