package story

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/application/world"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/queue"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateStoryUseCase handles story creation
type CreateStoryUseCase struct {
	storyRepo      repositories.StoryRepository
	tenantRepo     repositories.TenantRepository
	worldRepo      repositories.WorldRepository
	createWorldUC  *world.CreateWorldUseCase
	auditLogRepo   repositories.AuditLogRepository
	ingestionQueue queue.IngestionQueue
	logger         logger.Logger
}

// NewCreateStoryUseCase creates a new CreateStoryUseCase
func NewCreateStoryUseCase(
	storyRepo repositories.StoryRepository,
	tenantRepo repositories.TenantRepository,
	worldRepo repositories.WorldRepository,
	createWorldUC *world.CreateWorldUseCase,
	auditLogRepo repositories.AuditLogRepository,
	ingestionQueue queue.IngestionQueue,
	logger logger.Logger,
) *CreateStoryUseCase {
	return &CreateStoryUseCase{
		storyRepo:     storyRepo,
		tenantRepo:    tenantRepo,
		worldRepo:     worldRepo,
		createWorldUC: createWorldUC,
		auditLogRepo:  auditLogRepo,
		ingestionQueue: ingestionQueue,
		logger:        logger,
	}
}

// CreateStoryInput represents the input for creating a story
type CreateStoryInput struct {
	TenantID       uuid.UUID
	Title          string
	WorldID        *uuid.UUID
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

	// Handle world_id: if not provided, create implicit world
	worldID := input.WorldID
	if worldID == nil {
		// Create implicit world
		worldName := fmt.Sprintf("World of %s", input.Title)
		createWorldOutput, err := uc.createWorldUC.Execute(ctx, world.CreateWorldInput{
			TenantID:    input.TenantID,
			Name:        worldName,
			IsImplicit:  true,
		})
		if err != nil {
			uc.logger.Error("failed to create implicit world", "error", err, "title", input.Title)
			return nil, err
		}
		worldID = &createWorldOutput.World.ID
		uc.logger.Info("created implicit world for story", "world_id", worldID, "story_title", input.Title)
	} else {
		// Validate world exists
		_, err := uc.worldRepo.GetByID(ctx, input.TenantID, *worldID)
		if err != nil {
			return nil, err
		}
	}

	// Create story (as version 1, root_story_id = self, previous_story_id = NULL)
	newStory, err := story.NewStory(input.TenantID, input.Title, input.CreatedByUserID)
	if err != nil {
		return nil, err
	}

	// Set world_id
	newStory.WorldID = worldID

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
	uc.enqueueIngestion(ctx, input.TenantID, newStory.ID)

	return &CreateStoryOutput{
		Story: newStory,
	}, nil
}

func (uc *CreateStoryUseCase) enqueueIngestion(ctx context.Context, tenantID uuid.UUID, storyID uuid.UUID) {
	if uc.ingestionQueue == nil {
		return
	}
	if err := uc.ingestionQueue.Push(ctx, tenantID, "story", storyID); err != nil {
		uc.logger.Error("failed to enqueue story ingestion", "error", err, "story_id", storyID, "tenant_id", tenantID)
	}
}
