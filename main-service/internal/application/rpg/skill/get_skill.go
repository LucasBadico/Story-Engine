package skill

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetSkillUseCase handles retrieving a skill
type GetSkillUseCase struct {
	skillRepo repositories.SkillRepository
	logger    logger.Logger
}

// NewGetSkillUseCase creates a new GetSkillUseCase
func NewGetSkillUseCase(
	skillRepo repositories.SkillRepository,
	logger logger.Logger,
) *GetSkillUseCase {
	return &GetSkillUseCase{
		skillRepo: skillRepo,
		logger:    logger,
	}
}

// GetSkillInput represents the input for getting a skill
type GetSkillInput struct {
	ID uuid.UUID
}

// GetSkillOutput represents the output of getting a skill
type GetSkillOutput struct {
	Skill *rpg.Skill
}

// Execute retrieves a skill by ID
func (uc *GetSkillUseCase) Execute(ctx context.Context, input GetSkillInput) (*GetSkillOutput, error) {
	skill, err := uc.skillRepo.GetByID(ctx, input.ID)
	if err != nil {
		uc.logger.Error("failed to get skill", "error", err, "skill_id", input.ID)
		return nil, err
	}

	return &GetSkillOutput{
		Skill: skill,
	}, nil
}

