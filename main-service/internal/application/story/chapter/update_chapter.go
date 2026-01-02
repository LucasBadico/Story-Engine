package chapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateChapterUseCase handles chapter updates
type UpdateChapterUseCase struct {
	chapterRepo repositories.ChapterRepository
	logger      logger.Logger
}

// NewUpdateChapterUseCase creates a new UpdateChapterUseCase
func NewUpdateChapterUseCase(
	chapterRepo repositories.ChapterRepository,
	logger logger.Logger,
) *UpdateChapterUseCase {
	return &UpdateChapterUseCase{
		chapterRepo: chapterRepo,
		logger:      logger,
	}
}

// UpdateChapterInput represents the input for updating a chapter
type UpdateChapterInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
	Number   *int
	Title    *string
	Status   *story.ChapterStatus
}

// UpdateChapterOutput represents the output of updating a chapter
type UpdateChapterOutput struct {
	Chapter *story.Chapter
}

// Execute updates a chapter
func (uc *UpdateChapterUseCase) Execute(ctx context.Context, input UpdateChapterInput) (*UpdateChapterOutput, error) {
	// Get existing chapter
	chapter, err := uc.chapterRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Number != nil {
		if *input.Number < 1 {
			return nil, &platformerrors.ValidationError{
				Field:   "number",
				Message: "must be greater than 0",
			}
		}
		chapter.Number = *input.Number
	}

	if input.Title != nil {
		chapter.UpdateTitle(*input.Title)
	}

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

	if err := uc.chapterRepo.Update(ctx, chapter); err != nil {
		uc.logger.Error("failed to update chapter", "error", err, "chapter_id", input.ID, "tenant_id", input.TenantID)
		return nil, err
	}

	uc.logger.Info("chapter updated", "chapter_id", input.ID, "tenant_id", input.TenantID)

	return &UpdateChapterOutput{
		Chapter: chapter,
	}, nil
}

