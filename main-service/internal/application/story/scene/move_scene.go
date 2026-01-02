package scene

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// MoveSceneUseCase handles moving a scene to a different chapter
type MoveSceneUseCase struct {
	sceneRepo   repositories.SceneRepository
	chapterRepo repositories.ChapterRepository
	logger      logger.Logger
}

// NewMoveSceneUseCase creates a new MoveSceneUseCase
func NewMoveSceneUseCase(
	sceneRepo repositories.SceneRepository,
	chapterRepo repositories.ChapterRepository,
	logger logger.Logger,
) *MoveSceneUseCase {
	return &MoveSceneUseCase{
		sceneRepo:   sceneRepo,
		chapterRepo: chapterRepo,
		logger:      logger,
	}
}

// MoveSceneInput represents the input for moving a scene
type MoveSceneInput struct {
	TenantID    uuid.UUID
	SceneID     uuid.UUID
	NewChapterID *uuid.UUID // nil = move to root (no chapter)
}

// MoveSceneOutput represents the output of moving a scene
type MoveSceneOutput struct {
	Scene *story.Scene
}

// Execute moves a scene to a different chapter
func (uc *MoveSceneUseCase) Execute(ctx context.Context, input MoveSceneInput) (*MoveSceneOutput, error) {
	// Get existing scene
	scene, err := uc.sceneRepo.GetByID(ctx, input.TenantID, input.SceneID)
	if err != nil {
		return nil, err
	}

	// Validate new chapter exists if provided
	if input.NewChapterID != nil {
		_, err := uc.chapterRepo.GetByID(ctx, input.TenantID, *input.NewChapterID)
		if err != nil {
			return nil, err
		}
	}

	// Update chapter
	scene.UpdateChapter(input.NewChapterID)

	if err := scene.Validate(); err != nil {
		return nil, err
	}

	if err := uc.sceneRepo.Update(ctx, scene); err != nil {
		uc.logger.Error("failed to move scene", "error", err, "scene_id", input.SceneID, "tenant_id", input.TenantID)
		return nil, err
	}

	uc.logger.Info("scene moved", "scene_id", input.SceneID, "new_chapter_id", input.NewChapterID, "tenant_id", input.TenantID)

	return &MoveSceneOutput{
		Scene: scene,
	}, nil
}

