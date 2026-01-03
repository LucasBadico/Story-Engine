package handlers

import (
	"context"

	"github.com/google/uuid"
	contentblockapp "github.com/story-engine/main-service/internal/application/story/content_block"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	contentblockpb "github.com/story-engine/main-service/proto/content_block"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ContentBlockHandler implements the ContentBlockService gRPC service
type ContentBlockHandler struct {
	contentblockpb.UnimplementedContentBlockServiceServer
	createContentBlockUseCase *contentblockapp.CreateContentBlockUseCase
	getContentBlockUseCase    *contentblockapp.GetContentBlockUseCase
	updateContentBlockUseCase *contentblockapp.UpdateContentBlockUseCase
	deleteContentBlockUseCase *contentblockapp.DeleteContentBlockUseCase
	listContentBlocksUseCase  *contentblockapp.ListContentBlocksUseCase
	logger                    logger.Logger
}

// NewContentBlockHandler creates a new ContentBlockHandler
func NewContentBlockHandler(
	createContentBlockUseCase *contentblockapp.CreateContentBlockUseCase,
	getContentBlockUseCase *contentblockapp.GetContentBlockUseCase,
	updateContentBlockUseCase *contentblockapp.UpdateContentBlockUseCase,
	deleteContentBlockUseCase *contentblockapp.DeleteContentBlockUseCase,
	listContentBlocksUseCase *contentblockapp.ListContentBlocksUseCase,
	logger logger.Logger,
) *ContentBlockHandler {
	return &ContentBlockHandler{
		createContentBlockUseCase: createContentBlockUseCase,
		getContentBlockUseCase:    getContentBlockUseCase,
		updateContentBlockUseCase: updateContentBlockUseCase,
		deleteContentBlockUseCase: deleteContentBlockUseCase,
		listContentBlocksUseCase:  listContentBlocksUseCase,
		logger:                    logger,
	}
}

// CreateContentBlock creates a new content block
func (h *ContentBlockHandler) CreateContentBlock(ctx context.Context, req *contentblockpb.CreateContentBlockRequest) (*contentblockpb.CreateContentBlockResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	var chapterID *uuid.UUID
	if req.ChapterId != nil && *req.ChapterId != "" {
		parsedChapterID, err := uuid.Parse(*req.ChapterId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid chapter_id: %v", err)
		}
		chapterID = &parsedChapterID
	}

	var orderNum *int
	if req.OrderNum != nil {
		order := int(*req.OrderNum)
		orderNum = &order
	}

	contentType := story.ContentType(req.Type)
	if contentType == "" {
		contentType = story.ContentTypeText
	}

	kind := story.ContentKind(req.Kind)
	if kind == "" {
		kind = story.ContentKindFinal
	}

	metadata := story.ContentMetadata{}
	if req.Metadata != nil {
		parsedMetadata, err := mappers.ContentMetadataFromStruct(req.Metadata)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid metadata: %v", err)
		}
		metadata = parsedMetadata
	}

	output, err := h.createContentBlockUseCase.Execute(ctx, contentblockapp.CreateContentBlockInput{
		TenantID:  tenantUUID,
		ChapterID: chapterID,
		OrderNum:  orderNum,
		Type:      contentType,
		Kind:      kind,
		Content:   req.Content,
		Metadata:  metadata,
	})
	if err != nil {
		return nil, err
	}

	return &contentblockpb.CreateContentBlockResponse{
		ContentBlock: mappers.ContentBlockToProto(output.ContentBlock),
	}, nil
}

// GetContentBlock retrieves a content block by ID
func (h *ContentBlockHandler) GetContentBlock(ctx context.Context, req *contentblockpb.GetContentBlockRequest) (*contentblockpb.GetContentBlockResponse, error) {
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

	output, err := h.getContentBlockUseCase.Execute(ctx, contentblockapp.GetContentBlockInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &contentblockpb.GetContentBlockResponse{
		ContentBlock: mappers.ContentBlockToProto(output.ContentBlock),
	}, nil
}

// UpdateContentBlock updates an existing content block
func (h *ContentBlockHandler) UpdateContentBlock(ctx context.Context, req *contentblockpb.UpdateContentBlockRequest) (*contentblockpb.UpdateContentBlockResponse, error) {
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

	var orderNum *int
	if req.OrderNum != nil {
		order := int(*req.OrderNum)
		orderNum = &order
	}

	input := contentblockapp.UpdateContentBlockInput{
		TenantID: tenantUUID,
		ID:       id,
		OrderNum: orderNum,
		Content:  req.Content,
	}

	if req.Type != nil {
		contentType := story.ContentType(*req.Type)
		input.Type = &contentType
	}
	if req.Kind != nil {
		kind := story.ContentKind(*req.Kind)
		input.Kind = &kind
	}
	if req.Metadata != nil {
		metadata, err := mappers.ContentMetadataFromStruct(req.Metadata)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid metadata: %v", err)
		}
		input.Metadata = &metadata
	}

	output, err := h.updateContentBlockUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &contentblockpb.UpdateContentBlockResponse{
		ContentBlock: mappers.ContentBlockToProto(output.ContentBlock),
	}, nil
}

// DeleteContentBlock deletes a content block
func (h *ContentBlockHandler) DeleteContentBlock(ctx context.Context, req *contentblockpb.DeleteContentBlockRequest) (*contentblockpb.DeleteContentBlockResponse, error) {
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

	if err := h.deleteContentBlockUseCase.Execute(ctx, contentblockapp.DeleteContentBlockInput{
		TenantID: tenantUUID,
		ID:       id,
	}); err != nil {
		return nil, err
	}

	return &contentblockpb.DeleteContentBlockResponse{
		Success: true,
	}, nil
}

// ListContentBlocksByChapter lists content blocks for a chapter
func (h *ContentBlockHandler) ListContentBlocksByChapter(ctx context.Context, req *contentblockpb.ListContentBlocksByChapterRequest) (*contentblockpb.ListContentBlocksByChapterResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	chapterID, err := uuid.Parse(req.ChapterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid chapter_id: %v", err)
	}

	output, err := h.listContentBlocksUseCase.Execute(ctx, contentblockapp.ListContentBlocksInput{
		TenantID:  tenantUUID,
		ChapterID: chapterID,
	})
	if err != nil {
		return nil, err
	}

	protoContentBlocks := make([]*contentblockpb.ContentBlock, len(output.ContentBlocks))
	for i, c := range output.ContentBlocks {
		protoContentBlocks[i] = mappers.ContentBlockToProto(c)
	}

	return &contentblockpb.ListContentBlocksByChapterResponse{
		ContentBlocks: protoContentBlocks,
		TotalCount:    int32(output.Total),
	}, nil
}
