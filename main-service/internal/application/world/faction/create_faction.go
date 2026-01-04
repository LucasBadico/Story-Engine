package faction

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateFactionUseCase handles faction creation
type CreateFactionUseCase struct {
	factionRepo  repositories.FactionRepository
	worldRepo    repositories.WorldRepository
	auditLogRepo repositories.AuditLogRepository
	logger       logger.Logger
}

// NewCreateFactionUseCase creates a new CreateFactionUseCase
func NewCreateFactionUseCase(
	factionRepo repositories.FactionRepository,
	worldRepo repositories.WorldRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *CreateFactionUseCase {
	return &CreateFactionUseCase{
		factionRepo:  factionRepo,
		worldRepo:    worldRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// CreateFactionInput represents the input for creating a faction
type CreateFactionInput struct {
	TenantID    uuid.UUID
	WorldID     uuid.UUID
	ParentID    *uuid.UUID
	Name        string
	Type        *string
	Description string
	Beliefs     string
	Structure   string
	Symbols     string
}

// CreateFactionOutput represents the output of creating a faction
type CreateFactionOutput struct {
	Faction *world.Faction
}

// Execute creates a new faction
func (uc *CreateFactionUseCase) Execute(ctx context.Context, input CreateFactionInput) (*CreateFactionOutput, error) {
	// Validate world exists
	_, err := uc.worldRepo.GetByID(ctx, input.TenantID, input.WorldID)
	if err != nil {
		return nil, err
	}

	// Validate parent exists if provided
	if input.ParentID != nil {
		parent, err := uc.factionRepo.GetByID(ctx, input.TenantID, *input.ParentID)
		if err != nil {
			return nil, err
		}
		if parent.WorldID != input.WorldID {
			return nil, &platformerrors.ValidationError{
				Field:   "parent_id",
				Message: "parent faction must belong to the same world",
			}
		}
	}

	newFaction, err := world.NewFaction(input.TenantID, input.WorldID, input.Name, input.ParentID)
	if err != nil {
		if errors.Is(err, world.ErrFactionNameRequired) {
			return nil, &platformerrors.ValidationError{
				Field:   "name",
				Message: err.Error(),
			}
		}
		return nil, err
	}

	if input.Type != nil {
		newFaction.UpdateType(input.Type)
	}
	if input.Description != "" {
		newFaction.UpdateDescription(input.Description)
	}
	if input.Beliefs != "" {
		newFaction.UpdateBeliefs(input.Beliefs)
	}
	if input.Structure != "" {
		newFaction.UpdateStructure(input.Structure)
	}
	if input.Symbols != "" {
		newFaction.UpdateSymbols(input.Symbols)
	}

	if err := newFaction.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "faction",
			Message: err.Error(),
		}
	}

	if err := uc.factionRepo.Create(ctx, newFaction); err != nil {
		uc.logger.Error("failed to create faction", "error", err, "name", input.Name)
		return nil, err
	}

	auditLog := audit.NewAuditLog(
		input.TenantID,
		nil,
		audit.ActionCreate,
		audit.EntityTypeFaction,
		newFaction.ID,
		map[string]interface{}{
			"name":           newFaction.Name,
			"world_id":       newFaction.WorldID.String(),
			"hierarchy_level": newFaction.HierarchyLevel,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("faction created", "faction_id", newFaction.ID, "name", newFaction.Name)

	return &CreateFactionOutput{
		Faction: newFaction,
	}, nil
}

