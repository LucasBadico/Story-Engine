package inventory_slot

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListInventorySlotsUseCase handles listing inventory slots
type ListInventorySlotsUseCase struct {
	slotRepo repositories.InventorySlotRepository
	logger   logger.Logger
}

// NewListInventorySlotsUseCase creates a new ListInventorySlotsUseCase
func NewListInventorySlotsUseCase(
	slotRepo repositories.InventorySlotRepository,
	logger logger.Logger,
) *ListInventorySlotsUseCase {
	return &ListInventorySlotsUseCase{
		slotRepo: slotRepo,
		logger:   logger,
	}
}

// ListInventorySlotsInput represents the input for listing slots
type ListInventorySlotsInput struct {
	TenantID    uuid.UUID
	RPGSystemID uuid.UUID
}

// ListInventorySlotsOutput represents the output of listing slots
type ListInventorySlotsOutput struct {
	Slots []*rpg.InventorySlot
}

// Execute lists inventory slots for an RPG system
func (uc *ListInventorySlotsUseCase) Execute(ctx context.Context, input ListInventorySlotsInput) (*ListInventorySlotsOutput, error) {
	slots, err := uc.slotRepo.ListBySystem(ctx, input.TenantID, input.RPGSystemID)
	if err != nil {
		uc.logger.Error("failed to list inventory slots", "error", err, "rpg_system_id", input.RPGSystemID)
		return nil, err
	}

	return &ListInventorySlotsOutput{
		Slots: slots,
	}, nil
}


