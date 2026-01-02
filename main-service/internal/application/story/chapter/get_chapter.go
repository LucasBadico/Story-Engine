package chapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetChapterUseCase handles retrieving a chapter
type GetChapterUseCase struct {
	chapterRepo repositories.ChapterRepository
	logger      logger.Logger
}

// NewGetChapterUseCase creates a new GetChapterUseCase
func NewGetChapterUseCase(
	chapterRepo repositories.ChapterRepository,
	logger logger.Logger,
) *GetChapterUseCase {
	return &GetChapterUseCase{
		chapterRepo: chapterRepo,
		logger:      logger,
	}
}

// GetChapterInput represents the input for getting a chapter
type GetChapterInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// GetChapterOutput represents the output of getting a chapter
type GetChapterOutput struct {
	Chapter *story.Chapter
}

// Execute retrieves a chapter by ID
func (uc *GetChapterUseCase) Execute(ctx context.Context, input GetChapterInput) (*GetChapterOutput, error) {
	chapter, err := uc.chapterRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get chapter", "error", err, "chapter_id", input.ID, "tenant_id", input.TenantID)
		return nil, err
	}

	return &GetChapterOutput{
		Chapter: chapter,
	}, nil
}

