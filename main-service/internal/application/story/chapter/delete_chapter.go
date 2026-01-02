package chapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteChapterUseCase handles chapter deletion
type DeleteChapterUseCase struct {
	chapterRepo repositories.ChapterRepository
	logger      logger.Logger
}

// NewDeleteChapterUseCase creates a new DeleteChapterUseCase
func NewDeleteChapterUseCase(
	chapterRepo repositories.ChapterRepository,
	logger logger.Logger,
) *DeleteChapterUseCase {
	return &DeleteChapterUseCase{
		chapterRepo: chapterRepo,
		logger:      logger,
	}
}

// DeleteChapterInput represents the input for deleting a chapter
type DeleteChapterInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes a chapter
func (uc *DeleteChapterUseCase) Execute(ctx context.Context, input DeleteChapterInput) error {
	// Check if chapter exists
	_, err := uc.chapterRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	if err := uc.chapterRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete chapter", "error", err, "chapter_id", input.ID, "tenant_id", input.TenantID)
		return err
	}

	uc.logger.Info("chapter deleted", "chapter_id", input.ID, "tenant_id", input.TenantID)

	return nil
}

