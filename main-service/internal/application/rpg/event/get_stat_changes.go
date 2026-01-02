package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetEventStatChangesUseCase handles retrieving all stat changes caused by an event
type GetEventStatChangesUseCase struct {
	characterStatsRepo repositories.CharacterRPGStatsRepository
	artifactStatsRepo  repositories.ArtifactRPGStatsRepository
	logger             logger.Logger
}

// NewGetEventStatChangesUseCase creates a new GetEventStatChangesUseCase
func NewGetEventStatChangesUseCase(
	characterStatsRepo repositories.CharacterRPGStatsRepository,
	artifactStatsRepo repositories.ArtifactRPGStatsRepository,
	logger logger.Logger,
) *GetEventStatChangesUseCase {
	return &GetEventStatChangesUseCase{
		characterStatsRepo: characterStatsRepo,
		artifactStatsRepo:  artifactStatsRepo,
		logger:             logger,
	}
}

// GetEventStatChangesInput represents the input for getting stat changes
type GetEventStatChangesInput struct {
	EventID uuid.UUID
}

// GetEventStatChangesOutput represents the output of getting stat changes
type GetEventStatChangesOutput struct {
	CharacterStats []*rpg.CharacterRPGStats
	ArtifactStats  []*rpg.ArtifactRPGStats
}

// Execute retrieves all stat changes caused by an event
func (uc *GetEventStatChangesUseCase) Execute(ctx context.Context, input GetEventStatChangesInput) (*GetEventStatChangesOutput, error) {
	characterStats, err := uc.characterStatsRepo.ListByEvent(ctx, input.EventID)
	if err != nil {
		uc.logger.Error("failed to get character stats for event", "error", err, "event_id", input.EventID)
		return nil, err
	}

	artifactStats, err := uc.artifactStatsRepo.ListByEvent(ctx, input.EventID)
	if err != nil {
		uc.logger.Error("failed to get artifact stats for event", "error", err, "event_id", input.EventID)
		return nil, err
	}

	return &GetEventStatChangesOutput{
		CharacterStats: characterStats,
		ArtifactStats:  artifactStats,
	}, nil
}


