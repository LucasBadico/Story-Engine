package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/application/story"
	storycore "github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
	storypb "github.com/story-engine/main-service/proto/story"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StoryHandler implements the StoryService gRPC service
type StoryHandler struct {
	storypb.UnimplementedStoryServiceServer
	createStoryUseCase  *story.CreateStoryUseCase
	cloneStoryUseCase   *story.CloneStoryUseCase
	versionGraphUseCase *story.GetStoryVersionGraphUseCase
	storyRepo           repositories.StoryRepository
	logger              logger.Logger
}

// NewStoryHandler creates a new StoryHandler
func NewStoryHandler(
	createStoryUseCase *story.CreateStoryUseCase,
	cloneStoryUseCase *story.CloneStoryUseCase,
	versionGraphUseCase *story.GetStoryVersionGraphUseCase,
	storyRepo repositories.StoryRepository,
	logger logger.Logger,
) *StoryHandler {
	return &StoryHandler{
		createStoryUseCase:  createStoryUseCase,
		cloneStoryUseCase:   cloneStoryUseCase,
		versionGraphUseCase: versionGraphUseCase,
		storyRepo:           storyRepo,
		logger:              logger,
	}
}

// CreateStory creates a new story (version 1)
func (h *StoryHandler) CreateStory(ctx context.Context, req *storypb.CreateStoryRequest) (*storypb.CreateStoryResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		// Also check request for tenant_id (fallback)
		if req.TenantId != "" {
			tenantID = req.TenantId
		} else {
			return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
		}
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	if req.Title == "" {
		return nil, status.Errorf(codes.InvalidArgument, "title is required")
	}

	var createdBy *uuid.UUID
	if req.CreatedByUserId != "" {
		if id, err := uuid.Parse(req.CreatedByUserId); err == nil {
			createdBy = &id
		} else {
			return nil, status.Errorf(codes.InvalidArgument, "invalid created_by_user_id: %v", err)
		}
	}

	input := story.CreateStoryInput{
		TenantID:        tenantUUID,
		Title:           req.Title,
		CreatedByUserID: createdBy,
	}

	output, err := h.createStoryUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &storypb.CreateStoryResponse{
		Story: mappers.StoryToProto(output.Story),
	}, nil
}

// GetStory retrieves a story by ID
func (h *StoryHandler) GetStory(ctx context.Context, req *storypb.GetStoryRequest) (*storypb.GetStoryResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid story id: %v", err)
	}

	s, err := h.storyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &storypb.GetStoryResponse{
		Story: mappers.StoryToProto(s),
	}, nil
}

// UpdateStory updates an existing story
func (h *StoryHandler) UpdateStory(ctx context.Context, req *storypb.UpdateStoryRequest) (*storypb.UpdateStoryResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid story id: %v", err)
	}

	s, err := h.storyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "story not found: %v", err)
	}

	// Update fields if provided
	if req.Title != nil {
		if err := s.UpdateTitle(*req.Title); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid title: %v", err)
		}
	}

	if req.Status != nil {
		if err := s.UpdateStatus(storycore.StoryStatus(*req.Status)); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid status: %v", err)
		}
	}

	if err := h.storyRepo.Update(ctx, s); err != nil {
		h.logger.Error("failed to update story", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to update story: %v", err)
	}

	return &storypb.UpdateStoryResponse{
		Story: mappers.StoryToProto(s),
	}, nil
}

// ListStories lists stories for a tenant with pagination
func (h *StoryHandler) ListStories(ctx context.Context, req *storypb.ListStoriesRequest) (*storypb.ListStoriesResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		// Also check request for tenant_id (fallback)
		if req.TenantId != "" {
			tenantID = req.TenantId
		} else {
			return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
		}
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	// Extract pagination
	limit := 50 // default
	offset := 0
	if req.Pagination != nil {
		if req.Pagination.Limit > 0 {
			limit = int(req.Pagination.Limit)
		}
		if req.Pagination.Offset > 0 {
			offset = int(req.Pagination.Offset)
		}
	}

	stories, err := h.storyRepo.ListByTenant(ctx, tenantUUID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Convert to proto
	protoStories := make([]*storypb.Story, len(stories))
	for i, s := range stories {
		protoStories[i] = mappers.StoryToProto(s)
	}

	// Get total count (for pagination)
	// Note: This could be optimized with a separate count query
	totalCount := int32(len(stories))

	return &storypb.ListStoriesResponse{
		Stories:    protoStories,
		TotalCount: totalCount,
	}, nil
}

// CloneStory clones a story to create a new version
func (h *StoryHandler) CloneStory(ctx context.Context, req *storypb.CloneStoryRequest) (*storypb.CloneStoryResponse, error) {
	sourceStoryID, err := uuid.Parse(req.SourceStoryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid source_story_id: %v", err)
	}

	var createdBy *uuid.UUID
	if req.CreatedByUserId != "" {
		if id, err := uuid.Parse(req.CreatedByUserId); err == nil {
			createdBy = &id
		} else {
			return nil, status.Errorf(codes.InvalidArgument, "invalid created_by_user_id: %v", err)
		}
	}

	input := story.CloneStoryInput{
		SourceStoryID:   sourceStoryID,
		CreatedByUserID: createdBy,
	}

	output, err := h.cloneStoryUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	// Get the new story to return in response
	newStory, err := h.storyRepo.GetByID(ctx, output.NewStoryID)
	if err != nil {
		return nil, err
	}

	return &storypb.CloneStoryResponse{
		Story:            mappers.StoryToProto(newStory),
		NewVersionNumber: int32(newStory.VersionNumber),
	}, nil
}

// ListStoryVersions lists all versions of a story
func (h *StoryHandler) ListStoryVersions(ctx context.Context, req *storypb.ListStoryVersionsRequest) (*storypb.ListStoryVersionsResponse, error) {
	rootStoryID, err := uuid.Parse(req.RootStoryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid root_story_id: %v", err)
	}

	input := story.GetStoryVersionGraphInput{
		RootStoryID: rootStoryID,
	}

	output, err := h.versionGraphUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert version graph to flat list of versions
	protoVersions := make([]*storypb.Story, len(output.Versions))
	for i, s := range output.Versions {
		protoVersions[i] = mappers.StoryToProto(s)
	}

	return &storypb.ListStoryVersionsResponse{
		Versions: protoVersions,
	}, nil
}
