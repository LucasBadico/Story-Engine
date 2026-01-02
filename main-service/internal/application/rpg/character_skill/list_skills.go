package character_skill

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListCharacterSkillsUseCase handles listing character skills
type ListCharacterSkillsUseCase struct {
	characterSkillRepo repositories.CharacterSkillRepository
	logger             logger.Logger
}

// NewListCharacterSkillsUseCase creates a new ListCharacterSkillsUseCase
func NewListCharacterSkillsUseCase(
	characterSkillRepo repositories.CharacterSkillRepository,
	logger logger.Logger,
) *ListCharacterSkillsUseCase {
	return &ListCharacterSkillsUseCase{
		characterSkillRepo: characterSkillRepo,
		logger:             logger,
	}
}

// ListCharacterSkillsInput represents the input for listing character skills
type ListCharacterSkillsInput struct {
	TenantID    uuid.UUID
	CharacterID uuid.UUID
	ActiveOnly  bool
}

// ListCharacterSkillsOutput represents the output of listing character skills
type ListCharacterSkillsOutput struct {
	Skills []*rpg.CharacterSkill
}

// Execute lists skills for a character
func (uc *ListCharacterSkillsUseCase) Execute(ctx context.Context, input ListCharacterSkillsInput) (*ListCharacterSkillsOutput, error) {
	var skills []*rpg.CharacterSkill
	var err error

	if input.ActiveOnly {
		skills, err = uc.characterSkillRepo.ListActiveByCharacter(ctx, input.TenantID, input.CharacterID)
	} else {
		skills, err = uc.characterSkillRepo.ListByCharacter(ctx, input.TenantID, input.CharacterID)
	}

	if err != nil {
		uc.logger.Error("failed to list character skills", "error", err, "character_id", input.CharacterID)
		return nil, err
	}

	return &ListCharacterSkillsOutput{
		Skills: skills,
	}, nil
}


