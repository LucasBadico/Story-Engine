package handlers

import (
	"context"

	"github.com/google/uuid"
	chapterapp "github.com/story-engine/main-service/internal/application/story/chapter"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	chapterpb "github.com/story-engine/main-service/proto/chapter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ChapterHandler implements the ChapterService gRPC service
type ChapterHandler struct {
	chapterpb.UnimplementedChapterServiceServer
	createChapterUseCase *chapterapp.CreateChapterUseCase
	getChapterUseCase    *chapterapp.GetChapterUseCase
	updateChapterUseCase *chapterapp.UpdateChapterUseCase
	deleteChapterUseCase *chapterapp.DeleteChapterUseCase
	listChaptersUseCase  *chapterapp.ListChaptersUseCase
	logger               logger.Logger
}

// NewChapterHandler creates a new ChapterHandler
func NewChapterHandler(
	createChapterUseCase *chapterapp.CreateChapterUseCase,
	getChapterUseCase *chapterapp.GetChapterUseCase,
	updateChapterUseCase *chapterapp.UpdateChapterUseCase,
	deleteChapterUseCase *chapterapp.DeleteChapterUseCase,
	listChaptersUseCase *chapterapp.ListChaptersUseCase,
	logger logger.Logger,
) *ChapterHandler {
	return &ChapterHandler{
		createChapterUseCase: createChapterUseCase,
		getChapterUseCase:    getChapterUseCase,
		updateChapterUseCase: updateChapterUseCase,
		deleteChapterUseCase: deleteChapterUseCase,
		listChaptersUseCase:  listChaptersUseCase,
		logger:               logger,
	}
}

// CreateChapter creates a new chapter
func (h *ChapterHandler) CreateChapter(ctx context.Context, req *chapterpb.CreateChapterRequest) (*chapterpb.CreateChapterResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	storyID, err := uuid.Parse(req.StoryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid story_id: %v", err)
	}

	input := chapterapp.CreateChapterInput{
		TenantID: tenantUUID,
		StoryID:  storyID,
		Number:   int(req.Number),
		Title:    req.Title,
	}
	if req.Status != "" {
		status := story.ChapterStatus(req.Status)
		input.Status = &status
	}

	output, err := h.createChapterUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &chapterpb.CreateChapterResponse{
		Chapter: chapterToProto(output.Chapter),
	}, nil
}

// GetChapter retrieves a chapter by ID
func (h *ChapterHandler) GetChapter(ctx context.Context, req *chapterpb.GetChapterRequest) (*chapterpb.GetChapterResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	output, err := h.getChapterUseCase.Execute(ctx, chapterapp.GetChapterInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &chapterpb.GetChapterResponse{
		Chapter: chapterToProto(output.Chapter),
	}, nil
}

// UpdateChapter updates an existing chapter
func (h *ChapterHandler) UpdateChapter(ctx context.Context, req *chapterpb.UpdateChapterRequest) (*chapterpb.UpdateChapterResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	var number *int
	if req.Number != nil {
		n := int(*req.Number)
		number = &n
	}

	input := chapterapp.UpdateChapterInput{
		TenantID: tenantUUID,
		ID:       id,
		Number:   number,
		Title:    req.Title,
	}
	if req.Status != nil {
		status := story.ChapterStatus(*req.Status)
		input.Status = &status
	}

	output, err := h.updateChapterUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &chapterpb.UpdateChapterResponse{
		Chapter: chapterToProto(output.Chapter),
	}, nil
}

// DeleteChapter deletes a chapter
func (h *ChapterHandler) DeleteChapter(ctx context.Context, req *chapterpb.DeleteChapterRequest) (*chapterpb.DeleteChapterResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	if err := h.deleteChapterUseCase.Execute(ctx, chapterapp.DeleteChapterInput{
		TenantID: tenantUUID,
		ID:       id,
	}); err != nil {
		return nil, err
	}

	return &chapterpb.DeleteChapterResponse{
		Success: true,
	}, nil
}

// ListChaptersByStory lists all chapters for a story
func (h *ChapterHandler) ListChaptersByStory(ctx context.Context, req *chapterpb.ListChaptersByStoryRequest) (*chapterpb.ListChaptersByStoryResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	storyID, err := uuid.Parse(req.StoryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid story_id: %v", err)
	}

	output, err := h.listChaptersUseCase.Execute(ctx, chapterapp.ListChaptersInput{
		TenantID: tenantUUID,
		StoryID:  storyID,
	})
	if err != nil {
		return nil, err
	}

	protoChapters := make([]*chapterpb.Chapter, len(output.Chapters))
	for i, c := range output.Chapters {
		protoChapters[i] = chapterToProto(c)
	}

	return &chapterpb.ListChaptersByStoryResponse{
		Chapters:   protoChapters,
		TotalCount: int32(output.Total),
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


