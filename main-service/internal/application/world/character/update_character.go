package character

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateCharacterUseCase handles character updates
type UpdateCharacterUseCase struct {
	characterRepo repositories.CharacterRepository
	archetypeRepo repositories.ArchetypeRepository
	worldRepo     repositories.WorldRepository
	auditLogRepo  repositories.AuditLogRepository
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
		logger:        logger,
	}
}

// UpdateCharacterInput represents the input for updating a character
type UpdateCharacterInput struct {
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
	c, err := uc.characterRepo.GetByID(ctx, input.ID)
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
			_, err := uc.archetypeRepo.GetByID(ctx, *input.ArchetypeID)
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

	w, _ := uc.worldRepo.GetByID(ctx, c.WorldID)
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

	return &UpdateCharacterOutput{
		Character: c,
	}, nil
}


