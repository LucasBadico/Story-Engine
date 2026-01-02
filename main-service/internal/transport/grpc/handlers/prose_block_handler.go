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

// ProseBlockHandler implements the ProseBlockService gRPC service
type ProseBlockHandler struct {
	prosepb.UnimplementedProseBlockServiceServer
	createProseBlockUseCase *proseblockapp.CreateProseBlockUseCase
	getProseBlockUseCase    *proseblockapp.GetProseBlockUseCase
	updateProseBlockUseCase *proseblockapp.UpdateProseBlockUseCase
	deleteProseBlockUseCase *proseblockapp.DeleteProseBlockUseCase
	listProseBlocksUseCase  *proseblockapp.ListProseBlocksUseCase
	logger                   logger.Logger
}

// NewProseBlockHandler creates a new ProseBlockHandler
func NewProseBlockHandler(
	createProseBlockUseCase *proseblockapp.CreateProseBlockUseCase,
	getProseBlockUseCase *proseblockapp.GetProseBlockUseCase,
	updateProseBlockUseCase *proseblockapp.UpdateProseBlockUseCase,
	deleteProseBlockUseCase *proseblockapp.DeleteProseBlockUseCase,
	listProseBlocksUseCase *proseblockapp.ListProseBlocksUseCase,
	logger logger.Logger,
) *ProseBlockHandler {
	return &ProseBlockHandler{
		createProseBlockUseCase: createProseBlockUseCase,
		getProseBlockUseCase:    getProseBlockUseCase,
		updateProseBlockUseCase: updateProseBlockUseCase,
		deleteProseBlockUseCase: deleteProseBlockUseCase,
		listProseBlocksUseCase:  listProseBlocksUseCase,
		logger:                  logger,
	}
}

// CreateProseBlock creates a new prose block
func (h *ProseBlockHandler) CreateProseBlock(ctx context.Context, req *prosepb.CreateProseBlockRequest) (*prosepb.CreateProseBlockResponse, error) {
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

	kind := req.Kind
	if kind == "" {
		kind = "final"
	}

	output, err := h.createProseBlockUseCase.Execute(ctx, proseblockapp.CreateProseBlockInput{
		TenantID:  tenantUUID,
		ChapterID: chapterID,
		OrderNum:  orderNum,
		Kind:      story.ProseKind(kind),
		Content:   req.Content,
	})
	if err != nil {
		return nil, err
	}

	return &prosepb.CreateProseBlockResponse{
		ProseBlock: proseBlockToProto(output.ProseBlock),
	}, nil
}

// GetProseBlock retrieves a prose block by ID
func (h *ProseBlockHandler) GetProseBlock(ctx context.Context, req *prosepb.GetProseBlockRequest) (*prosepb.GetProseBlockResponse, error) {
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

	output, err := h.getProseBlockUseCase.Execute(ctx, proseblockapp.GetProseBlockInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &prosepb.GetProseBlockResponse{
		ProseBlock: proseBlockToProto(output.ProseBlock),
	}, nil
}

// UpdateProseBlock updates an existing prose block
func (h *ProseBlockHandler) UpdateProseBlock(ctx context.Context, req *prosepb.UpdateProseBlockRequest) (*prosepb.UpdateProseBlockResponse, error) {
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

	input := proseblockapp.UpdateProseBlockInput{
		TenantID: tenantUUID,
		ID:       id,
		OrderNum: orderNum,
		Content:  req.Content,
	}
	if req.Kind != nil {
		kind := story.ProseKind(*req.Kind)
		input.Kind = &kind
	}

	output, err := h.updateProseBlockUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &prosepb.UpdateProseBlockResponse{
		ProseBlock: proseBlockToProto(output.ProseBlock),
	}, nil
}

// DeleteProseBlock deletes a prose block
func (h *ProseBlockHandler) DeleteProseBlock(ctx context.Context, req *prosepb.DeleteProseBlockRequest) (*prosepb.DeleteProseBlockResponse, error) {
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

	if err := h.deleteProseBlockUseCase.Execute(ctx, proseblockapp.DeleteProseBlockInput{
		TenantID: tenantUUID,
		ID:       id,
	}); err != nil {
		return nil, err
	}

	return &prosepb.DeleteProseBlockResponse{
		Success: true,
	}, nil
}

// ListProseBlocksByChapter lists prose blocks for a chapter
func (h *ProseBlockHandler) ListProseBlocksByChapter(ctx context.Context, req *prosepb.ListProseBlocksByChapterRequest) (*prosepb.ListProseBlocksByChapterResponse, error) {
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

	output, err := h.listProseBlocksUseCase.Execute(ctx, proseblockapp.ListProseBlocksInput{
		TenantID:  tenantUUID,
		ChapterID: chapterID,
	})
	if err != nil {
		return nil, err
	}

	protoProseBlocks := make([]*prosepb.ProseBlock, len(output.ProseBlocks))
	for i, p := range output.ProseBlocks {
		protoProseBlocks[i] = proseBlockToProto(p)
	}

	return &prosepb.ListProseBlocksByChapterResponse{
		ProseBlocks: protoProseBlocks,
		TotalCount:  int32(output.Total),
	}, nil
}

// proseBlockToProto converts a domain ProseBlock to a proto ProseBlock message
func proseBlockToProto(p *story.ProseBlock) *prosepb.ProseBlock {
	pb := &prosepb.ProseBlock{
		Id:        p.ID.String(),
		Kind:      string(p.Kind),
		Content:   p.Content,
		WordCount: int32(p.WordCount),
		CreatedAt: timestamppb.New(p.CreatedAt),
		UpdatedAt: timestamppb.New(p.UpdatedAt),
	}
	if p.ChapterID != nil {
		chapterIDStr := p.ChapterID.String()
		pb.ChapterId = &chapterIDStr
	}
	if p.OrderNum != nil {
		orderNum := int32(*p.OrderNum)
		pb.OrderNum = &orderNum
	}
	return pb
}

