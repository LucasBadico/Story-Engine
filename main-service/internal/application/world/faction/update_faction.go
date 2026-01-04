package faction

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateFactionUseCase handles faction updates
type UpdateFactionUseCase struct {
	factionRepo  repositories.FactionRepository
	auditLogRepo repositories.AuditLogRepository
	logger       logger.Logger
}

// NewUpdateFactionUseCase creates a new UpdateFactionUseCase
func NewUpdateFactionUseCase(
	factionRepo repositories.FactionRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *UpdateFactionUseCase {
	return &UpdateFactionUseCase{
		factionRepo:  factionRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// UpdateFactionInput represents the input for updating a faction
type UpdateFactionInput struct {
	TenantID    uuid.UUID
	ID          uuid.UUID
	Name        *string
	Type        *string
	Description *string
	Beliefs     *string
	Structure   *string
	Symbols     *string
}

// UpdateFactionOutput represents the output of updating a faction
type UpdateFactionOutput struct {
	Faction *world.Faction
}

// Execute updates a faction
func (uc *UpdateFactionUseCase) Execute(ctx context.Context, input UpdateFactionInput) (*UpdateFactionOutput, error) {
	// Get existing faction
	faction, err := uc.factionRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Name != nil {
		if err := faction.UpdateName(*input.Name); err != nil {
			return nil, err
		}
	}
	if input.Type != nil {
		faction.UpdateType(input.Type)
	}
	if input.Description != nil {
		faction.UpdateDescription(*input.Description)
	}
	if input.Beliefs != nil {
		faction.UpdateBeliefs(*input.Beliefs)
	}
	if input.Structure != nil {
		faction.UpdateStructure(*input.Structure)
	}
	if input.Symbols != nil {
		faction.UpdateSymbols(*input.Symbols)
	}

	if err := faction.Validate(); err != nil {
		return nil, err
	}

	if err := uc.factionRepo.Update(ctx, faction); err != nil {
		uc.logger.Error("failed to update faction", "error", err, "faction_id", input.ID)
		return nil, err
	}

	// Log audit event
	auditLog := audit.NewAuditLog(
		faction.TenantID,
		nil,
		audit.ActionUpdate,
		audit.EntityTypeFaction,
		faction.ID,
		map[string]interface{}{},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Error("failed to create audit log", "error", err)
	}

	uc.logger.Info("faction updated", "faction_id", input.ID)

	return &UpdateFactionOutput{
		Faction: faction,
	}, nil
}

