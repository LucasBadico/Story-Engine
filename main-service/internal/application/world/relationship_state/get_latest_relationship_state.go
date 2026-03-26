package relationship_state

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetLatestRelationshipStateUseCase handles retrieving the latest relationship state
type GetLatestRelationshipStateUseCase struct {
	relationshipStateRepo repositories.RelationshipStateRepository
	logger                logger.Logger
}

// NewGetLatestRelationshipStateUseCase creates a new GetLatestRelationshipStateUseCase
func NewGetLatestRelationshipStateUseCase(
	relationshipStateRepo repositories.RelationshipStateRepository,
	logger logger.Logger,
) *GetLatestRelationshipStateUseCase {
	return &GetLatestRelationshipStateUseCase{
		relationshipStateRepo: relationshipStateRepo,
		logger:                logger,
	}
}

// GetLatestRelationshipStateInput represents the input for getting the latest relationship state
type GetLatestRelationshipStateInput struct {
	TenantID         uuid.UUID
	SourceEntityType string
	SourceEntityID   uuid.UUID
	TargetEntityType string
	TargetEntityID   uuid.UUID
	AsOfChapter      int
}

// GetLatestRelationshipStateOutput represents the output of getting the latest relationship state
type GetLatestRelationshipStateOutput struct {
	RelationshipState *world.RelationshipState
}

// Execute retrieves the latest relationship state for an entity pair at or before the given chapter
func (uc *GetLatestRelationshipStateUseCase) Execute(ctx context.Context, input GetLatestRelationshipStateInput) (*GetLatestRelationshipStateOutput, error) {
	rs, err := uc.relationshipStateRepo.GetLatestByEntityPair(ctx, input.TenantID, input.SourceEntityType, input.SourceEntityID, input.TargetEntityType, input.TargetEntityID, input.AsOfChapter)
	if err != nil {
		uc.logger.Error("failed to get latest relationship state", "error", err, "source_entity_id", input.SourceEntityID, "target_entity_id", input.TargetEntityID, "as_of_chapter", input.AsOfChapter)
		return nil, err
	}

	return &GetLatestRelationshipStateOutput{
		RelationshipState: rs,
	}, nil
}
