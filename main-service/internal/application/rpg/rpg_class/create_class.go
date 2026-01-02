package rpg_class

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateRPGClassUseCase handles RPG class creation
type CreateRPGClassUseCase struct {
	classRepo    repositories.RPGClassRepository
	rpgSystemRepo repositories.RPGSystemRepository
	logger       logger.Logger
}

// NewCreateRPGClassUseCase creates a new CreateRPGClassUseCase
func NewCreateRPGClassUseCase(
	classRepo repositories.RPGClassRepository,
	rpgSystemRepo repositories.RPGSystemRepository,
	logger logger.Logger,
) *CreateRPGClassUseCase {
	return &CreateRPGClassUseCase{
		classRepo:    classRepo,
		rpgSystemRepo: rpgSystemRepo,
		logger:       logger,
	}
}

// CreateRPGClassInput represents the input for creating an RPG class
type CreateRPGClassInput struct {
	RPGSystemID  uuid.UUID
	ParentClassID *uuid.UUID
	Name         string
	Tier         *int
	Description  *string
	Requirements *json.RawMessage
	StatBonuses  *json.RawMessage
}

// CreateRPGClassOutput represents the output of creating an RPG class
type CreateRPGClassOutput struct {
	Class *rpg.RPGClass
}

// Execute creates a new RPG class
func (uc *CreateRPGClassUseCase) Execute(ctx context.Context, input CreateRPGClassInput) (*CreateRPGClassOutput, error) {
	// Validate RPG system exists
	_, err := uc.rpgSystemRepo.GetByID(ctx, input.RPGSystemID)
	if err != nil {
		return nil, err
	}

	// Validate parent class exists if provided
	if input.ParentClassID != nil {
		_, err := uc.classRepo.GetByID(ctx, *input.ParentClassID)
		if err != nil {
			return nil, err
		}
	}

	// Create class
	class, err := rpg.NewRPGClass(input.RPGSystemID, input.Name)
	if err != nil {
		return nil, err
	}

	if input.ParentClassID != nil {
		class.SetParentClass(input.ParentClassID)
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

	if err := uc.classRepo.Create(ctx, class); err != nil {
		uc.logger.Error("failed to create RPG class", "error", err, "rpg_system_id", input.RPGSystemID)
		return nil, err
	}

	uc.logger.Info("RPG class created", "class_id", class.ID, "name", class.Name)

	return &CreateRPGClassOutput{
		Class: class,
	}, nil
}


