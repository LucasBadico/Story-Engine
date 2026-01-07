package story

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/queue"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateStoryUseCase handles story updates
type UpdateStoryUseCase struct {
	storyRepo repositories.StoryRepository
	ingestionQueue queue.IngestionQueue
	logger    logger.Logger
}

// NewUpdateStoryUseCase creates a new UpdateStoryUseCase
func NewUpdateStoryUseCase(
	storyRepo repositories.StoryRepository,
	ingestionQueue queue.IngestionQueue,
	logger logger.Logger,
) *UpdateStoryUseCase {
	return &UpdateStoryUseCase{
		storyRepo: storyRepo,
		ingestionQueue: ingestionQueue,
		logger:    logger,
	}
}

// UpdateStoryInput represents the input for updating a story
type UpdateStoryInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
	Title    *string
	Status   *story.StoryStatus
}

// UpdateStoryOutput represents the output of updating a story
type UpdateStoryOutput struct {
	Story *story.Story
}

// Execute updates a story
func (uc *UpdateStoryUseCase) Execute(ctx context.Context, input UpdateStoryInput) (*UpdateStoryOutput, error) {
	// Get existing story
	s, err := uc.storyRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Title != nil {
		if err := s.UpdateTitle(*input.Title); err != nil {
			return nil, &platformerrors.ValidationError{
				Field:   "title",
				Message: err.Error(),
			}
		}
	}

	if input.Status != nil {
		if err := s.UpdateStatus(*input.Status); err != nil {
			return nil, &platformerrors.ValidationError{
				Field:   "status",
				Message: "invalid status",
			}
		}
	}

	if err := s.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "story",
			Message: err.Error(),
		}
	}

	if err := uc.storyRepo.Update(ctx, s); err != nil {
		uc.logger.Error("failed to update story", "error", err, "story_id", input.ID, "tenant_id", input.TenantID)
		return nil, err
	}

	uc.logger.Info("story updated", "story_id", input.ID, "tenant_id", input.TenantID)
	uc.enqueueIngestion(ctx, input.TenantID, s.ID)

	return &UpdateStoryOutput{
		Story: s,
	}, nil
}

func (uc *UpdateStoryUseCase) enqueueIngestion(ctx context.Context, tenantID uuid.UUID, storyID uuid.UUID) {
	if uc.ingestionQueue == nil {
		return
	}
	if err := uc.ingestionQueue.Push(ctx, tenantID, "story", storyID); err != nil {
		uc.logger.Error("failed to enqueue story ingestion", "error", err, "story_id", storyID, "tenant_id", tenantID)
	}
}
