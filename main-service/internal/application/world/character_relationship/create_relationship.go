package character_relationship

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateCharacterRelationshipUseCase handles creating character relationships
type CreateCharacterRelationshipUseCase struct {
	characterRelationshipRepo repositories.CharacterRelationshipRepository
	characterRepo             repositories.CharacterRepository
	logger                    logger.Logger
}

// NewCreateCharacterRelationshipUseCase creates a new CreateCharacterRelationshipUseCase
func NewCreateCharacterRelationshipUseCase(
	characterRelationshipRepo repositories.CharacterRelationshipRepository,
	characterRepo repositories.CharacterRepository,
	logger logger.Logger,
) *CreateCharacterRelationshipUseCase {
	return &CreateCharacterRelationshipUseCase{
		characterRelationshipRepo: characterRelationshipRepo,
		characterRepo:             characterRepo,
		logger:                    logger,
	}
}

// CreateCharacterRelationshipInput represents the input for creating a character relationship
type CreateCharacterRelationshipInput struct {
	TenantID         uuid.UUID
	Character1ID     uuid.UUID
	Character2ID     uuid.UUID
	RelationshipType string
	Description      string
	Bidirectional    bool
}

// CreateCharacterRelationshipOutput represents the output of creating a character relationship
type CreateCharacterRelationshipOutput struct {
	Relationship *world.CharacterRelationship
}

// Execute creates a new character relationship
func (uc *CreateCharacterRelationshipUseCase) Execute(ctx context.Context, input CreateCharacterRelationshipInput) (*CreateCharacterRelationshipOutput, error) {
	// Validate that characters exist and belong to the same tenant
	char1, err := uc.characterRepo.GetByID(ctx, input.TenantID, input.Character1ID)
	if err != nil {
		uc.logger.Error("failed to get character1", "error", err, "character1_id", input.Character1ID)
		return nil, err
	}

	char2, err := uc.characterRepo.GetByID(ctx, input.TenantID, input.Character2ID)
	if err != nil {
		uc.logger.Error("failed to get character2", "error", err, "character2_id", input.Character2ID)
		return nil, err
	}

	// Validate they're in the same world
	if char1.WorldID != char2.WorldID {
		uc.logger.Error("characters must be in the same world", "character1_world", char1.WorldID, "character2_world", char2.WorldID)
		return nil, &platformerrors.ValidationError{
			Field:   "character2_id",
			Message: "characters must be in the same world",
		}
	}

	// Create the relationship
	relationship, err := world.NewCharacterRelationship(input.TenantID, input.Character1ID, input.Character2ID, input.RelationshipType)
	if err != nil {
		return nil, err
	}

	relationship.Description = input.Description
	relationship.Bidirectional = input.Bidirectional

	if err := relationship.Validate(); err != nil {
		return nil, err
	}

	if err := uc.characterRelationshipRepo.Create(ctx, relationship); err != nil {
		uc.logger.Error("failed to create character relationship", "error", err)
		return nil, err
	}

	uc.logger.Info("character relationship created", "relationship_id", relationship.ID, "character1_id", input.Character1ID, "character2_id", input.Character2ID)

	return &CreateCharacterRelationshipOutput{
		Relationship: relationship,
	}, nil
}

