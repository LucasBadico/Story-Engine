package lore

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/queue"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateLoreUseCase handles lore creation
type CreateLoreUseCase struct {
	loreRepo     repositories.LoreRepository
	worldRepo    repositories.WorldRepository
	auditLogRepo repositories.AuditLogRepository
	ingestionQueue queue.IngestionQueue
	logger       logger.Logger
}

// NewCreateLoreUseCase creates a new CreateLoreUseCase
func NewCreateLoreUseCase(
	loreRepo repositories.LoreRepository,
	worldRepo repositories.WorldRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *CreateLoreUseCase {
	return &CreateLoreUseCase{
		loreRepo:     loreRepo,
		worldRepo:    worldRepo,
		auditLogRepo: auditLogRepo,
		ingestionQueue: nil,
		logger:       logger,
	}
}

func (uc *CreateLoreUseCase) SetIngestionQueue(queue queue.IngestionQueue) {
	uc.ingestionQueue = queue
}

// CreateLoreInput represents the input for creating a lore
type CreateLoreInput struct {
	TenantID    uuid.UUID
	WorldID     uuid.UUID
	ParentID    *uuid.UUID
	Name        string
	Category    *string
	Description string
	Rules       string
	Limitations string
	Requirements string
}

// CreateLoreOutput represents the output of creating a lore
type CreateLoreOutput struct {
	Lore *world.Lore
}

// Execute creates a new lore
func (uc *CreateLoreUseCase) Execute(ctx context.Context, input CreateLoreInput) (*CreateLoreOutput, error) {
	// Validate world exists
	_, err := uc.worldRepo.GetByID(ctx, input.TenantID, input.WorldID)
	if err != nil {
		return nil, err
	}

	// Validate parent exists if provided
	if input.ParentID != nil {
		parent, err := uc.loreRepo.GetByID(ctx, input.TenantID, *input.ParentID)
		if err != nil {
			return nil, err
		}
		if parent.WorldID != input.WorldID {
			return nil, &platformerrors.ValidationError{
				Field:   "parent_id",
				Message: "parent lore must belong to the same world",
			}
		}
	}

	newLore, err := world.NewLore(input.TenantID, input.WorldID, input.Name, input.ParentID)
	if err != nil {
		if errors.Is(err, world.ErrLoreNameRequired) {
			return nil, &platformerrors.ValidationError{
				Field:   "name",
				Message: err.Error(),
			}
		}
		return nil, err
	}

	if input.Category != nil {
		newLore.UpdateCategory(input.Category)
	}
	if input.Description != "" {
		newLore.UpdateDescription(input.Description)
	}
	if input.Rules != "" {
		newLore.UpdateRules(input.Rules)
	}
	if input.Limitations != "" {
		newLore.UpdateLimitations(input.Limitations)
	}
	if input.Requirements != "" {
		newLore.UpdateRequirements(input.Requirements)
	}

	if err := newLore.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "lore",
			Message: err.Error(),
		}
	}

	if err := uc.loreRepo.Create(ctx, newLore); err != nil {
		uc.logger.Error("failed to create lore", "error", err, "name", input.Name)
		return nil, err
	}

	auditLog := audit.NewAuditLog(
		input.TenantID,
		nil,
		audit.ActionCreate,
		audit.EntityTypeLore,
		newLore.ID,
		map[string]interface{}{
			"name":           newLore.Name,
			"world_id":       newLore.WorldID.String(),
			"hierarchy_level": newLore.HierarchyLevel,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("lore created", "lore_id", newLore.ID, "name", newLore.Name)
	uc.enqueueIngestion(ctx, input.TenantID, newLore.ID)

	return &CreateLoreOutput{
		Lore: newLore,
	}, nil
}

func (uc *CreateLoreUseCase) enqueueIngestion(ctx context.Context, tenantID uuid.UUID, loreID uuid.UUID) {
	if uc.ingestionQueue == nil {
		return
	}
	if err := uc.ingestionQueue.Push(ctx, tenantID, "lore", loreID); err != nil {
		uc.logger.Error("failed to enqueue lore ingestion", "error", err, "lore_id", loreID, "tenant_id", tenantID)
	}
}
