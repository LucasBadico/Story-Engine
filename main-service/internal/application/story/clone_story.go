package story

import (
	"context"
	"errors"

	"github.com/google/uuid"
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
	contentBlockRepo repositories.ContentBlockRepository
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
	contentBlockRepo repositories.ContentBlockRepository,
	auditLogRepo repositories.AuditLogRepository,
	transactionRepo repositories.TransactionRepository,
	logger logger.Logger,
) *CloneStoryUseCase {
	return &CloneStoryUseCase{
		storyRepo:       storyRepo,
		chapterRepo:    chapterRepo,
		sceneRepo:      sceneRepo,
		beatRepo:       beatRepo,
		contentBlockRepo: contentBlockRepo,
		auditLogRepo:   auditLogRepo,
		transactionRepo: transactionRepo,
		logger:         logger,
	}
}

// CloneStoryInput represents the input for cloning a story
type CloneStoryInput struct {
	TenantID        uuid.UUID
	SourceStoryID  uuid.UUID
	CreatedByUserID *uuid.UUID
}

// CloneStoryOutput represents the output of cloning a story
type CloneStoryOutput struct {
	NewStoryID uuid.UUID
}

// Execute clones a story transactionally
func (uc *CloneStoryUseCase) Execute(ctx context.Context, input CloneStoryInput) (*CloneStoryOutput, error) {
	// Validate source story exists
	sourceStory, err := uc.storyRepo.GetByID(ctx, input.TenantID, input.SourceStoryID)
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
	versions, err := uc.storyRepo.ListVersionsByRoot(ctx, input.TenantID, sourceStory.RootStoryID)
	if err != nil {
		return nil, err
	}
	nextVersionNumber := len(versions) + 1

	// Clone the story and all related entities
	// Note: Currently repositories don't support transactions, so operations are not atomic
	// TODO: Add transaction support to repositories for proper atomicity
	
	// Create new story entity with proper versioning fields
	newStory, err := versioning.CloneStory(sourceStory, nextVersionNumber)
	if err != nil {
		return nil, err
	}
	newStoryID := newStory.ID

	// Create the new story
	if err := uc.storyRepo.Create(ctx, newStory); err != nil {
		return nil, err
	}

	// Clone all chapters
	chapters, err := uc.chapterRepo.ListByStory(ctx, input.TenantID, sourceStory.ID)
	if err != nil {
		return nil, err
	}

	chapterIDMap := make(map[uuid.UUID]uuid.UUID) // old chapter ID -> new chapter ID
	for _, oldChapter := range chapters {
		newChapter := versioning.CloneChapter(oldChapter, newStoryID)
		if err := uc.chapterRepo.Create(ctx, newChapter); err != nil {
			return nil, err
		}
		chapterIDMap[oldChapter.ID] = newChapter.ID
	}

	// Clone all scenes
	scenes, err := uc.sceneRepo.ListByStory(ctx, input.TenantID, sourceStory.ID)
	if err != nil {
		return nil, err
	}

	sceneIDMap := make(map[uuid.UUID]uuid.UUID) // old scene ID -> new scene ID
	for _, oldScene := range scenes {
		var newChapterID *uuid.UUID
		if oldScene.ChapterID != nil {
			mappedID, ok := chapterIDMap[*oldScene.ChapterID]
			if !ok {
				return nil, errors.New("chapter mapping not found for scene")
			}
			newChapterID = &mappedID
		}
		newScene := versioning.CloneScene(oldScene, newStoryID, newChapterID)
		if err := uc.sceneRepo.Create(ctx, newScene); err != nil {
			return nil, err
		}
		sceneIDMap[oldScene.ID] = newScene.ID
	}

	// Clone all beats
	for _, oldScene := range scenes {
		beats, err := uc.beatRepo.ListByScene(ctx, input.TenantID, oldScene.ID)
		if err != nil {
			return nil, err
		}

		newSceneID, ok := sceneIDMap[oldScene.ID]
		if !ok {
			return nil, errors.New("scene mapping not found for beat")
		}

		for _, oldBeat := range beats {
			newBeat := versioning.CloneBeat(oldBeat, newSceneID)
			if err := uc.beatRepo.Create(ctx, newBeat); err != nil {
				return nil, err
			}
		}
	}

	// Clone all content blocks
	for _, oldChapter := range chapters {
		contentBlocks, err := uc.contentBlockRepo.ListByChapter(ctx, input.TenantID, oldChapter.ID)
		if err != nil {
			return nil, err
		}

		newChapterID, ok := chapterIDMap[oldChapter.ID]
		if !ok {
			return nil, errors.New("chapter mapping not found for content block")
		}

		for _, oldContent := range contentBlocks {
			var newOrderNum *int
			if oldContent.OrderNum != nil {
				order := *oldContent.OrderNum
				newOrderNum = &order
			}
			newContent := versioning.CloneContentBlock(oldContent, &newChapterID, newOrderNum)
			if err := uc.contentBlockRepo.Create(ctx, newContent); err != nil {
				return nil, err
			}
		}
	}

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

