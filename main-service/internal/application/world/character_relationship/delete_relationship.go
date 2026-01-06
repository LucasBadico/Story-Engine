package character_relationship

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteCharacterRelationshipUseCase handles deleting character relationships
type DeleteCharacterRelationshipUseCase struct {
	characterRelationshipRepo repositories.CharacterRelationshipRepository
	logger                    logger.Logger
}

// NewDeleteCharacterRelationshipUseCase creates a new DeleteCharacterRelationshipUseCase
func NewDeleteCharacterRelationshipUseCase(
	characterRelationshipRepo repositories.CharacterRelationshipRepository,
	logger logger.Logger,
) *DeleteCharacterRelationshipUseCase {
	return &DeleteCharacterRelationshipUseCase{
		characterRelationshipRepo: characterRelationshipRepo,
		logger:                    logger,
	}
}

// DeleteCharacterRelationshipInput represents the input for deleting a character relationship
type DeleteCharacterRelationshipInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes a character relationship
func (uc *DeleteCharacterRelationshipUseCase) Execute(ctx context.Context, input DeleteCharacterRelationshipInput) error {
	if err := uc.characterRelationshipRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete character relationship", "error", err, "relationship_id", input.ID)
		return err
	}

	uc.logger.Info("character relationship deleted", "relationship_id", input.ID)

	return nil
}

