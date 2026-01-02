package chapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateChapterUseCase handles chapter creation
type CreateChapterUseCase struct {
	chapterRepo repositories.ChapterRepository
	storyRepo    repositories.StoryRepository
	logger       logger.Logger
}

// NewCreateChapterUseCase creates a new CreateChapterUseCase
func NewCreateChapterUseCase(
	chapterRepo repositories.ChapterRepository,
	storyRepo repositories.StoryRepository,
	logger logger.Logger,
) *CreateChapterUseCase {
	return &CreateChapterUseCase{
		chapterRepo: chapterRepo,
		storyRepo:   storyRepo,
		logger:      logger,
	}
}

// CreateChapterInput represents the input for creating a chapter
type CreateChapterInput struct {
	TenantID uuid.UUID
	StoryID  uuid.UUID
	Number   int
	Title    string
	Status   *story.ChapterStatus
}

// CreateChapterOutput represents the output of creating a chapter
type CreateChapterOutput struct {
	Chapter *story.Chapter
}

// Execute creates a new chapter
func (uc *CreateChapterUseCase) Execute(ctx context.Context, input CreateChapterInput) (*CreateChapterOutput, error) {
	// Validate story exists
	_, err := uc.storyRepo.GetByID(ctx, input.TenantID, input.StoryID)
	if err != nil {
		return nil, err
	}

	if input.Number < 1 {
		return nil, &platformerrors.ValidationError{
			Field:   "number",
			Message: "must be greater than 0",
		}
	}

	if input.Title == "" {
		return nil, &platformerrors.ValidationError{
			Field:   "title",
			Message: "title is required",
		}
	}

	chapter, err := story.NewChapter(input.TenantID, input.StoryID, input.Number, input.Title)
	if err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "chapter",
			Message: err.Error(),
		}
	}

	// Set status if provided
	if input.Status != nil {
		if err := chapter.UpdateStatus(*input.Status); err != nil {
			return nil, &platformerrors.ValidationError{
				Field:   "status",
				Message: "invalid status",
			}
		}
	}

	if err := chapter.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "chapter",
			Message: err.Error(),
		}
	}

	if err := uc.chapterRepo.Create(ctx, chapter); err != nil {
		uc.logger.Error("failed to create chapter", "error", err, "story_id", input.StoryID, "tenant_id", input.TenantID)
		return nil, err
	}

	uc.logger.Info("chapter created", "chapter_id", chapter.ID, "story_id", input.StoryID, "tenant_id", input.TenantID)

	return &CreateChapterOutput{
		Chapter: chapter,
	}, nil
}

