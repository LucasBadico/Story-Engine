package relation

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteRelationUseCase handles deleting entity relations
type DeleteRelationUseCase struct {
	relationRepo repositories.EntityRelationRepository
	logger       logger.Logger
}

// NewDeleteRelationUseCase creates a new DeleteRelationUseCase
func NewDeleteRelationUseCase(
	relationRepo repositories.EntityRelationRepository,
	logger logger.Logger,
) *DeleteRelationUseCase {
	return &DeleteRelationUseCase{
		relationRepo: relationRepo,
		logger:       logger,
	}
}

// DeleteRelationInput represents the input for deleting a relation
type DeleteRelationInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes an entity relation (and its mirror if exists)
// Note: Embedding chunk deletion is handled automatically by llm-gateway-service
// which periodically verifies entity existence and removes orphaned chunks
func (uc *DeleteRelationUseCase) Execute(ctx context.Context, input DeleteRelationInput) error {
	if err := uc.relationRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete relation", "error", err, "relation_id", input.ID)
		return err
	}

	uc.logger.Info("relation deleted", "relation_id", input.ID)

	return nil
}

