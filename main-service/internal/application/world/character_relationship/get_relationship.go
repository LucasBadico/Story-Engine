package character_relationship

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetCharacterRelationshipUseCase handles getting a character relationship
type GetCharacterRelationshipUseCase struct {
	characterRelationshipRepo repositories.CharacterRelationshipRepository
	logger                    logger.Logger
}

// NewGetCharacterRelationshipUseCase creates a new GetCharacterRelationshipUseCase
func NewGetCharacterRelationshipUseCase(
	characterRelationshipRepo repositories.CharacterRelationshipRepository,
	logger logger.Logger,
) *GetCharacterRelationshipUseCase {
	return &GetCharacterRelationshipUseCase{
		characterRelationshipRepo: characterRelationshipRepo,
		logger:                    logger,
	}
}

// GetCharacterRelationshipInput represents the input for getting a character relationship
type GetCharacterRelationshipInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetCharacterRelationshipOutput represents the output of getting a character relationship
type GetCharacterRelationshipOutput struct {
	Relationship *world.CharacterRelationship
}

// Execute retrieves a character relationship by ID
func (uc *GetCharacterRelationshipUseCase) Execute(ctx context.Context, input GetCharacterRelationshipInput) (*GetCharacterRelationshipOutput, error) {
	relationship, err := uc.characterRelationshipRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get character relationship", "error", err, "relationship_id", input.ID)
		return nil, err
	}

	return &GetCharacterRelationshipOutput{
		Relationship: relationship,
	}, nil
}

