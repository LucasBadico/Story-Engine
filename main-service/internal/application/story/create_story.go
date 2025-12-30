package story

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateStoryUseCase handles story creation
type CreateStoryUseCase struct {
	storyRepo     repositories.StoryRepository
	tenantRepo    repositories.TenantRepository
	auditLogRepo  repositories.AuditLogRepository
	logger        logger.Logger
}

// NewCreateStoryUseCase creates a new CreateStoryUseCase
func NewCreateStoryUseCase(
	storyRepo repositories.StoryRepository,
	tenantRepo repositories.TenantRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *CreateStoryUseCase {
	return &CreateStoryUseCase{
		storyRepo:    storyRepo,
		tenantRepo:  tenantRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// CreateStoryInput represents the input for creating a story
type CreateStoryInput struct {
	TenantID      uuid.UUID
	Title         string
	CreatedByUserID *uuid.UUID
}

// CreateStoryOutput represents the output of creating a story
type CreateStoryOutput struct {
	Story *story.Story
}

// Execute creates a new story
func (uc *CreateStoryUseCase) Execute(ctx context.Context, input CreateStoryInput) (*CreateStoryOutput, error) {
	// Validate tenant exists
	_, err := uc.tenantRepo.GetByID(ctx, input.TenantID)
	if err != nil {
		// Return the error directly - repository already wraps it as NotFoundError
		return nil, err
	}

	// Validate title
	if input.Title == "" {
		return nil, &platformerrors.ValidationError{
			Field:   "title",
			Message: "story title is required",
		}
	}

	// Create story (as version 1, root_story_id = self, previous_story_id = NULL)
	newStory, err := story.NewStory(input.TenantID, input.Title, input.CreatedByUserID)
	if err != nil {
		return nil, err
	}

	if err := newStory.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "story",
			Message: err.Error(),
		}
	}

	if err := uc.storyRepo.Create(ctx, newStory); err != nil {
		uc.logger.Error("failed to create story", "error", err, "title", input.Title)
		return nil, err
	}

	// Log audit event
	auditLog := audit.NewAuditLog(
		input.TenantID,
		input.CreatedByUserID,
		audit.ActionCreate,
		audit.EntityTypeStory,
		newStory.ID,
		map[string]interface{}{
			"title":         newStory.Title,
			"version_number": newStory.VersionNumber,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
		// Don't fail the operation if audit logging fails
	}

	uc.logger.Info("story created", "story_id", newStory.ID, "title", newStory.Title, "tenant_id", input.TenantID)

	return &CreateStoryOutput{
		Story: newStory,
	}, nil
}

