package character_relationship

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListCharacterRelationshipsUseCase handles listing character relationships
type ListCharacterRelationshipsUseCase struct {
	characterRelationshipRepo repositories.CharacterRelationshipRepository
	logger                    logger.Logger
}

// NewListCharacterRelationshipsUseCase creates a new ListCharacterRelationshipsUseCase
func NewListCharacterRelationshipsUseCase(
	characterRelationshipRepo repositories.CharacterRelationshipRepository,
	logger logger.Logger,
) *ListCharacterRelationshipsUseCase {
	return &ListCharacterRelationshipsUseCase{
		characterRelationshipRepo: characterRelationshipRepo,
		logger:                    logger,
	}
}

// ListCharacterRelationshipsInput represents the input for listing character relationships
type ListCharacterRelationshipsInput struct {
	TenantID    uuid.UUID
	CharacterID uuid.UUID
}

// ListCharacterRelationshipsOutput represents the output of listing character relationships
type ListCharacterRelationshipsOutput struct {
	Relationships []*world.CharacterRelationship
}

// Execute lists all relationships for a character
func (uc *ListCharacterRelationshipsUseCase) Execute(ctx context.Context, input ListCharacterRelationshipsInput) (*ListCharacterRelationshipsOutput, error) {
	relationships, err := uc.characterRelationshipRepo.ListByCharacter(ctx, input.TenantID, input.CharacterID)
	if err != nil {
		uc.logger.Error("failed to list character relationships", "error", err, "character_id", input.CharacterID)
		return nil, err
	}

	return &ListCharacterRelationshipsOutput{
		Relationships: relationships,
	}, nil
}

