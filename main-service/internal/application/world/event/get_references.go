package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetReferencesUseCase handles retrieving references for an event
type GetReferencesUseCase struct {
	eventReferenceRepo repositories.EventReferenceRepository
	logger             logger.Logger
}

// NewGetReferencesUseCase creates a new GetReferencesUseCase
func NewGetReferencesUseCase(
	eventReferenceRepo repositories.EventReferenceRepository,
	logger logger.Logger,
) *GetReferencesUseCase {
	return &GetReferencesUseCase{
		eventReferenceRepo: eventReferenceRepo,
		logger:            logger,
	}
}

// GetReferencesInput represents the input for getting references
type GetReferencesInput struct {
	TenantID uuid.UUID
	EventID  uuid.UUID
}

// GetReferencesOutput represents the output of getting references
type GetReferencesOutput struct {
	References []*world.EventReference
}

// Execute retrieves references for an event
func (uc *GetReferencesUseCase) Execute(ctx context.Context, input GetReferencesInput) (*GetReferencesOutput, error) {
	references, err := uc.eventReferenceRepo.ListByEvent(ctx, input.TenantID, input.EventID)
	if err != nil {
		uc.logger.Error("failed to get event references", "error", err, "event_id", input.EventID)
		return nil, err
	}

	return &GetReferencesOutput{
		References: references,
	}, nil
}

