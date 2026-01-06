package character_relationship

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateCharacterRelationshipUseCase handles updating character relationships
type UpdateCharacterRelationshipUseCase struct {
	characterRelationshipRepo repositories.CharacterRelationshipRepository
	logger                    logger.Logger
}

// NewUpdateCharacterRelationshipUseCase creates a new UpdateCharacterRelationshipUseCase
func NewUpdateCharacterRelationshipUseCase(
	characterRelationshipRepo repositories.CharacterRelationshipRepository,
	logger logger.Logger,
) *UpdateCharacterRelationshipUseCase {
	return &UpdateCharacterRelationshipUseCase{
		characterRelationshipRepo: characterRelationshipRepo,
		logger:                    logger,
	}
}

// UpdateCharacterRelationshipInput represents the input for updating a character relationship
type UpdateCharacterRelationshipInput struct {
	TenantID         uuid.UUID
	ID               uuid.UUID
	RelationshipType *string
	Description      *string
	Bidirectional    *bool
}

// UpdateCharacterRelationshipOutput represents the output of updating a character relationship
type UpdateCharacterRelationshipOutput struct {
	Relationship *world.CharacterRelationship
}

// Execute updates a character relationship
func (uc *UpdateCharacterRelationshipUseCase) Execute(ctx context.Context, input UpdateCharacterRelationshipInput) (*UpdateCharacterRelationshipOutput, error) {
	relationship, err := uc.characterRelationshipRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get character relationship", "error", err, "relationship_id", input.ID)
		return nil, err
	}

	if input.RelationshipType != nil {
		if err := relationship.UpdateRelationshipType(*input.RelationshipType); err != nil {
			return nil, err
		}
	}

	if input.Description != nil {
		relationship.UpdateDescription(*input.Description)
	}

	if input.Bidirectional != nil {
		relationship.UpdateBidirectional(*input.Bidirectional)
	}

	if err := relationship.Validate(); err != nil {
		return nil, err
	}

	if err := uc.characterRelationshipRepo.Update(ctx, relationship); err != nil {
		uc.logger.Error("failed to update character relationship", "error", err, "relationship_id", input.ID)
		return nil, err
	}

	uc.logger.Info("character relationship updated", "relationship_id", input.ID)

	return &UpdateCharacterRelationshipOutput{
		Relationship: relationship,
	}, nil
}

