package rpg_system

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateRPGSystemUseCase handles RPG system creation
type CreateRPGSystemUseCase struct {
	rpgSystemRepo repositories.RPGSystemRepository
	tenantRepo    repositories.TenantRepository
	logger        logger.Logger
}

// NewCreateRPGSystemUseCase creates a new CreateRPGSystemUseCase
func NewCreateRPGSystemUseCase(
	rpgSystemRepo repositories.RPGSystemRepository,
	tenantRepo repositories.TenantRepository,
	logger logger.Logger,
) *CreateRPGSystemUseCase {
	return &CreateRPGSystemUseCase{
		rpgSystemRepo: rpgSystemRepo,
		tenantRepo:    tenantRepo,
		logger:        logger,
	}
}

// CreateRPGSystemInput represents the input for creating an RPG system
type CreateRPGSystemInput struct {
	TenantID          *uuid.UUID
	Name              string
	Description       *string
	BaseStatsSchema   json.RawMessage
	DerivedStatsSchema *json.RawMessage
	ProgressionSchema *json.RawMessage
}

// CreateRPGSystemOutput represents the output of creating an RPG system
type CreateRPGSystemOutput struct {
	RPGSystem *rpg.RPGSystem
}

// Execute creates a new RPG system
func (uc *CreateRPGSystemUseCase) Execute(ctx context.Context, input CreateRPGSystemInput) (*CreateRPGSystemOutput, error) {
	// Validate tenant exists if provided
	if input.TenantID != nil {
		_, err := uc.tenantRepo.GetByID(ctx, *input.TenantID)
		if err != nil {
			return nil, err
		}
	}

	// Create RPG system
	system, err := rpg.NewRPGSystem(input.TenantID, input.Name, input.BaseStatsSchema)
	if err != nil {
		return nil, err
	}

	if input.Description != nil {
		system.UpdateDescription(input.Description)
	}
	if input.DerivedStatsSchema != nil || input.ProgressionSchema != nil {
		if err := system.UpdateSchemas(input.BaseStatsSchema, input.DerivedStatsSchema, input.ProgressionSchema); err != nil {
			return nil, err
		}
	}

	if err := system.Validate(); err != nil {
		return nil, err
	}

	if err := uc.rpgSystemRepo.Create(ctx, system); err != nil {
		uc.logger.Error("failed to create RPG system", "error", err)
		return nil, err
	}

	uc.logger.Info("RPG system created", "rpg_system_id", system.ID, "name", system.Name)

	return &CreateRPGSystemOutput{
		RPGSystem: system,
	}, nil
}


