package rpg_class

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateRPGClassUseCase handles RPG class updates
type UpdateRPGClassUseCase struct {
	classRepo repositories.RPGClassRepository
	logger    logger.Logger
}

// NewUpdateRPGClassUseCase creates a new UpdateRPGClassUseCase
func NewUpdateRPGClassUseCase(
	classRepo repositories.RPGClassRepository,
	logger logger.Logger,
) *UpdateRPGClassUseCase {
	return &UpdateRPGClassUseCase{
		classRepo: classRepo,
		logger:    logger,
	}
}

// UpdateRPGClassInput represents the input for updating an RPG class
type UpdateRPGClassInput struct {
	ID            uuid.UUID
	ParentClassID *uuid.UUID
	Name          *string
	Tier          *int
	Description   *string
	Requirements  *json.RawMessage
	StatBonuses   *json.RawMessage
}

// UpdateRPGClassOutput represents the output of updating an RPG class
type UpdateRPGClassOutput struct {
	Class *rpg.RPGClass
}

// Execute updates an RPG class
func (uc *UpdateRPGClassUseCase) Execute(ctx context.Context, input UpdateRPGClassInput) (*UpdateRPGClassOutput, error) {
	// Get existing class
	class, err := uc.classRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// Validate parent class exists if provided
	if input.ParentClassID != nil {
		_, err := uc.classRepo.GetByID(ctx, *input.ParentClassID)
		if err != nil {
			return nil, err
		}
		class.SetParentClass(input.ParentClassID)
	}

	// Update fields if provided
	if input.Name != nil {
		if err := class.UpdateName(*input.Name); err != nil {
			return nil, err
		}
	}
	if input.Tier != nil {
		if err := class.UpdateTier(*input.Tier); err != nil {
			return nil, err
		}
	}
	if input.Description != nil {
		class.UpdateDescription(input.Description)
	}
	if input.Requirements != nil {
		if err := class.UpdateRequirements(input.Requirements); err != nil {
			return nil, err
		}
	}
	if input.StatBonuses != nil {
		if err := class.UpdateStatBonuses(input.StatBonuses); err != nil {
			return nil, err
		}
	}

	if err := class.Validate(); err != nil {
		return nil, err
	}

	if err := uc.classRepo.Update(ctx, class); err != nil {
		uc.logger.Error("failed to update RPG class", "error", err, "class_id", input.ID)
		return nil, err
	}

	uc.logger.Info("RPG class updated", "class_id", input.ID)

	return &UpdateRPGClassOutput{
		Class: class,
	}, nil
}

