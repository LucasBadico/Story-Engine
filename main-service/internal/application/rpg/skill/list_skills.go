package skill

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListSkillsUseCase handles listing skills
type ListSkillsUseCase struct {
	skillRepo repositories.SkillRepository
	logger    logger.Logger
}

// NewListSkillsUseCase creates a new ListSkillsUseCase
func NewListSkillsUseCase(
	skillRepo repositories.SkillRepository,
	logger logger.Logger,
) *ListSkillsUseCase {
	return &ListSkillsUseCase{
		skillRepo: skillRepo,
		logger:    logger,
	}
}

// ListSkillsInput represents the input for listing skills
type ListSkillsInput struct {
	RPGSystemID uuid.UUID
}

// ListSkillsOutput represents the output of listing skills
type ListSkillsOutput struct {
	Skills []*rpg.Skill
}

// Execute lists skills for an RPG system
func (uc *ListSkillsUseCase) Execute(ctx context.Context, input ListSkillsInput) (*ListSkillsOutput, error) {
	skills, err := uc.skillRepo.ListBySystem(ctx, input.RPGSystemID)
	if err != nil {
		uc.logger.Error("failed to list skills", "error", err, "rpg_system_id", input.RPGSystemID)
		return nil, err
	}

	return &ListSkillsOutput{
		Skills: skills,
	}, nil
}

