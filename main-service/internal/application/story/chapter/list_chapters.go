package chapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListChaptersUseCase handles listing chapters
type ListChaptersUseCase struct {
	chapterRepo repositories.ChapterRepository
	logger      logger.Logger
}

// NewListChaptersUseCase creates a new ListChaptersUseCase
func NewListChaptersUseCase(
	chapterRepo repositories.ChapterRepository,
	logger logger.Logger,
) *ListChaptersUseCase {
	return &ListChaptersUseCase{
		chapterRepo: chapterRepo,
		logger:      logger,
	}
}

// ListChaptersInput represents the input for listing chapters
type ListChaptersInput struct {
	TenantID uuid.UUID
	StoryID  uuid.UUID
}

// ListChaptersOutput represents the output of listing chapters
type ListChaptersOutput struct {
	Chapters []*story.Chapter
	Total    int
}

// Execute lists chapters for a story
func (uc *ListChaptersUseCase) Execute(ctx context.Context, input ListChaptersInput) (*ListChaptersOutput, error) {
	chapters, err := uc.chapterRepo.ListByStory(ctx, input.TenantID, input.StoryID)
	if err != nil {
		uc.logger.Error("failed to list chapters", "error", err, "story_id", input.StoryID, "tenant_id", input.TenantID)
		return nil, err
	}

	return &ListChaptersOutput{
		Chapters: chapters,
		Total:    len(chapters),
	}, nil
}

