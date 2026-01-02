package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddArtifactToEventUseCase handles adding an artifact to an event
type AddArtifactToEventUseCase struct {
	eventRepo         repositories.EventRepository
	artifactRepo      repositories.ArtifactRepository
	eventArtifactRepo repositories.EventArtifactRepository
	logger            logger.Logger
}

// NewAddArtifactToEventUseCase creates a new AddArtifactToEventUseCase
func NewAddArtifactToEventUseCase(
	eventRepo repositories.EventRepository,
	artifactRepo repositories.ArtifactRepository,
	eventArtifactRepo repositories.EventArtifactRepository,
	logger logger.Logger,
) *AddArtifactToEventUseCase {
	return &AddArtifactToEventUseCase{
		eventRepo:         eventRepo,
		artifactRepo:      artifactRepo,
		eventArtifactRepo: eventArtifactRepo,
		logger:            logger,
	}
}

// AddArtifactToEventInput represents the input for adding an artifact to an event
type AddArtifactToEventInput struct {
	EventID    uuid.UUID
	ArtifactID uuid.UUID
	Role       *string
}

// Execute adds an artifact to an event
func (uc *AddArtifactToEventUseCase) Execute(ctx context.Context, input AddArtifactToEventInput) error {
	// Validate event exists
	event, err := uc.eventRepo.GetByID(ctx, input.EventID)
	if err != nil {
		return err
	}

	// Validate artifact exists and belongs to same world
	artifact, err := uc.artifactRepo.GetByID(ctx, input.ArtifactID)
	if err != nil {
		return err
	}
	if artifact.WorldID != event.WorldID {
		return &platformerrors.ValidationError{
			Field:   "artifact_id",
			Message: "artifact must belong to the same world as the event",
		}
	}

	// Create relationship
	ea := world.NewEventArtifact(input.EventID, input.ArtifactID, input.Role)
	if err := uc.eventArtifactRepo.Create(ctx, ea); err != nil {
		uc.logger.Error("failed to add artifact to event", "error", err, "event_id", input.EventID, "artifact_id", input.ArtifactID)
		return err
	}

	uc.logger.Info("artifact added to event", "event_id", input.EventID, "artifact_id", input.ArtifactID)
	return nil
}


