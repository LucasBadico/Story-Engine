package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetEventArtifactsUseCase handles retrieving artifacts for an event
type GetEventArtifactsUseCase struct {
	eventArtifactRepo repositories.EventArtifactRepository
	logger            logger.Logger
}

// NewGetEventArtifactsUseCase creates a new GetEventArtifactsUseCase
func NewGetEventArtifactsUseCase(
	eventArtifactRepo repositories.EventArtifactRepository,
	logger logger.Logger,
) *GetEventArtifactsUseCase {
	return &GetEventArtifactsUseCase{
		eventArtifactRepo: eventArtifactRepo,
		logger:            logger,
	}
}

// GetEventArtifactsInput represents the input for getting artifacts
type GetEventArtifactsInput struct {
	EventID uuid.UUID
}

// GetEventArtifactsOutput represents the output of getting artifacts
type GetEventArtifactsOutput struct {
	Artifacts []*world.EventArtifact
}

// Execute retrieves artifacts for an event
func (uc *GetEventArtifactsUseCase) Execute(ctx context.Context, input GetEventArtifactsInput) (*GetEventArtifactsOutput, error) {
	artifacts, err := uc.eventArtifactRepo.ListByEvent(ctx, input.EventID)
	if err != nil {
		uc.logger.Error("failed to get event artifacts", "error", err, "event_id", input.EventID)
		return nil, err
	}

	return &GetEventArtifactsOutput{
		Artifacts: artifacts,
	}, nil
}

