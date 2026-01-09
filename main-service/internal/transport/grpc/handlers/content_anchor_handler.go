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

// ContentAnchorHandler implements the ContentAnchorService gRPC service
type ContentAnchorHandler struct {
	contentblockpb.UnimplementedContentAnchorServiceServer
	createAnchorUC       *contentblockapp.CreateContentAnchorUseCase
	listByContentBlockUC *contentblockapp.ListContentAnchorsByContentBlockUseCase
	listByEntityUC       *contentblockapp.ListContentBlocksByEntityUseCase
	deleteAnchorUC       *contentblockapp.DeleteContentAnchorUseCase
	logger               logger.Logger
}

// NewContentAnchorHandler creates a new ContentAnchorHandler
func NewContentAnchorHandler(
	createAnchorUC *contentblockapp.CreateContentAnchorUseCase,
	listByContentBlockUC *contentblockapp.ListContentAnchorsByContentBlockUseCase,
	listByEntityUC *contentblockapp.ListContentBlocksByEntityUseCase,
	deleteAnchorUC *contentblockapp.DeleteContentAnchorUseCase,
	logger logger.Logger,
) *ContentAnchorHandler {
	return &ContentAnchorHandler{
		createAnchorUC:       createAnchorUC,
		listByContentBlockUC: listByContentBlockUC,
		listByEntityUC:       listByEntityUC,
		deleteAnchorUC:       deleteAnchorUC,
		logger:               logger,
	}
}

// CreateContentAnchor creates a new anchor
func (h *ContentAnchorHandler) CreateContentAnchor(ctx context.Context, req *contentblockpb.CreateContentAnchorRequest) (*contentblockpb.CreateContentAnchorResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	contentBlockID, err := uuid.Parse(req.ContentBlockId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid content_block_id: %v", err)
	}

	entityID, err := uuid.Parse(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id: %v", err)
	}

	output, err := h.createAnchorUC.Execute(ctx, contentblockapp.CreateContentAnchorInput{
		TenantID:       tenantUUID,
		ContentBlockID: contentBlockID,
		EntityType:     story.EntityType(req.EntityType),
		EntityID:       entityID,
	})
	if err != nil {
		return nil, err
	}

	return &contentblockpb.CreateContentAnchorResponse{
		Anchor: mappers.ContentAnchorToProto(output.Anchor),
	}, nil
}

// ListContentAnchorsByContentBlock lists anchors for a content block
func (h *ContentAnchorHandler) ListContentAnchorsByContentBlock(ctx context.Context, req *contentblockpb.ListContentAnchorsByContentBlockRequest) (*contentblockpb.ListContentAnchorsByContentBlockResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	contentBlockID, err := uuid.Parse(req.ContentBlockId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid content_block_id: %v", err)
	}

	output, err := h.listByContentBlockUC.Execute(ctx, contentblockapp.ListContentAnchorsByContentBlockInput{
		TenantID:       tenantUUID,
		ContentBlockID: contentBlockID,
	})
	if err != nil {
		return nil, err
	}

	protoAnchors := make([]*contentblockpb.ContentAnchor, len(output.Anchors))
	for i, anchor := range output.Anchors {
		protoAnchors[i] = mappers.ContentAnchorToProto(anchor)
	}

	return &contentblockpb.ListContentAnchorsByContentBlockResponse{
		Anchors:    protoAnchors,
		TotalCount: int32(output.Total),
	}, nil
}

// ListContentBlocksByEntity lists content blocks associated with an entity
func (h *ContentAnchorHandler) ListContentBlocksByEntity(ctx context.Context, req *contentblockpb.ListContentBlocksByEntityRequest) (*contentblockpb.ListContentBlocksByEntityResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	entityID, err := uuid.Parse(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id: %v", err)
	}

	output, err := h.listByEntityUC.Execute(ctx, contentblockapp.ListContentBlocksByEntityInput{
		TenantID:   tenantUUID,
		EntityType: story.EntityType(req.EntityType),
		EntityID:   entityID,
	})
	if err != nil {
		return nil, err
	}

	protoContentBlocks := make([]*contentblockpb.ContentBlock, len(output.ContentBlocks))
	for i, block := range output.ContentBlocks {
		protoContentBlocks[i] = mappers.ContentBlockToProto(block)
	}

	return &contentblockpb.ListContentBlocksByEntityResponse{
		ContentBlocks: protoContentBlocks,
		TotalCount:    int32(output.Total),
	}, nil
}

// DeleteContentAnchor deletes an anchor
func (h *ContentAnchorHandler) DeleteContentAnchor(ctx context.Context, req *contentblockpb.DeleteContentAnchorRequest) (*contentblockpb.DeleteContentAnchorResponse, error) {
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

	if err := h.deleteAnchorUC.Execute(ctx, contentblockapp.DeleteContentAnchorInput{
		TenantID: tenantUUID,
		ID:       id,
	}); err != nil {
		return nil, err
	}

	return &contentblockpb.DeleteContentAnchorResponse{
		Success: true,
	}, nil
}


