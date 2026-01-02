package skill

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateSkillUseCase handles skill creation
type CreateSkillUseCase struct {
	skillRepo    repositories.SkillRepository
	rpgSystemRepo repositories.RPGSystemRepository
	logger       logger.Logger
}

// NewCreateSkillUseCase creates a new CreateSkillUseCase
func NewCreateSkillUseCase(
	skillRepo repositories.SkillRepository,
	rpgSystemRepo repositories.RPGSystemRepository,
	logger logger.Logger,
) *CreateSkillUseCase {
	return &CreateSkillUseCase{
		skillRepo:    skillRepo,
		rpgSystemRepo: rpgSystemRepo,
		logger:       logger,
	}
}

// CreateSkillInput represents the input for creating a skill
type CreateSkillInput struct {
	RPGSystemID   uuid.UUID
	Name          string
	Category      *rpg.SkillCategory
	Type          *rpg.SkillType
	Description   *string
	Prerequisites *json.RawMessage
	MaxRank       *int
	EffectsSchema *json.RawMessage
}

// CreateSkillOutput represents the output of creating a skill
type CreateSkillOutput struct {
	Skill *rpg.Skill
}

// Execute creates a new skill
func (uc *CreateSkillUseCase) Execute(ctx context.Context, input CreateSkillInput) (*CreateSkillOutput, error) {
	// Validate RPG system exists
	_, err := uc.rpgSystemRepo.GetByID(ctx, input.RPGSystemID)
	if err != nil {
		return nil, err
	}

	// Create skill
	skill, err := rpg.NewSkill(input.RPGSystemID, input.Name)
	if err != nil {
		return nil, err
	}

	if input.Category != nil {
		skill.UpdateCategory(input.Category)
	}
	if input.Type != nil {
		skill.UpdateType(input.Type)
	}
	if input.Description != nil {
		skill.UpdateDescription(input.Description)
	}
	if input.Prerequisites != nil {
		if err := skill.UpdatePrerequisites(input.Prerequisites); err != nil {
			return nil, err
		}
	}
	if input.MaxRank != nil {
		if err := skill.UpdateMaxRank(*input.MaxRank); err != nil {
			return nil, err
		}
	}
	if input.EffectsSchema != nil {
		if err := skill.UpdateEffectsSchema(input.EffectsSchema); err != nil {
			return nil, err
		}
	}

	if err := skill.Validate(); err != nil {
		return nil, err
	}

	if err := uc.skillRepo.Create(ctx, skill); err != nil {
		uc.logger.Error("failed to create skill", "error", err, "rpg_system_id", input.RPGSystemID)
		return nil, err
	}

	uc.logger.Info("skill created", "skill_id", skill.ID, "name", skill.Name)

	return &CreateSkillOutput{
		Skill: skill,
	}, nil
}


