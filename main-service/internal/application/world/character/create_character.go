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

// CreateCharacterUseCase handles character creation
type CreateCharacterUseCase struct {
	characterRepo repositories.CharacterRepository
	worldRepo     repositories.WorldRepository
	archetypeRepo repositories.ArchetypeRepository
	auditLogRepo  repositories.AuditLogRepository
	ingestionQueue queue.IngestionQueue
	logger        logger.Logger
}

// NewCreateCharacterUseCase creates a new CreateCharacterUseCase
func NewCreateCharacterUseCase(
	characterRepo repositories.CharacterRepository,
	worldRepo repositories.WorldRepository,
	archetypeRepo repositories.ArchetypeRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *CreateCharacterUseCase {
	return &CreateCharacterUseCase{
		characterRepo: characterRepo,
		worldRepo:     worldRepo,
		archetypeRepo: archetypeRepo,
		auditLogRepo:  auditLogRepo,
		ingestionQueue: nil,
		logger:        logger,
	}
}

func (uc *CreateCharacterUseCase) SetIngestionQueue(queue queue.IngestionQueue) {
	uc.ingestionQueue = queue
}

// CreateCharacterInput represents the input for creating a character
type CreateCharacterInput struct {
	TenantID    uuid.UUID
	WorldID     uuid.UUID
	ArchetypeID *uuid.UUID
	Name        string
	Description string
}

// CreateCharacterOutput represents the output of creating a character
type CreateCharacterOutput struct {
	Character *world.Character
}

// Execute creates a new character
func (uc *CreateCharacterUseCase) Execute(ctx context.Context, input CreateCharacterInput) (*CreateCharacterOutput, error) {
	// Validate world exists
	w, err := uc.worldRepo.GetByID(ctx, input.TenantID, input.WorldID)
	if err != nil {
		return nil, err
	}

	// Validate archetype exists if provided
	if input.ArchetypeID != nil {
		_, err := uc.archetypeRepo.GetByID(ctx, input.TenantID, *input.ArchetypeID)
		if err != nil {
			return nil, err
		}
	}

	newCharacter, err := world.NewCharacter(input.TenantID, input.WorldID, input.Name)
	if err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "name",
			Message: err.Error(),
		}
	}

	if input.ArchetypeID != nil {
		newCharacter.SetArchetype(input.ArchetypeID)
	}
	if input.Description != "" {
		newCharacter.UpdateDescription(input.Description)
	}

	if err := newCharacter.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "character",
			Message: err.Error(),
		}
	}

	if err := uc.characterRepo.Create(ctx, newCharacter); err != nil {
		uc.logger.Error("failed to create character", "error", err, "name", input.Name)
		return nil, err
	}

	auditLog := audit.NewAuditLog(
		w.TenantID,
		nil,
		audit.ActionCreate,
		audit.EntityTypeCharacter,
		newCharacter.ID,
		map[string]interface{}{
			"name":     newCharacter.Name,
			"world_id": newCharacter.WorldID.String(),
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("character created", "character_id", newCharacter.ID, "name", newCharacter.Name)
	uc.enqueueIngestion(ctx, input.TenantID, newCharacter.ID)

	return &CreateCharacterOutput{
		Character: newCharacter,
	}, nil
}

func (uc *CreateCharacterUseCase) enqueueIngestion(ctx context.Context, tenantID uuid.UUID, characterID uuid.UUID) {
	if uc.ingestionQueue == nil {
		return
	}
	if err := uc.ingestionQueue.Push(ctx, tenantID, "character", characterID); err != nil {
		uc.logger.Error("failed to enqueue character ingestion", "error", err, "character_id", characterID, "tenant_id", tenantID)
	}
}

