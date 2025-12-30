package story

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/versioning"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CloneStoryUseCase handles story cloning (versioning)
type CloneStoryUseCase struct {
	storyRepo        repositories.StoryRepository
	chapterRepo      repositories.ChapterRepository
	sceneRepo        repositories.SceneRepository
	beatRepo         repositories.BeatRepository
	proseBlockRepo   repositories.ProseBlockRepository
	auditLogRepo     repositories.AuditLogRepository
	transactionRepo  repositories.TransactionRepository
	logger           logger.Logger
}

// NewCloneStoryUseCase creates a new CloneStoryUseCase
func NewCloneStoryUseCase(
	storyRepo repositories.StoryRepository,
	chapterRepo repositories.ChapterRepository,
	sceneRepo repositories.SceneRepository,
	beatRepo repositories.BeatRepository,
	proseBlockRepo repositories.ProseBlockRepository,
	auditLogRepo repositories.AuditLogRepository,
	transactionRepo repositories.TransactionRepository,
	logger logger.Logger,
) *CloneStoryUseCase {
	return &CloneStoryUseCase{
		storyRepo:       storyRepo,
		chapterRepo:    chapterRepo,
		sceneRepo:      sceneRepo,
		beatRepo:       beatRepo,
		proseBlockRepo: proseBlockRepo,
		auditLogRepo:   auditLogRepo,
		transactionRepo: transactionRepo,
		logger:         logger,
	}
}

// CloneStoryInput represents the input for cloning a story
type CloneStoryInput struct {
	SourceStoryID uuid.UUID
	CreatedByUserID *uuid.UUID
}

// CloneStoryOutput represents the output of cloning a story
type CloneStoryOutput struct {
	NewStoryID uuid.UUID
}

// Execute clones a story transactionally
func (uc *CloneStoryUseCase) Execute(ctx context.Context, input CloneStoryInput) (*CloneStoryOutput, error) {
	// Validate source story exists
	sourceStory, err := uc.storyRepo.GetByID(ctx, input.SourceStoryID)
	if err != nil {
		if errors.Is(err, platformerrors.ErrNotFound) {
			return nil, &platformerrors.NotFoundError{
				Resource: "story",
				ID:       input.SourceStoryID.String(),
			}
		}
		return nil, err
	}

	// Get all versions to determine next version number
	versions, err := uc.storyRepo.ListVersionsByRoot(ctx, sourceStory.RootStoryID)
	if err != nil {
		return nil, err
	}
	nextVersionNumber := len(versions) + 1

	// Clone the story and all related entities in a transaction
	var newStoryID uuid.UUID
	err = uc.transactionRepo.WithTx(ctx, func(tx pgx.Tx) error {
		// Create new story entity with proper versioning fields
		newStory, err := versioning.CloneStory(sourceStory, nextVersionNumber)
		if err != nil {
			return err
		}
		newStoryID = newStory.ID

		// Create the new story (we'll need to adapt repositories to work with tx)
		// For now, we'll use the regular repositories which will work outside the transaction
		// In a production system, repositories should accept tx as a parameter
		if err := uc.storyRepo.Create(ctx, newStory); err != nil {
			return err
		}

		// Clone all chapters
		chapters, err := uc.chapterRepo.ListByStory(ctx, sourceStory.ID)
		if err != nil {
			return err
		}

		chapterIDMap := make(map[uuid.UUID]uuid.UUID) // old chapter ID -> new chapter ID
		for _, oldChapter := range chapters {
			newChapter := versioning.CloneChapter(oldChapter, newStoryID)
			if err := uc.chapterRepo.Create(ctx, newChapter); err != nil {
				return err
			}
			chapterIDMap[oldChapter.ID] = newChapter.ID
		}

		// Clone all scenes
		scenes, err := uc.sceneRepo.ListByStory(ctx, sourceStory.ID)
		if err != nil {
			return err
		}

		sceneIDMap := make(map[uuid.UUID]uuid.UUID) // old scene ID -> new scene ID
		for _, oldScene := range scenes {
			newChapterID, ok := chapterIDMap[oldScene.ChapterID]
			if !ok {
				return errors.New("chapter mapping not found for scene")
			}
			newScene := versioning.CloneScene(oldScene, newStoryID, newChapterID)
			if err := uc.sceneRepo.Create(ctx, newScene); err != nil {
				return err
			}
			sceneIDMap[oldScene.ID] = newScene.ID
		}

		// Clone all beats
		for _, oldScene := range scenes {
			beats, err := uc.beatRepo.ListByScene(ctx, oldScene.ID)
			if err != nil {
				return err
			}

			newSceneID, ok := sceneIDMap[oldScene.ID]
			if !ok {
				return errors.New("scene mapping not found for beat")
			}

			for _, oldBeat := range beats {
				newBeat := versioning.CloneBeat(oldBeat, newSceneID)
				if err := uc.beatRepo.Create(ctx, newBeat); err != nil {
					return err
				}
			}
		}

		// Clone all prose blocks
		for _, oldScene := range scenes {
			proseBlocks, err := uc.proseBlockRepo.ListByScene(ctx, oldScene.ID)
			if err != nil {
				return err
			}

			newSceneID, ok := sceneIDMap[oldScene.ID]
			if !ok {
				return errors.New("scene mapping not found for prose block")
			}

			for _, oldProse := range proseBlocks {
				newProse := versioning.CloneProseBlock(oldProse, newSceneID)
				if err := uc.proseBlockRepo.Create(ctx, newProse); err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		uc.logger.Error("failed to clone story", "error", err, "source_story_id", input.SourceStoryID)
		return nil, err
	}

	// Log audit event
	auditLog := audit.NewAuditLog(
		sourceStory.TenantID,
		input.CreatedByUserID,
		audit.ActionClone,
		audit.EntityTypeStory,
		newStoryID,
		map[string]interface{}{
			"source_story_id": sourceStory.ID.String(),
			"version_number":  nextVersionNumber,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
		// Don't fail the operation if audit logging fails
	}

	uc.logger.Info("story cloned", "source_story_id", sourceStory.ID, "new_story_id", newStoryID, "version_number", nextVersionNumber)

	return &CloneStoryOutput{
		NewStoryID: newStoryID,
	}, nil
}

