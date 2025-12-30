package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/application/story"
	grpcctx "github.com/story-engine/main-service/internal/transport/grpc"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StoryHandler implements the StoryService gRPC service
// Note: This handler depends on generated proto code.
// After running `make proto-gen`, update the method signatures to use actual proto types.
type StoryHandler struct {
	// pb.UnimplementedStoryServiceServer // Uncomment after proto generation
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
// TODO: Update signature after proto generation
func (h *StoryHandler) CreateStory(ctx context.Context, req interface{}) (interface{}, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		// Also check request for tenant_id (fallback)
		reqMap, _ := req.(map[string]interface{})
		if tenantIDStr, ok := reqMap["tenant_id"].(string); ok && tenantIDStr != "" {
			tenantID = tenantIDStr
		} else {
			return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
		}
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	reqMap, _ := req.(map[string]interface{})
	title, _ := reqMap["title"].(string)
	if title == "" {
		return nil, status.Errorf(codes.InvalidArgument, "title is required")
	}

	var createdByUserID *uuid.UUID
	if createdByStr, ok := reqMap["created_by_user_id"].(string); ok && createdByStr != "" {
		if id, err := uuid.Parse(createdByStr); err == nil {
			createdByUserID = &id
		}
	}

	input := story.CreateStoryInput{
		TenantID:       tenantUUID,
		Title:          title,
		CreatedByUserID: createdByUserID,
	}

	output, err := h.createStoryUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"story": mappers.StoryToProto(output.Story),
	}, nil
}

// GetStory retrieves a story by ID
// TODO: Update signature after proto generation
func (h *StoryHandler) GetStory(ctx context.Context, req interface{}) (interface{}, error) {
	reqMap, _ := req.(map[string]interface{})
	idStr, _ := reqMap["id"].(string)
	
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid story id: %v", err)
	}

	story, err := h.storyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"story": mappers.StoryToProto(story),
	}, nil
}

// ListStories lists stories for a tenant with pagination
// TODO: Update signature after proto generation
func (h *StoryHandler) ListStories(ctx context.Context, req interface{}) (interface{}, error) {
	// Extract tenant_id from context
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		reqMap, _ := req.(map[string]interface{})
		if tenantIDStr, ok := reqMap["tenant_id"].(string); ok && tenantIDStr != "" {
			tenantID = tenantIDStr
		} else {
			return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
		}
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	reqMap, _ := req.(map[string]interface{})
	limit := 10
	offset := 0
	
	if pagination, ok := reqMap["pagination"].(map[string]interface{}); ok {
		if l, ok := pagination["limit"].(int32); ok {
			limit = int(l)
		}
		if o, ok := pagination["offset"].(int32); ok {
			offset = int(o)
		}
	}

	stories, err := h.storyRepo.ListByTenant(ctx, tenantUUID, limit, offset)
	if err != nil {
		return nil, err
	}

	totalCount, err := h.storyRepo.CountByTenant(ctx, tenantUUID)
	if err != nil {
		return nil, err
	}

	protoStories := make([]interface{}, len(stories))
	for i, s := range stories {
		protoStories[i] = mappers.StoryToProto(s)
	}

	return map[string]interface{}{
		"stories":     protoStories,
		"total_count": int32(totalCount),
	}, nil
}

// CloneStory clones a story to create a new version
// TODO: Update signature after proto generation
func (h *StoryHandler) CloneStory(ctx context.Context, req interface{}) (interface{}, error) {
	reqMap, _ := req.(map[string]interface{})
	sourceStoryIDStr, _ := reqMap["source_story_id"].(string)
	
	sourceStoryID, err := uuid.Parse(sourceStoryIDStr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid source_story_id: %v", err)
	}

	var createdByUserID *uuid.UUID
	if createdByStr, ok := reqMap["created_by_user_id"].(string); ok && createdByStr != "" {
		if id, err := uuid.Parse(createdByStr); err == nil {
			createdByUserID = &id
		}
	}

	input := story.CloneStoryInput{
		SourceStoryID:   sourceStoryID,
		CreatedByUserID: createdByUserID,
	}

	output, err := h.cloneStoryUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	// Get the cloned story to return
	clonedStory, err := h.storyRepo.GetByID(ctx, output.NewStoryID)
	if err != nil {
		return nil, err
	}

	// Get version number from the cloned story
	versions, err := h.storyRepo.ListVersionsByRoot(ctx, clonedStory.RootStoryID)
	if err != nil {
		return nil, err
	}
	versionNumber := len(versions)

	return map[string]interface{}{
		"story":            mappers.StoryToProto(clonedStory),
		"new_version_number": int32(versionNumber),
	}, nil
}

// ListStoryVersions lists all versions of a story
// TODO: Update signature after proto generation
func (h *StoryHandler) ListStoryVersions(ctx context.Context, req interface{}) (interface{}, error) {
	reqMap, _ := req.(map[string]interface{})
	rootStoryIDStr, _ := reqMap["root_story_id"].(string)
	
	rootStoryID, err := uuid.Parse(rootStoryIDStr)
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

	protoVersions := make([]interface{}, len(output.Versions))
	for i, v := range output.Versions {
		protoVersions[i] = mappers.StoryToProto(v)
	}

	return map[string]interface{}{
		"versions": protoVersions,
	}, nil
}

