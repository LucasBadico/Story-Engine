package artifact_stats

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateArtifactStatsUseCase handles creating artifact RPG stats
type CreateArtifactStatsUseCase struct {
	artifactStatsRepo repositories.ArtifactRPGStatsRepository
	artifactRepo      repositories.ArtifactRepository
	eventRepo         repositories.EventRepository
	logger            logger.Logger
}

// NewCreateArtifactStatsUseCase creates a new CreateArtifactStatsUseCase
func NewCreateArtifactStatsUseCase(
	artifactStatsRepo repositories.ArtifactRPGStatsRepository,
	artifactRepo repositories.ArtifactRepository,
	eventRepo repositories.EventRepository,
	logger logger.Logger,
) *CreateArtifactStatsUseCase {
	return &CreateArtifactStatsUseCase{
		artifactStatsRepo: artifactStatsRepo,
		artifactRepo:      artifactRepo,
		eventRepo:         eventRepo,
		logger:            logger,
	}
}

// CreateArtifactStatsInput represents the input for creating artifact stats
type CreateArtifactStatsInput struct {
	ArtifactID        uuid.UUID
	EventID           *uuid.UUID
	Stats             json.RawMessage
	Reason            *string
	Timeline          *string
	DeactivatePrevious bool // true = deactivate previous versions
}

// CreateArtifactStatsOutput represents the output of creating artifact stats
type CreateArtifactStatsOutput struct {
	Stats *rpg.ArtifactRPGStats
}

// Execute creates new artifact RPG stats
func (uc *CreateArtifactStatsUseCase) Execute(ctx context.Context, input CreateArtifactStatsInput) (*CreateArtifactStatsOutput, error) {
	// Validate artifact exists
	_, err := uc.artifactRepo.GetByID(ctx, input.ArtifactID)
	if err != nil {
		return nil, err
	}

	// Validate event exists if provided
	if input.EventID != nil {
		_, err := uc.eventRepo.GetByID(ctx, *input.EventID)
		if err != nil {
			return nil, err
		}
	}

	// Create stats
	stats, err := rpg.NewArtifactRPGStats(input.ArtifactID, input.Stats)
	if err != nil {
		return nil, err
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
	version, err := uc.artifactStatsRepo.GetNextVersion(ctx, input.ArtifactID)
	if err != nil {
		return nil, err
	}
	stats.SetVersion(version)

	// Deactivate previous versions if requested
	if input.DeactivatePrevious {
		if err := uc.artifactStatsRepo.DeactivateAllByArtifact(ctx, input.ArtifactID); err != nil {
			return nil, err
		}
		stats.SetActive(true)
	} else {
		// Check if there's an active version
		activeStats, err := uc.artifactStatsRepo.GetActiveByArtifact(ctx, input.ArtifactID)
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

	if err := uc.artifactStatsRepo.Create(ctx, stats); err != nil {
		uc.logger.Error("failed to create artifact RPG stats", "error", err, "artifact_id", input.ArtifactID)
		return nil, err
	}

	uc.logger.Info("artifact RPG stats created", "artifact_id", input.ArtifactID, "version", stats.Version)

	return &CreateArtifactStatsOutput{
		Stats: stats,
	}, nil
}

