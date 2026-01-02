package inventory_slot

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateInventorySlotUseCase handles inventory slot creation
type CreateInventorySlotUseCase struct {
	slotRepo    repositories.InventorySlotRepository
	rpgSystemRepo repositories.RPGSystemRepository
	logger      logger.Logger
}

// NewCreateInventorySlotUseCase creates a new CreateInventorySlotUseCase
func NewCreateInventorySlotUseCase(
	slotRepo repositories.InventorySlotRepository,
	rpgSystemRepo repositories.RPGSystemRepository,
	logger logger.Logger,
) *CreateInventorySlotUseCase {
	return &CreateInventorySlotUseCase{
		slotRepo:    slotRepo,
		rpgSystemRepo: rpgSystemRepo,
		logger:      logger,
	}
}

// CreateInventorySlotInput represents the input for creating an inventory slot
type CreateInventorySlotInput struct {
	RPGSystemID uuid.UUID
	Name        string
	SlotType    *rpg.SlotType
}

// CreateInventorySlotOutput represents the output of creating an inventory slot
type CreateInventorySlotOutput struct {
	Slot *rpg.InventorySlot
}

// Execute creates a new inventory slot
func (uc *CreateInventorySlotUseCase) Execute(ctx context.Context, input CreateInventorySlotInput) (*CreateInventorySlotOutput, error) {
	// Validate RPG system exists
	_, err := uc.rpgSystemRepo.GetByID(ctx, input.RPGSystemID)
	if err != nil {
		return nil, err
	}

	// Create slot
	slot, err := rpg.NewInventorySlot(input.RPGSystemID, input.Name)
	if err != nil {
		return nil, err
	}

	if input.SlotType != nil {
		slot.UpdateSlotType(input.SlotType)
	}

	if err := slot.Validate(); err != nil {
		return nil, err
	}

	if err := uc.slotRepo.Create(ctx, slot); err != nil {
		uc.logger.Error("failed to create inventory slot", "error", err, "rpg_system_id", input.RPGSystemID)
		return nil, err
	}

	uc.logger.Info("inventory slot created", "slot_id", slot.ID, "name", slot.Name)

	return &CreateInventorySlotOutput{
		Slot: slot,
	}, nil
}

