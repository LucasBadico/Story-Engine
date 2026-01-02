package scene

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateSceneUseCase handles scene creation
type CreateSceneUseCase struct {
	sceneRepo   repositories.SceneRepository
	chapterRepo repositories.ChapterRepository
	storyRepo   repositories.StoryRepository
	logger      logger.Logger
}

// NewCreateSceneUseCase creates a new CreateSceneUseCase
func NewCreateSceneUseCase(
	sceneRepo repositories.SceneRepository,
	chapterRepo repositories.ChapterRepository,
	storyRepo repositories.StoryRepository,
	logger logger.Logger,
) *CreateSceneUseCase {
	return &CreateSceneUseCase{
		sceneRepo:   sceneRepo,
		chapterRepo: chapterRepo,
		storyRepo:   storyRepo,
		logger:      logger,
	}
}

// CreateSceneInput represents the input for creating a scene
type CreateSceneInput struct {
	TenantID       uuid.UUID
	StoryID        uuid.UUID
	ChapterID      *uuid.UUID
	OrderNum       int
	POVCharacterID *uuid.UUID
	TimeRef        string
	Goal           string
}

// CreateSceneOutput represents the output of creating a scene
type CreateSceneOutput struct {
	Scene *story.Scene
}

// Execute creates a new scene
func (uc *CreateSceneUseCase) Execute(ctx context.Context, input CreateSceneInput) (*CreateSceneOutput, error) {
	// Validate story exists
	_, err := uc.storyRepo.GetByID(ctx, input.TenantID, input.StoryID)
	if err != nil {
		return nil, err
	}

	// Validate chapter exists if provided
	if input.ChapterID != nil {
		_, err := uc.chapterRepo.GetByID(ctx, input.TenantID, *input.ChapterID)
		if err != nil {
			return nil, err
		}
	}

	if input.OrderNum < 1 {
		return nil, &platformerrors.ValidationError{
			Field:   "order_num",
			Message: "must be greater than 0",
		}
	}

	scene, err := story.NewScene(input.TenantID, input.StoryID, input.ChapterID, input.OrderNum)
	if err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "scene",
			Message: err.Error(),
		}
	}

	if input.Goal != "" {
		scene.UpdateGoal(input.Goal)
	}
	if input.TimeRef != "" {
		scene.TimeRef = input.TimeRef
	}
	if input.POVCharacterID != nil {
		scene.UpdatePOV(input.POVCharacterID)
	}

	if err := scene.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "scene",
			Message: err.Error(),
		}
	}

	if err := uc.sceneRepo.Create(ctx, scene); err != nil {
		uc.logger.Error("failed to create scene", "error", err, "story_id", input.StoryID, "tenant_id", input.TenantID)
		return nil, err
	}

	uc.logger.Info("scene created", "scene_id", scene.ID, "story_id", input.StoryID, "tenant_id", input.TenantID)

	return &CreateSceneOutput{
		Scene: scene,
	}, nil
}

