package character_skill

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// LearnSkillUseCase handles character learning a skill
type LearnSkillUseCase struct {
	characterSkillRepo repositories.CharacterSkillRepository
	characterRepo      repositories.CharacterRepository
	skillRepo          repositories.SkillRepository
	logger             logger.Logger
}

// NewLearnSkillUseCase creates a new LearnSkillUseCase
func NewLearnSkillUseCase(
	characterSkillRepo repositories.CharacterSkillRepository,
	characterRepo repositories.CharacterRepository,
	skillRepo repositories.SkillRepository,
	logger logger.Logger,
) *LearnSkillUseCase {
	return &LearnSkillUseCase{
		characterSkillRepo: characterSkillRepo,
		characterRepo:      characterRepo,
		skillRepo:          skillRepo,
		logger:             logger,
	}
}

// LearnSkillInput represents the input for learning a skill
type LearnSkillInput struct {
	TenantID    uuid.UUID
	CharacterID uuid.UUID
	SkillID     uuid.UUID
}

// LearnSkillOutput represents the output of learning a skill
type LearnSkillOutput struct {
	CharacterSkill *rpg.CharacterSkill
}

// Execute makes a character learn a skill
func (uc *LearnSkillUseCase) Execute(ctx context.Context, input LearnSkillInput) (*LearnSkillOutput, error) {
	// Validate character exists
	_, err := uc.characterRepo.GetByID(ctx, input.TenantID, input.CharacterID)
	if err != nil {
		return nil, err
	}

	// Validate skill exists
	_, err = uc.skillRepo.GetByID(ctx, input.TenantID, input.SkillID)
	if err != nil {
		return nil, err
	}

	// Check if character already knows this skill
	existing, err := uc.characterSkillRepo.GetByCharacterAndSkill(ctx, input.TenantID, input.CharacterID, input.SkillID)
	if err == nil && existing != nil {
		// Character already knows this skill
		return &LearnSkillOutput{
			CharacterSkill: existing,
		}, nil
	}

	// Create character skill
	characterSkill, err := rpg.NewCharacterSkill(input.CharacterID, input.SkillID)
	if err != nil {
		return nil, err
	}

	if err := characterSkill.Validate(); err != nil {
		return nil, err
	}

	if err := uc.characterSkillRepo.Create(ctx, characterSkill); err != nil {
		uc.logger.Error("failed to learn skill", "error", err, "character_id", input.CharacterID, "skill_id", input.SkillID)
		return nil, err
	}

	uc.logger.Info("character learned skill", "character_id", input.CharacterID, "skill_id", input.SkillID)

	return &LearnSkillOutput{
		CharacterSkill: characterSkill,
	}, nil
}
