package lore

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// RemoveReferenceUseCase handles removing a reference from a lore
type RemoveReferenceUseCase struct {
	loreReferenceRepo repositories.LoreReferenceRepository
	logger            logger.Logger
}

// NewRemoveReferenceUseCase creates a new RemoveReferenceUseCase
func NewRemoveReferenceUseCase(
	loreReferenceRepo repositories.LoreReferenceRepository,
	logger logger.Logger,
) *RemoveReferenceUseCase {
	return &RemoveReferenceUseCase{
		loreReferenceRepo: loreReferenceRepo,
		logger:            logger,
	}
}

// RemoveReferenceInput represents the input for removing a reference
type RemoveReferenceInput struct {
	TenantID   uuid.UUID
	LoreID     uuid.UUID
	EntityType string
	EntityID   uuid.UUID
}

// Execute removes a reference from a lore
func (uc *RemoveReferenceUseCase) Execute(ctx context.Context, input RemoveReferenceInput) error {
	err := uc.loreReferenceRepo.DeleteByLoreAndEntity(ctx, input.TenantID, input.LoreID, input.EntityType, input.EntityID)
	if err != nil {
		uc.logger.Error("failed to remove lore reference", "error", err, "lore_id", input.LoreID)
		return err
	}

	uc.logger.Info("lore reference removed", "lore_id", input.LoreID, "entity_type", input.EntityType, "entity_id", input.EntityID)
	return nil
}

