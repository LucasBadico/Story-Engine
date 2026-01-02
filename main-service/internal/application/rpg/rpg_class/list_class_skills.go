package rpg_class

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListClassSkillsUseCase handles listing skills for a class
type ListClassSkillsUseCase struct {
	classSkillRepo repositories.RPGClassSkillRepository
	logger         logger.Logger
}

// NewListClassSkillsUseCase creates a new ListClassSkillsUseCase
func NewListClassSkillsUseCase(
	classSkillRepo repositories.RPGClassSkillRepository,
	logger logger.Logger,
) *ListClassSkillsUseCase {
	return &ListClassSkillsUseCase{
		classSkillRepo: classSkillRepo,
		logger:         logger,
	}
}

// ListClassSkillsInput represents the input for listing class skills
type ListClassSkillsInput struct {
	ClassID uuid.UUID
}

// ListClassSkillsOutput represents the output of listing class skills
type ListClassSkillsOutput struct {
	Skills []*rpg.RPGClassSkill
}

// Execute lists skills for a class
func (uc *ListClassSkillsUseCase) Execute(ctx context.Context, input ListClassSkillsInput) (*ListClassSkillsOutput, error) {
	skills, err := uc.classSkillRepo.ListByClass(ctx, input.ClassID)
	if err != nil {
		uc.logger.Error("failed to list class skills", "error", err, "class_id", input.ClassID)
		return nil, err
	}

	return &ListClassSkillsOutput{
		Skills: skills,
	}, nil
}


