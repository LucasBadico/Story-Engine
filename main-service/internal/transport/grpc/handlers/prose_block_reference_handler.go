package handlers

import (
	"context"

	"github.com/google/uuid"
	proseblockapp "github.com/story-engine/main-service/internal/application/story/prose_block"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	prosepb "github.com/story-engine/main-service/proto/prose"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProseBlockReferenceHandler implements the ProseBlockReferenceService gRPC service
type ProseBlockReferenceHandler struct {
	prosepb.UnimplementedProseBlockReferenceServiceServer
	createReferenceUC              *proseblockapp.CreateProseBlockReferenceUseCase
	listByProseBlockUC            *proseblockapp.ListProseBlockReferencesByProseBlockUseCase
	listByEntityUC                *proseblockapp.ListProseBlocksByEntityUseCase
	deleteReferenceUC             *proseblockapp.DeleteProseBlockReferenceUseCase
	logger                         logger.Logger
}

// NewProseBlockReferenceHandler creates a new ProseBlockReferenceHandler
func NewProseBlockReferenceHandler(
	createReferenceUC *proseblockapp.CreateProseBlockReferenceUseCase,
	listByProseBlockUC *proseblockapp.ListProseBlockReferencesByProseBlockUseCase,
	listByEntityUC *proseblockapp.ListProseBlocksByEntityUseCase,
	deleteReferenceUC *proseblockapp.DeleteProseBlockReferenceUseCase,
	logger logger.Logger,
) *ProseBlockReferenceHandler {
	return &ProseBlockReferenceHandler{
		createReferenceUC:  createReferenceUC,
		listByProseBlockUC: listByProseBlockUC,
		listByEntityUC:     listByEntityUC,
		deleteReferenceUC:  deleteReferenceUC,
		logger:              logger,
	}
}

// CreateProseBlockReference creates a new reference
func (h *ProseBlockReferenceHandler) CreateProseBlockReference(ctx context.Context, req *prosepb.CreateProseBlockReferenceRequest) (*prosepb.CreateProseBlockReferenceResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	proseBlockID, err := uuid.Parse(req.ProseBlockId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid prose_block_id: %v", err)
	}

	entityID, err := uuid.Parse(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id: %v", err)
	}

	output, err := h.createReferenceUC.Execute(ctx, proseblockapp.CreateProseBlockReferenceInput{
		TenantID:     tenantUUID,
		ProseBlockID: proseBlockID,
		EntityType:   story.EntityType(req.EntityType),
		EntityID:     entityID,
	})
	if err != nil {
		return nil, err
	}

	return &prosepb.CreateProseBlockReferenceResponse{
		Reference: proseBlockReferenceToProto(output.Reference),
	}, nil
}

// ListProseBlockReferencesByProseBlock lists references for a prose block
func (h *ProseBlockReferenceHandler) ListProseBlockReferencesByProseBlock(ctx context.Context, req *prosepb.ListProseBlockReferencesByProseBlockRequest) (*prosepb.ListProseBlockReferencesByProseBlockResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	proseBlockID, err := uuid.Parse(req.ProseBlockId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid prose_block_id: %v", err)
	}

	output, err := h.listByProseBlockUC.Execute(ctx, proseblockapp.ListProseBlockReferencesByProseBlockInput{
		TenantID:     tenantUUID,
		ProseBlockID: proseBlockID,
	})
	if err != nil {
		return nil, err
	}

	protoRefs := make([]*prosepb.ProseBlockReference, len(output.References))
	for i, ref := range output.References {
		protoRefs[i] = proseBlockReferenceToProto(ref)
	}

	return &prosepb.ListProseBlockReferencesByProseBlockResponse{
		References: protoRefs,
		TotalCount: int32(output.Total),
	}, nil
}

// ListProseBlocksByEntity lists prose blocks associated with an entity
func (h *ProseBlockReferenceHandler) ListProseBlocksByEntity(ctx context.Context, req *prosepb.ListProseBlocksByEntityRequest) (*prosepb.ListProseBlocksByEntityResponse, error) {
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

	output, err := h.listByEntityUC.Execute(ctx, proseblockapp.ListProseBlocksByEntityInput{
		TenantID:   tenantUUID,
		EntityType: story.EntityType(req.EntityType),
		EntityID:   entityID,
	})
	if err != nil {
		return nil, err
	}

	protoProseBlocks := make([]*prosepb.ProseBlock, len(output.ProseBlocks))
	for i, p := range output.ProseBlocks {
		protoProseBlocks[i] = proseBlockToProto(p)
	}

	return &prosepb.ListProseBlocksByEntityResponse{
		ProseBlocks: protoProseBlocks,
		TotalCount:  int32(output.Total),
	}, nil
}

// DeleteProseBlockReference deletes a reference
func (h *ProseBlockReferenceHandler) DeleteProseBlockReference(ctx context.Context, req *prosepb.DeleteProseBlockReferenceRequest) (*prosepb.DeleteProseBlockReferenceResponse, error) {
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

	if err := h.deleteReferenceUC.Execute(ctx, proseblockapp.DeleteProseBlockReferenceInput{
		TenantID: tenantUUID,
		ID:       id,
	}); err != nil {
		return nil, err
	}

	return &prosepb.DeleteProseBlockReferenceResponse{
		Success: true,
	}, nil
}

// proseBlockReferenceToProto converts a domain ProseBlockReference to a proto message
func proseBlockReferenceToProto(ref *story.ProseBlockReference) *prosepb.ProseBlockReference {
	return &prosepb.ProseBlockReference{
		Id:           ref.ID.String(),
		ProseBlockId: ref.ProseBlockID.String(),
		EntityType:   string(ref.EntityType),
		EntityId:     ref.EntityID.String(),
		CreatedAt:    timestamppb.New(ref.CreatedAt),
	}
}


