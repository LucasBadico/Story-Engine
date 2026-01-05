package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// RemoveReferenceUseCase handles removing a reference from an event
type RemoveReferenceUseCase struct {
	eventReferenceRepo repositories.EventReferenceRepository
	logger             logger.Logger
}

// NewRemoveReferenceUseCase creates a new RemoveReferenceUseCase
func NewRemoveReferenceUseCase(
	eventReferenceRepo repositories.EventReferenceRepository,
	logger logger.Logger,
) *RemoveReferenceUseCase {
	return &RemoveReferenceUseCase{
		eventReferenceRepo: eventReferenceRepo,
		logger:            logger,
	}
}

// RemoveReferenceInput represents the input for removing a reference
type RemoveReferenceInput struct {
	TenantID   uuid.UUID
	EventID    uuid.UUID
	EntityType string
	EntityID   uuid.UUID
}

// Execute removes a reference from an event
func (uc *RemoveReferenceUseCase) Execute(ctx context.Context, input RemoveReferenceInput) error {
	err := uc.eventReferenceRepo.DeleteByEventAndEntity(ctx, input.TenantID, input.EventID, input.EntityType, input.EntityID)
	if err != nil {
		uc.logger.Error("failed to remove event reference", "error", err, "event_id", input.EventID)
		return err
	}

	uc.logger.Info("event reference removed", "event_id", input.EventID, "entity_type", input.EntityType, "entity_id", input.EntityID)
	return nil
}

