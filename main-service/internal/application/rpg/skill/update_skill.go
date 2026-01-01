package skill

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateSkillUseCase handles skill updates
type UpdateSkillUseCase struct {
	skillRepo repositories.SkillRepository
	logger    logger.Logger
}

// NewUpdateSkillUseCase creates a new UpdateSkillUseCase
func NewUpdateSkillUseCase(
	skillRepo repositories.SkillRepository,
	logger logger.Logger,
) *UpdateSkillUseCase {
	return &UpdateSkillUseCase{
		skillRepo: skillRepo,
		logger:    logger,
	}
}

// UpdateSkillInput represents the input for updating a skill
type UpdateSkillInput struct {
	ID            uuid.UUID
	Name          *string
	Category      *rpg.SkillCategory
	Type          *rpg.SkillType
	Description   *string
	Prerequisites *json.RawMessage
	MaxRank       *int
	EffectsSchema *json.RawMessage
}

// UpdateSkillOutput represents the output of updating a skill
type UpdateSkillOutput struct {
	Skill *rpg.Skill
}

// Execute updates a skill
func (uc *UpdateSkillUseCase) Execute(ctx context.Context, input UpdateSkillInput) (*UpdateSkillOutput, error) {
	// Get existing skill
	skill, err := uc.skillRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Name != nil {
		if err := skill.UpdateName(*input.Name); err != nil {
			return nil, err
		}
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

	if err := uc.skillRepo.Update(ctx, skill); err != nil {
		uc.logger.Error("failed to update skill", "error", err, "skill_id", input.ID)
		return nil, err
	}

	uc.logger.Info("skill updated", "skill_id", input.ID)

	return &UpdateSkillOutput{
		Skill: skill,
	}, nil
}

