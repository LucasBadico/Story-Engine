package rpg_class

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// RemoveSkillFromClassUseCase handles removing a skill from a class
type RemoveSkillFromClassUseCase struct {
	classSkillRepo repositories.RPGClassSkillRepository
	classRepo      repositories.RPGClassRepository
	logger         logger.Logger
}

// NewRemoveSkillFromClassUseCase creates a new RemoveSkillFromClassUseCase
func NewRemoveSkillFromClassUseCase(
	classSkillRepo repositories.RPGClassSkillRepository,
	classRepo repositories.RPGClassRepository,
	logger logger.Logger,
) *RemoveSkillFromClassUseCase {
	return &RemoveSkillFromClassUseCase{
		classSkillRepo: classSkillRepo,
		classRepo:      classRepo,
		logger:         logger,
	}
}

// RemoveSkillFromClassInput represents the input for removing a skill from a class
type RemoveSkillFromClassInput struct {
	TenantID uuid.UUID
	ClassID  uuid.UUID
	SkillID  uuid.UUID
}

// RemoveSkillFromClassOutput represents the output of removing a skill from a class
type RemoveSkillFromClassOutput struct{}

// Execute removes a skill from a class
func (uc *RemoveSkillFromClassUseCase) Execute(ctx context.Context, input RemoveSkillFromClassInput) (*RemoveSkillFromClassOutput, error) {
	// Validate class exists
	_, err := uc.classRepo.GetByID(ctx, input.TenantID, input.ClassID)
	if err != nil {
		return nil, err
	}

	// Get class skill
	classSkill, err := uc.classSkillRepo.GetByClassAndSkill(ctx, input.TenantID, input.ClassID, input.SkillID)
	if err != nil {
		return nil, err
	}

	// Delete class skill
	if err := uc.classSkillRepo.Delete(ctx, input.TenantID, classSkill.ID); err != nil {
		uc.logger.Error("failed to remove skill from class", "error", err, "class_id", input.ClassID, "skill_id", input.SkillID)
		return nil, err
	}

	uc.logger.Info("skill removed from class", "class_id", input.ClassID, "skill_id", input.SkillID)

	return &RemoveSkillFromClassOutput{}, nil
}

