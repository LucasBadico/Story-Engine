package character_skill

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateCharacterSkillUseCase handles updating character skill (evolve rank, add XP, toggle active)
type UpdateCharacterSkillUseCase struct {
	characterSkillRepo repositories.CharacterSkillRepository
	skillRepo          repositories.SkillRepository
	logger             logger.Logger
}

// NewUpdateCharacterSkillUseCase creates a new UpdateCharacterSkillUseCase
func NewUpdateCharacterSkillUseCase(
	characterSkillRepo repositories.CharacterSkillRepository,
	skillRepo repositories.SkillRepository,
	logger logger.Logger,
) *UpdateCharacterSkillUseCase {
	return &UpdateCharacterSkillUseCase{
		characterSkillRepo: characterSkillRepo,
		skillRepo:          skillRepo,
		logger:             logger,
	}
}

// UpdateCharacterSkillInput represents the input for updating a character skill
type UpdateCharacterSkillInput struct {
	TenantID  uuid.UUID
	ID        uuid.UUID
	Rank      *int
	AddXP     *int
	IsActive  *bool
}

// UpdateCharacterSkillOutput represents the output of updating a character skill
type UpdateCharacterSkillOutput struct {
	CharacterSkill *rpg.CharacterSkill
}

// Execute updates a character skill
func (uc *UpdateCharacterSkillUseCase) Execute(ctx context.Context, input UpdateCharacterSkillInput) (*UpdateCharacterSkillOutput, error) {
	// Get existing character skill
	characterSkill, err := uc.characterSkillRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return nil, err
	}

	// Get skill to check max rank
	skill, err := uc.skillRepo.GetByID(ctx, input.TenantID, characterSkill.SkillID)
	if err != nil {
		return nil, err
	}

	// Update rank if provided
	if input.Rank != nil {
		if *input.Rank > skill.MaxRank {
			return nil, &platformerrors.ValidationError{
				Field:   "rank",
				Message: "rank cannot exceed skill max_rank",
			}
		}
		if err := characterSkill.SetRank(*input.Rank); err != nil {
			return nil, err
		}
	}

	// Add XP if provided
	if input.AddXP != nil {
		characterSkill.AddXP(*input.AddXP)
	}

	// Update active status if provided
	if input.IsActive != nil {
		characterSkill.SetActive(*input.IsActive)
	}

	if err := characterSkill.Validate(); err != nil {
		return nil, err
	}

	if err := uc.characterSkillRepo.Update(ctx, characterSkill); err != nil {
		uc.logger.Error("failed to update character skill", "error", err, "id", input.ID)
		return nil, err
	}

	uc.logger.Info("character skill updated", "id", input.ID, "character_id", characterSkill.CharacterID)

	return &UpdateCharacterSkillOutput{
		CharacterSkill: characterSkill,
	}, nil
}

