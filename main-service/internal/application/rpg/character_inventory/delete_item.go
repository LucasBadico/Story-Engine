package character_inventory

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteCharacterInventoryUseCase handles deleting an item from character inventory
type DeleteCharacterInventoryUseCase struct {
	inventoryRepo repositories.CharacterInventoryRepository
	logger        logger.Logger
}

// NewDeleteCharacterInventoryUseCase creates a new DeleteCharacterInventoryUseCase
func NewDeleteCharacterInventoryUseCase(
	inventoryRepo repositories.CharacterInventoryRepository,
	logger logger.Logger,
) *DeleteCharacterInventoryUseCase {
	return &DeleteCharacterInventoryUseCase{
		inventoryRepo: inventoryRepo,
		logger:        logger,
	}
}

// DeleteCharacterInventoryInput represents the input for deleting inventory
type DeleteCharacterInventoryInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes an item from character inventory
func (uc *DeleteCharacterInventoryUseCase) Execute(ctx context.Context, input DeleteCharacterInventoryInput) error {
	if err := uc.inventoryRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete character inventory", "error", err, "id", input.ID)
		return err
	}

	uc.logger.Info("item removed from inventory", "id", input.ID)

	return nil
}


