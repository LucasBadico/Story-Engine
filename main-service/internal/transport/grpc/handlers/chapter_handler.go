package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
	chapterpb "github.com/story-engine/main-service/proto/chapter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ChapterHandler implements the ChapterService gRPC service
type ChapterHandler struct {
	chapterpb.UnimplementedChapterServiceServer
	chapterRepo repositories.ChapterRepository
	storyRepo   repositories.StoryRepository
	logger      logger.Logger
}

// NewChapterHandler creates a new ChapterHandler
func NewChapterHandler(
	chapterRepo repositories.ChapterRepository,
	storyRepo repositories.StoryRepository,
	logger logger.Logger,
) *ChapterHandler {
	return &ChapterHandler{
		chapterRepo: chapterRepo,
		storyRepo:   storyRepo,
		logger:      logger,
	}
}

// CreateChapter creates a new chapter
func (h *ChapterHandler) CreateChapter(ctx context.Context, req *chapterpb.CreateChapterRequest) (*chapterpb.CreateChapterResponse, error) {
	storyID, err := uuid.Parse(req.StoryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid story_id: %v", err)
	}

	// Validate story exists
	_, err = h.storyRepo.GetByID(ctx, storyID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "story not found: %v", err)
	}

	if req.Number < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "number must be greater than 0")
	}

	if req.Title == "" {
		return nil, status.Errorf(codes.InvalidArgument, "title is required")
	}

	chapter, err := story.NewChapter(storyID, int(req.Number), req.Title)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid chapter: %v", err)
	}

	// Set status if provided
	if req.Status != "" {
		if err := chapter.UpdateStatus(story.ChapterStatus(req.Status)); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid status: %v", err)
		}
	}

	if err := h.chapterRepo.Create(ctx, chapter); err != nil {
		h.logger.Error("failed to create chapter", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create chapter: %v", err)
	}

	return &chapterpb.CreateChapterResponse{
		Chapter: chapterToProto(chapter),
	}, nil
}

// GetChapter retrieves a chapter by ID
func (h *ChapterHandler) GetChapter(ctx context.Context, req *chapterpb.GetChapterRequest) (*chapterpb.GetChapterResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	chapter, err := h.chapterRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "chapter not found: %v", err)
	}

	return &chapterpb.GetChapterResponse{
		Chapter: chapterToProto(chapter),
	}, nil
}

// UpdateChapter updates an existing chapter
func (h *ChapterHandler) UpdateChapter(ctx context.Context, req *chapterpb.UpdateChapterRequest) (*chapterpb.UpdateChapterResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	chapter, err := h.chapterRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "chapter not found: %v", err)
	}

	// Update fields if provided
	if req.Number != nil {
		if *req.Number < 1 {
			return nil, status.Errorf(codes.InvalidArgument, "number must be greater than 0")
		}
		chapter.Number = int(*req.Number)
	}

	if req.Title != nil {
		chapter.UpdateTitle(*req.Title)
	}

	if req.Status != nil {
		if err := chapter.UpdateStatus(story.ChapterStatus(*req.Status)); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid status: %v", err)
		}
	}

	if err := h.chapterRepo.Update(ctx, chapter); err != nil {
		h.logger.Error("failed to update chapter", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to update chapter: %v", err)
	}

	return &chapterpb.UpdateChapterResponse{
		Chapter: chapterToProto(chapter),
	}, nil
}

// DeleteChapter deletes a chapter
func (h *ChapterHandler) DeleteChapter(ctx context.Context, req *chapterpb.DeleteChapterRequest) (*chapterpb.DeleteChapterResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	// Check if chapter exists
	_, err = h.chapterRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "chapter not found: %v", err)
	}

	if err := h.chapterRepo.Delete(ctx, id); err != nil {
		h.logger.Error("failed to delete chapter", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to delete chapter: %v", err)
	}

	return &chapterpb.DeleteChapterResponse{
		Success: true,
	}, nil
}

// ListChaptersByStory lists all chapters for a story
func (h *ChapterHandler) ListChaptersByStory(ctx context.Context, req *chapterpb.ListChaptersByStoryRequest) (*chapterpb.ListChaptersByStoryResponse, error) {
	storyID, err := uuid.Parse(req.StoryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid story_id: %v", err)
	}

	chapters, err := h.chapterRepo.ListByStory(ctx, storyID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list chapters: %v", err)
	}

	protoChapters := make([]*chapterpb.Chapter, len(chapters))
	for i, c := range chapters {
		protoChapters[i] = chapterToProto(c)
	}

	return &chapterpb.ListChaptersByStoryResponse{
		Chapters:   protoChapters,
		TotalCount: int32(len(chapters)),
	}, nil
}

// chapterToProto converts a domain Chapter to a proto Chapter message
func chapterToProto(c *story.Chapter) *chapterpb.Chapter {
	return &chapterpb.Chapter{
		Id:        c.ID.String(),
		StoryId:   c.StoryID.String(),
		Number:    int32(c.Number),
		Title:     c.Title,
		Status:    string(c.Status),
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
}

