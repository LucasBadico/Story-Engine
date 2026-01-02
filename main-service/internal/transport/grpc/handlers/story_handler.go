package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/application/story"
	storycore "github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	"github.com/story-engine/main-service/internal/platform/logger"
	storypb "github.com/story-engine/main-service/proto/story"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StoryHandler implements the StoryService gRPC service
type StoryHandler struct {
	storypb.UnimplementedStoryServiceServer
	createStoryUseCase  *story.CreateStoryUseCase
	getStoryUseCase     *story.GetStoryUseCase
	updateStoryUseCase  *story.UpdateStoryUseCase
	listStoriesUseCase  *story.ListStoriesUseCase
	cloneStoryUseCase   *story.CloneStoryUseCase
	versionGraphUseCase *story.GetStoryVersionGraphUseCase
	logger              logger.Logger
}

// NewStoryHandler creates a new StoryHandler
func NewStoryHandler(
	createStoryUseCase *story.CreateStoryUseCase,
	getStoryUseCase *story.GetStoryUseCase,
	updateStoryUseCase *story.UpdateStoryUseCase,
	listStoriesUseCase *story.ListStoriesUseCase,
	cloneStoryUseCase *story.CloneStoryUseCase,
	versionGraphUseCase *story.GetStoryVersionGraphUseCase,
	logger logger.Logger,
) *StoryHandler {
	return &StoryHandler{
		createStoryUseCase:  createStoryUseCase,
		getStoryUseCase:     getStoryUseCase,
		updateStoryUseCase:  updateStoryUseCase,
		listStoriesUseCase:  listStoriesUseCase,
		cloneStoryUseCase:   cloneStoryUseCase,
		versionGraphUseCase: versionGraphUseCase,
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid story id: %v", err)
	}

	output, err := h.getStoryUseCase.Execute(ctx, story.GetStoryInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &storypb.GetStoryResponse{
		Story: mappers.StoryToProto(output.Story),
	}, nil
}

// UpdateStory updates an existing story
func (h *StoryHandler) UpdateStory(ctx context.Context, req *storypb.UpdateStoryRequest) (*storypb.UpdateStoryResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid story id: %v", err)
	}

	input := story.UpdateStoryInput{
		TenantID: tenantUUID,
		ID:       id,
		Title:    req.Title,
	}
	if req.Status != nil {
		status := storycore.StoryStatus(*req.Status)
		input.Status = &status
	}

	output, err := h.updateStoryUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &storypb.UpdateStoryResponse{
		Story: mappers.StoryToProto(output.Story),
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

	output, err := h.listStoriesUseCase.Execute(ctx, story.ListStoriesInput{
		TenantID: tenantUUID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}

	// Convert to proto
	protoStories := make([]*storypb.Story, len(output.Stories))
	for i, s := range output.Stories {
		protoStories[i] = mappers.StoryToProto(s)
	}

	return &storypb.ListStoriesResponse{
		Stories:    protoStories,
		TotalCount: int32(output.Total),
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

	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	input := story.CloneStoryInput{
		TenantID:        tenantUUID,
		SourceStoryID:   sourceStoryID,
		CreatedByUserID: createdBy,
	}

	output, err := h.cloneStoryUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	// Get the new story to return in response
	getOutput, err := h.getStoryUseCase.Execute(ctx, story.GetStoryInput{
		TenantID: tenantUUID,
		ID:       output.NewStoryID,
	})
	if err != nil {
		return nil, err
	}

	return &storypb.CloneStoryResponse{
		Story:            mappers.StoryToProto(getOutput.Story),
		NewVersionNumber: int32(getOutput.Story.VersionNumber),
	}, nil
}

// ListStoryVersions lists all versions of a story
func (h *StoryHandler) ListStoryVersions(ctx context.Context, req *storypb.ListStoryVersionsRequest) (*storypb.ListStoryVersionsResponse, error) {
	rootStoryID, err := uuid.Parse(req.RootStoryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid root_story_id: %v", err)
	}

	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	input := story.GetStoryVersionGraphInput{
		TenantID:    tenantUUID,
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
