package character_inventory

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListCharacterInventoryUseCase handles listing character inventory
type ListCharacterInventoryUseCase struct {
	inventoryRepo repositories.CharacterInventoryRepository
	logger        logger.Logger
}

// NewListCharacterInventoryUseCase creates a new ListCharacterInventoryUseCase
func NewListCharacterInventoryUseCase(
	inventoryRepo repositories.CharacterInventoryRepository,
	logger logger.Logger,
) *ListCharacterInventoryUseCase {
	return &ListCharacterInventoryUseCase{
		inventoryRepo: inventoryRepo,
		logger:        logger,
	}
}

// ListCharacterInventoryInput represents the input for listing inventory
type ListCharacterInventoryInput struct {
	TenantID     uuid.UUID
	CharacterID  uuid.UUID
	EquippedOnly bool
}

// ListCharacterInventoryOutput represents the output of listing inventory
type ListCharacterInventoryOutput struct {
	Items []*rpg.CharacterInventory
}

// Execute lists character inventory
func (uc *ListCharacterInventoryUseCase) Execute(ctx context.Context, input ListCharacterInventoryInput) (*ListCharacterInventoryOutput, error) {
	var items []*rpg.CharacterInventory
	var err error

	if input.EquippedOnly {
		items, err = uc.inventoryRepo.ListEquippedByCharacter(ctx, input.TenantID, input.CharacterID)
	} else {
		items, err = uc.inventoryRepo.ListByCharacter(ctx, input.TenantID, input.CharacterID)
	}

	if err != nil {
		uc.logger.Error("failed to list character inventory", "error", err, "character_id", input.CharacterID)
		return nil, err
	}

	return &ListCharacterInventoryOutput{
		Items: items,
	}, nil
}


