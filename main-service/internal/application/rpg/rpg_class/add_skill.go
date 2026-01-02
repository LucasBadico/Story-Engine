package rpg_class

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddSkillToClassUseCase handles adding a skill to a class
type AddSkillToClassUseCase struct {
	classSkillRepo repositories.RPGClassSkillRepository
	classRepo      repositories.RPGClassRepository
	skillRepo      repositories.SkillRepository
	logger         logger.Logger
}

// NewAddSkillToClassUseCase creates a new AddSkillToClassUseCase
func NewAddSkillToClassUseCase(
	classSkillRepo repositories.RPGClassSkillRepository,
	classRepo repositories.RPGClassRepository,
	skillRepo repositories.SkillRepository,
	logger logger.Logger,
) *AddSkillToClassUseCase {
	return &AddSkillToClassUseCase{
		classSkillRepo: classSkillRepo,
		classRepo:      classRepo,
		skillRepo:      skillRepo,
		logger:         logger,
	}
}

// AddSkillToClassInput represents the input for adding a skill to a class
type AddSkillToClassInput struct {
	TenantID     uuid.UUID
	ClassID      uuid.UUID
	SkillID      uuid.UUID
	UnlockLevel  int
}

// AddSkillToClassOutput represents the output of adding a skill to a class
type AddSkillToClassOutput struct {
	ClassSkill *rpg.RPGClassSkill
}

// Execute adds a skill to a class
func (uc *AddSkillToClassUseCase) Execute(ctx context.Context, input AddSkillToClassInput) (*AddSkillToClassOutput, error) {
	// Validate class exists
	_, err := uc.classRepo.GetByID(ctx, input.TenantID, input.ClassID)
	if err != nil {
		return nil, err
	}

	// Validate skill exists
	_, err = uc.skillRepo.GetByID(ctx, input.TenantID, input.SkillID)
	if err != nil {
		return nil, err
	}

	// Check if skill already exists for this class
	existing, err := uc.classSkillRepo.GetByClassAndSkill(ctx, input.TenantID, input.ClassID, input.SkillID)
	if err == nil && existing != nil {
		// Update unlock level if different
		if existing.UnlockLevel != input.UnlockLevel {
			existing.SetUnlockLevel(input.UnlockLevel)
			if err := uc.classSkillRepo.Update(ctx, existing); err != nil {
				return nil, err
			}
		}
		return &AddSkillToClassOutput{
			ClassSkill: existing,
		}, nil
	}

	// Create class skill
	classSkill, err := rpg.NewRPGClassSkill(input.ClassID, input.SkillID, input.UnlockLevel)
	if err != nil {
		return nil, err
	}

	if err := classSkill.Validate(); err != nil {
		return nil, err
	}

	if err := uc.classSkillRepo.Create(ctx, classSkill); err != nil {
		uc.logger.Error("failed to add skill to class", "error", err, "class_id", input.ClassID, "skill_id", input.SkillID)
		return nil, err
	}

	uc.logger.Info("skill added to class", "class_id", input.ClassID, "skill_id", input.SkillID)

	return &AddSkillToClassOutput{
		ClassSkill: classSkill,
	}, nil
}


