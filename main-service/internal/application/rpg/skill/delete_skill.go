package skill

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteSkillUseCase handles skill deletion
type DeleteSkillUseCase struct {
	skillRepo repositories.SkillRepository
	logger    logger.Logger
}

// NewDeleteSkillUseCase creates a new DeleteSkillUseCase
func NewDeleteSkillUseCase(
	skillRepo repositories.SkillRepository,
	logger logger.Logger,
) *DeleteSkillUseCase {
	return &DeleteSkillUseCase{
		skillRepo: skillRepo,
		logger:    logger,
	}
}

// DeleteSkillInput represents the input for deleting a skill
type DeleteSkillInput struct {
	ID uuid.UUID
}

// Execute deletes a skill
func (uc *DeleteSkillUseCase) Execute(ctx context.Context, input DeleteSkillInput) error {
	if err := uc.skillRepo.Delete(ctx, input.ID); err != nil {
		uc.logger.Error("failed to delete skill", "error", err, "skill_id", input.ID)
		return err
	}

	uc.logger.Info("skill deleted", "skill_id", input.ID)

	return nil
}

