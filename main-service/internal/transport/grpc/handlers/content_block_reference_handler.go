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

// ContentBlockReferenceHandler implements the ContentBlockReferenceService gRPC service
type ContentBlockReferenceHandler struct {
	contentblockpb.UnimplementedContentBlockReferenceServiceServer
	createReferenceUC        *contentblockapp.CreateContentBlockReferenceUseCase
	listByContentBlockUC     *contentblockapp.ListContentBlockReferencesByContentBlockUseCase
	listByEntityUC           *contentblockapp.ListContentBlocksByEntityUseCase
	deleteReferenceUC        *contentblockapp.DeleteContentBlockReferenceUseCase
	logger                    logger.Logger
}

// NewContentBlockReferenceHandler creates a new ContentBlockReferenceHandler
func NewContentBlockReferenceHandler(
	createReferenceUC *contentblockapp.CreateContentBlockReferenceUseCase,
	listByContentBlockUC *contentblockapp.ListContentBlockReferencesByContentBlockUseCase,
	listByEntityUC *contentblockapp.ListContentBlocksByEntityUseCase,
	deleteReferenceUC *contentblockapp.DeleteContentBlockReferenceUseCase,
	logger logger.Logger,
) *ContentBlockReferenceHandler {
	return &ContentBlockReferenceHandler{
		createReferenceUC:    createReferenceUC,
		listByContentBlockUC: listByContentBlockUC,
		listByEntityUC:       listByEntityUC,
		deleteReferenceUC:    deleteReferenceUC,
		logger:               logger,
	}
}

// CreateContentBlockReference creates a new reference
func (h *ContentBlockReferenceHandler) CreateContentBlockReference(ctx context.Context, req *contentblockpb.CreateContentBlockReferenceRequest) (*contentblockpb.CreateContentBlockReferenceResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
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

	output, err := h.createReferenceUC.Execute(ctx, contentblockapp.CreateContentBlockReferenceInput{
		TenantID:      tenantUUID,
		ContentBlockID: contentBlockID,
		EntityType:    story.EntityType(req.EntityType),
		EntityID:      entityID,
	})
	if err != nil {
		return nil, err
	}

	return &contentblockpb.CreateContentBlockReferenceResponse{
		Reference: mappers.ContentBlockReferenceToProto(output.Reference),
	}, nil
}

// ListContentBlockReferencesByContentBlock lists references for a content block
func (h *ContentBlockReferenceHandler) ListContentBlockReferencesByContentBlock(ctx context.Context, req *contentblockpb.ListContentBlockReferencesByContentBlockRequest) (*contentblockpb.ListContentBlockReferencesByContentBlockResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
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

	output, err := h.listByContentBlockUC.Execute(ctx, contentblockapp.ListContentBlockReferencesByContentBlockInput{
		TenantID:      tenantUUID,
		ContentBlockID: contentBlockID,
	})
	if err != nil {
		return nil, err
	}

	protoRefs := make([]*contentblockpb.ContentBlockReference, len(output.References))
	for i, ref := range output.References {
		protoRefs[i] = mappers.ContentBlockReferenceToProto(ref)
	}

	return &contentblockpb.ListContentBlockReferencesByContentBlockResponse{
		References: protoRefs,
		TotalCount: int32(output.Total),
	}, nil
}

// ListContentBlocksByEntity lists content blocks associated with an entity
func (h *ContentBlockReferenceHandler) ListContentBlocksByEntity(ctx context.Context, req *contentblockpb.ListContentBlocksByEntityRequest) (*contentblockpb.ListContentBlocksByEntityResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
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
	for i, c := range output.ContentBlocks {
		protoContentBlocks[i] = mappers.ContentBlockToProto(c)
	}

	return &contentblockpb.ListContentBlocksByEntityResponse{
		ContentBlocks: protoContentBlocks,
		TotalCount:    int32(output.Total),
	}, nil
}

// DeleteContentBlockReference deletes a reference
func (h *ContentBlockReferenceHandler) DeleteContentBlockReference(ctx context.Context, req *contentblockpb.DeleteContentBlockReferenceRequest) (*contentblockpb.DeleteContentBlockReferenceResponse, error) {
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

	if err := h.deleteReferenceUC.Execute(ctx, contentblockapp.DeleteContentBlockReferenceInput{
		TenantID: tenantUUID,
		ID:       id,
	}); err != nil {
		return nil, err
	}

	return &contentblockpb.DeleteContentBlockReferenceResponse{
		Success: true,
	}, nil
}
