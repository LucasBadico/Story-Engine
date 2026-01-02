package character_stats

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateCharacterStatsUseCase handles creating character RPG stats
type CreateCharacterStatsUseCase struct {
	characterStatsRepo repositories.CharacterRPGStatsRepository
	characterRepo      repositories.CharacterRepository
	eventRepo          repositories.EventRepository
	logger             logger.Logger
}

// NewCreateCharacterStatsUseCase creates a new CreateCharacterStatsUseCase
func NewCreateCharacterStatsUseCase(
	characterStatsRepo repositories.CharacterRPGStatsRepository,
	characterRepo repositories.CharacterRepository,
	eventRepo repositories.EventRepository,
	logger logger.Logger,
) *CreateCharacterStatsUseCase {
	return &CreateCharacterStatsUseCase{
		characterStatsRepo: characterStatsRepo,
		characterRepo:      characterRepo,
		eventRepo:          eventRepo,
		logger:             logger,
	}
}

// CreateCharacterStatsInput represents the input for creating character stats
type CreateCharacterStatsInput struct {
	TenantID          uuid.UUID
	CharacterID       uuid.UUID
	EventID           *uuid.UUID
	BaseStats         json.RawMessage
	DerivedStats      *json.RawMessage
	Progression       *json.RawMessage
	Reason            *string
	Timeline          *string
	DeactivatePrevious bool // true = deactivate previous versions
}

// CreateCharacterStatsOutput represents the output of creating character stats
type CreateCharacterStatsOutput struct {
	Stats *rpg.CharacterRPGStats
}

// Execute creates new character RPG stats
func (uc *CreateCharacterStatsUseCase) Execute(ctx context.Context, input CreateCharacterStatsInput) (*CreateCharacterStatsOutput, error) {
	// Validate character exists
	_, err := uc.characterRepo.GetByID(ctx, input.TenantID, input.CharacterID)
	if err != nil {
		return nil, err
	}

	// Validate event exists if provided
	if input.EventID != nil {
		_, err := uc.eventRepo.GetByID(ctx, input.TenantID, *input.EventID)
		if err != nil {
			return nil, err
		}
	}

	// Create stats
	stats, err := rpg.NewCharacterRPGStats(input.TenantID, input.CharacterID, input.BaseStats)
	if err != nil {
		return nil, err
	}

	if input.DerivedStats != nil {
		stats.DerivedStats = input.DerivedStats
	}
	if input.Progression != nil {
		stats.Progression = input.Progression
	}
	if input.EventID != nil {
		stats.SetEventID(input.EventID)
	}
	if input.Reason != nil {
		stats.SetReason(input.Reason)
	}
	if input.Timeline != nil {
		stats.SetTimeline(input.Timeline)
	}

	// Get next version number
	version, err := uc.characterStatsRepo.GetNextVersion(ctx, input.TenantID, input.CharacterID)
	if err != nil {
		return nil, err
	}
	stats.SetVersion(version)

	// Deactivate previous versions if requested
	if input.DeactivatePrevious {
		if err := uc.characterStatsRepo.DeactivateAllByCharacter(ctx, input.TenantID, input.CharacterID); err != nil {
			return nil, err
		}
		stats.SetActive(true)
	} else {
		// Check if there's an active version
		activeStats, err := uc.characterStatsRepo.GetActiveByCharacter(ctx, input.TenantID, input.CharacterID)
		if err == nil && activeStats != nil {
			// There's an active version, this one should be inactive
			stats.SetActive(false)
		} else {
			// No active version, make this one active
			stats.SetActive(true)
		}
	}

	if err := stats.Validate(); err != nil {
		return nil, err
	}

	if err := uc.characterStatsRepo.Create(ctx, stats); err != nil {
		uc.logger.Error("failed to create character RPG stats", "error", err, "character_id", input.CharacterID)
		return nil, err
	}

	uc.logger.Info("character RPG stats created", "character_id", input.CharacterID, "version", stats.Version)

	return &CreateCharacterStatsOutput{
		Stats: stats,
	}, nil
}


