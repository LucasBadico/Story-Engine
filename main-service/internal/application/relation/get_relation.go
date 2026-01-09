package relation

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/relation"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetRelationUseCase handles retrieving a relation by ID
type GetRelationUseCase struct {
	relationRepo repositories.EntityRelationRepository
	logger       logger.Logger
}

// NewGetRelationUseCase creates a new GetRelationUseCase
func NewGetRelationUseCase(
	relationRepo repositories.EntityRelationRepository,
	logger logger.Logger,
) *GetRelationUseCase {
	return &GetRelationUseCase{
		relationRepo: relationRepo,
		logger:       logger,
	}
}

// GetRelationInput represents the input for getting a relation
type GetRelationInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetRelationOutput represents the output of getting a relation
type GetRelationOutput struct {
	Relation *relation.EntityRelation
}

// Execute retrieves a relation by ID
func (uc *GetRelationUseCase) Execute(ctx context.Context, input GetRelationInput) (*GetRelationOutput, error) {
	rel, err := uc.relationRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get relation", "error", err, "relation_id", input.ID)
		return nil, err
	}

	return &GetRelationOutput{
		Relation: rel,
	}, nil
}

