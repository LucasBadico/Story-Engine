package relationship_state

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListRelationshipStatesUseCase handles listing relationship states
type ListRelationshipStatesUseCase struct {
	relationshipStateRepo repositories.RelationshipStateRepository
	logger                logger.Logger
}

// NewListRelationshipStatesUseCase creates a new ListRelationshipStatesUseCase
func NewListRelationshipStatesUseCase(
	relationshipStateRepo repositories.RelationshipStateRepository,
	logger logger.Logger,
) *ListRelationshipStatesUseCase {
	return &ListRelationshipStatesUseCase{
		relationshipStateRepo: relationshipStateRepo,
		logger:                logger,
	}
}

// ListRelationshipStatesInput represents the input for listing relationship states
type ListRelationshipStatesInput struct {
	TenantID  uuid.UUID
	ChapterID uuid.UUID
}

// ListRelationshipStatesOutput represents the output of listing relationship states
type ListRelationshipStatesOutput struct {
	RelationshipStates []*world.RelationshipState
}

// Execute lists relationship states for a chapter
func (uc *ListRelationshipStatesUseCase) Execute(ctx context.Context, input ListRelationshipStatesInput) (*ListRelationshipStatesOutput, error) {
	states, err := uc.relationshipStateRepo.ListByChapter(ctx, input.TenantID, input.ChapterID)
	if err != nil {
		uc.logger.Error("failed to list relationship states", "error", err, "chapter_id", input.ChapterID)
		return nil, err
	}

	return &ListRelationshipStatesOutput{
		RelationshipStates: states,
	}, nil
}
