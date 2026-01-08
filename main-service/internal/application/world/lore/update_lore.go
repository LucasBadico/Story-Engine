package lore

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/queue"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateLoreUseCase handles lore updates
type UpdateLoreUseCase struct {
	loreRepo     repositories.LoreRepository
	auditLogRepo repositories.AuditLogRepository
	ingestionQueue queue.IngestionQueue
	logger       logger.Logger
}

// NewUpdateLoreUseCase creates a new UpdateLoreUseCase
func NewUpdateLoreUseCase(
	loreRepo repositories.LoreRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *UpdateLoreUseCase {
	return &UpdateLoreUseCase{
		loreRepo:     loreRepo,
		auditLogRepo: auditLogRepo,
		ingestionQueue: nil,
		logger:       logger,
	}
}

func (uc *UpdateLoreUseCase) SetIngestionQueue(queue queue.IngestionQueue) {
	uc.ingestionQueue = queue
}

// UpdateLoreInput represents the input for updating a lore
type UpdateLoreInput struct {
	TenantID     uuid.UUID
	ID           uuid.UUID
	Name         *string
	Category     *string
	Description  *string
	Rules        *string
	Limitations  *string
	Requirements *string
}

// UpdateLoreOutput represents the output of updating a lore
type UpdateLoreOutput struct {
	Lore *world.Lore
}

// Execute updates a lore
func (uc *UpdateLoreUseCase) Execute(ctx context.Context, input UpdateLoreInput) (*UpdateLoreOutput, error) {
	// Get existing lore
	lore, err := uc.loreRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Name != nil {
		if err := lore.UpdateName(*input.Name); err != nil {
			return nil, err
		}
	}
	if input.Category != nil {
		lore.UpdateCategory(input.Category)
	}
	if input.Description != nil {
		lore.UpdateDescription(*input.Description)
	}
	if input.Rules != nil {
		lore.UpdateRules(*input.Rules)
	}
	if input.Limitations != nil {
		lore.UpdateLimitations(*input.Limitations)
	}
	if input.Requirements != nil {
		lore.UpdateRequirements(*input.Requirements)
	}

	if err := lore.Validate(); err != nil {
		return nil, err
	}

	if err := uc.loreRepo.Update(ctx, lore); err != nil {
		uc.logger.Error("failed to update lore", "error", err, "lore_id", input.ID)
		return nil, err
	}

	// Log audit event
	auditLog := audit.NewAuditLog(
		lore.TenantID,
		nil,
		audit.ActionUpdate,
		audit.EntityTypeLore,
		lore.ID,
		map[string]interface{}{},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Error("failed to create audit log", "error", err)
	}

	uc.logger.Info("lore updated", "lore_id", input.ID)
	uc.enqueueIngestion(ctx, input.TenantID, lore.ID)

	return &UpdateLoreOutput{
		Lore: lore,
	}, nil
}

func (uc *UpdateLoreUseCase) enqueueIngestion(ctx context.Context, tenantID uuid.UUID, loreID uuid.UUID) {
	if uc.ingestionQueue == nil {
		return
	}
	if err := uc.ingestionQueue.Push(ctx, tenantID, "lore", loreID); err != nil {
		uc.logger.Error("failed to enqueue lore ingestion", "error", err, "lore_id", loreID, "tenant_id", tenantID)
	}
}
