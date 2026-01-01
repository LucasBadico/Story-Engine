package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// RemoveArtifactFromEventUseCase handles removing an artifact from an event
type RemoveArtifactFromEventUseCase struct {
	eventArtifactRepo repositories.EventArtifactRepository
	logger            logger.Logger
}

// NewRemoveArtifactFromEventUseCase creates a new RemoveArtifactFromEventUseCase
func NewRemoveArtifactFromEventUseCase(
	eventArtifactRepo repositories.EventArtifactRepository,
	logger logger.Logger,
) *RemoveArtifactFromEventUseCase {
	return &RemoveArtifactFromEventUseCase{
		eventArtifactRepo: eventArtifactRepo,
		logger:            logger,
	}
}

// RemoveArtifactFromEventInput represents the input for removing an artifact from an event
type RemoveArtifactFromEventInput struct {
	EventID    uuid.UUID
	ArtifactID uuid.UUID
}

// Execute removes an artifact from an event
func (uc *RemoveArtifactFromEventUseCase) Execute(ctx context.Context, input RemoveArtifactFromEventInput) error {
	if err := uc.eventArtifactRepo.DeleteByEventAndArtifact(ctx, input.EventID, input.ArtifactID); err != nil {
		uc.logger.Error("failed to remove artifact from event", "error", err, "event_id", input.EventID, "artifact_id", input.ArtifactID)
		return err
	}

	uc.logger.Info("artifact removed from event", "event_id", input.EventID, "artifact_id", input.ArtifactID)
	return nil
}

