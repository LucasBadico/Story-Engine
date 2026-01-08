package character

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/queue"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateCharacterUseCase handles character updates
type UpdateCharacterUseCase struct {
	characterRepo repositories.CharacterRepository
	archetypeRepo repositories.ArchetypeRepository
	worldRepo     repositories.WorldRepository
	auditLogRepo  repositories.AuditLogRepository
	ingestionQueue queue.IngestionQueue
	logger        logger.Logger
}

// NewUpdateCharacterUseCase creates a new UpdateCharacterUseCase
func NewUpdateCharacterUseCase(
	characterRepo repositories.CharacterRepository,
	archetypeRepo repositories.ArchetypeRepository,
	worldRepo repositories.WorldRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *UpdateCharacterUseCase {
	return &UpdateCharacterUseCase{
		characterRepo: characterRepo,
		archetypeRepo: archetypeRepo,
		worldRepo:     worldRepo,
		auditLogRepo:  auditLogRepo,
		ingestionQueue: nil,
		logger:        logger,
	}
}

func (uc *UpdateCharacterUseCase) SetIngestionQueue(queue queue.IngestionQueue) {
	uc.ingestionQueue = queue
}

// UpdateCharacterInput represents the input for updating a character
type UpdateCharacterInput struct {
	TenantID    uuid.UUID
	ID          uuid.UUID
	Name        *string
	Description *string
	ArchetypeID *uuid.UUID
}

// UpdateCharacterOutput represents the output of updating a character
type UpdateCharacterOutput struct {
	Character *world.Character
}

// Execute updates a character
func (uc *UpdateCharacterUseCase) Execute(ctx context.Context, input UpdateCharacterInput) (*UpdateCharacterOutput, error) {
	c, err := uc.characterRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		if err := c.UpdateName(*input.Name); err != nil {
			return nil, &platformerrors.ValidationError{
				Field:   "name",
				Message: err.Error(),
			}
		}
	}
	if input.Description != nil {
		c.UpdateDescription(*input.Description)
	}
	if input.ArchetypeID != nil {
		// Validate archetype exists if provided
		if *input.ArchetypeID != uuid.Nil {
			_, err := uc.archetypeRepo.GetByID(ctx, input.TenantID, *input.ArchetypeID)
			if err != nil {
				return nil, err
			}
		}
		c.SetArchetype(input.ArchetypeID)
	}

	if err := c.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "character",
			Message: err.Error(),
		}
	}

	if err := uc.characterRepo.Update(ctx, c); err != nil {
		uc.logger.Error("failed to update character", "error", err, "character_id", input.ID)
		return nil, err
	}

	w, _ := uc.worldRepo.GetByID(ctx, input.TenantID, c.WorldID)
	auditLog := audit.NewAuditLog(
		w.TenantID,
		nil,
		audit.ActionUpdate,
		audit.EntityTypeCharacter,
		c.ID,
		map[string]interface{}{
			"name": c.Name,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("character updated", "character_id", c.ID, "name", c.Name)
	uc.enqueueIngestion(ctx, input.TenantID, c.ID)

	return &UpdateCharacterOutput{
		Character: c,
	}, nil
}

func (uc *UpdateCharacterUseCase) enqueueIngestion(ctx context.Context, tenantID uuid.UUID, characterID uuid.UUID) {
	if uc.ingestionQueue == nil {
		return
	}
	if err := uc.ingestionQueue.Push(ctx, tenantID, "character", characterID); err != nil {
		uc.logger.Error("failed to enqueue character ingestion", "error", err, "character_id", characterID, "tenant_id", tenantID)
	}
}

