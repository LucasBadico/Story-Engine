package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
	prosepb "github.com/story-engine/main-service/proto/prose"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProseBlockHandler implements the ProseBlockService gRPC service
type ProseBlockHandler struct {
	prosepb.UnimplementedProseBlockServiceServer
	proseBlockRepo repositories.ProseBlockRepository
	chapterRepo    repositories.ChapterRepository
	logger         logger.Logger
}

// NewProseBlockHandler creates a new ProseBlockHandler
func NewProseBlockHandler(
	proseBlockRepo repositories.ProseBlockRepository,
	chapterRepo repositories.ChapterRepository,
	logger logger.Logger,
) *ProseBlockHandler {
	return &ProseBlockHandler{
		proseBlockRepo: proseBlockRepo,
		chapterRepo:    chapterRepo,
		logger:         logger,
	}
}

// CreateProseBlock creates a new prose block
func (h *ProseBlockHandler) CreateProseBlock(ctx context.Context, req *prosepb.CreateProseBlockRequest) (*prosepb.CreateProseBlockResponse, error) {
	chapterID, err := uuid.Parse(req.ChapterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid chapter_id: %v", err)
	}

	// Validate chapter exists
	_, err = h.chapterRepo.GetByID(ctx, chapterID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "chapter not found: %v", err)
	}

	if req.OrderNum < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "order_num must be greater than 0")
	}

	kind := req.Kind
	if kind == "" {
		kind = "final"
	}

	if req.Content == "" {
		return nil, status.Errorf(codes.InvalidArgument, "content is required")
	}

	proseBlock, err := story.NewProseBlock(chapterID, int(req.OrderNum), story.ProseKind(kind), req.Content)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid prose block: %v", err)
	}

	if err := h.proseBlockRepo.Create(ctx, proseBlock); err != nil {
		h.logger.Error("failed to create prose block", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create prose block: %v", err)
	}

	return &prosepb.CreateProseBlockResponse{
		ProseBlock: proseBlockToProto(proseBlock),
	}, nil
}

// GetProseBlock retrieves a prose block by ID
func (h *ProseBlockHandler) GetProseBlock(ctx context.Context, req *prosepb.GetProseBlockRequest) (*prosepb.GetProseBlockResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	proseBlock, err := h.proseBlockRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "prose block not found: %v", err)
	}

	return &prosepb.GetProseBlockResponse{
		ProseBlock: proseBlockToProto(proseBlock),
	}, nil
}

// UpdateProseBlock updates an existing prose block
func (h *ProseBlockHandler) UpdateProseBlock(ctx context.Context, req *prosepb.UpdateProseBlockRequest) (*prosepb.UpdateProseBlockResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	proseBlock, err := h.proseBlockRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "prose block not found: %v", err)
	}

	// Update fields if provided
	if req.OrderNum != nil {
		if *req.OrderNum < 1 {
			return nil, status.Errorf(codes.InvalidArgument, "order_num must be greater than 0")
		}
		proseBlock.OrderNum = int(*req.OrderNum)
	}

	if req.Kind != nil {
		proseBlock.Kind = story.ProseKind(*req.Kind)
	}

	if req.Content != nil {
		proseBlock.UpdateContent(*req.Content)
	}

	if err := proseBlock.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid prose block: %v", err)
	}

	if err := h.proseBlockRepo.Update(ctx, proseBlock); err != nil {
		h.logger.Error("failed to update prose block", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to update prose block: %v", err)
	}

	return &prosepb.UpdateProseBlockResponse{
		ProseBlock: proseBlockToProto(proseBlock),
	}, nil
}

// DeleteProseBlock deletes a prose block
func (h *ProseBlockHandler) DeleteProseBlock(ctx context.Context, req *prosepb.DeleteProseBlockRequest) (*prosepb.DeleteProseBlockResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	// Check if prose block exists
	_, err = h.proseBlockRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "prose block not found: %v", err)
	}

	if err := h.proseBlockRepo.Delete(ctx, id); err != nil {
		h.logger.Error("failed to delete prose block", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to delete prose block: %v", err)
	}

	return &prosepb.DeleteProseBlockResponse{
		Success: true,
	}, nil
}

// ListProseBlocksByChapter lists prose blocks for a chapter
func (h *ProseBlockHandler) ListProseBlocksByChapter(ctx context.Context, req *prosepb.ListProseBlocksByChapterRequest) (*prosepb.ListProseBlocksByChapterResponse, error) {
	chapterID, err := uuid.Parse(req.ChapterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid chapter_id: %v", err)
	}

	proseBlocks, err := h.proseBlockRepo.ListByChapter(ctx, chapterID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list prose blocks: %v", err)
	}

	protoProseBlocks := make([]*prosepb.ProseBlock, len(proseBlocks))
	for i, p := range proseBlocks {
		protoProseBlocks[i] = proseBlockToProto(p)
	}

	return &prosepb.ListProseBlocksByChapterResponse{
		ProseBlocks: protoProseBlocks,
		TotalCount:  int32(len(proseBlocks)),
	}, nil
}

// proseBlockToProto converts a domain ProseBlock to a proto ProseBlock message
func proseBlockToProto(p *story.ProseBlock) *prosepb.ProseBlock {
	return &prosepb.ProseBlock{
		Id:        p.ID.String(),
		ChapterId: p.ChapterID.String(),
		OrderNum:  int32(p.OrderNum),
		Kind:      string(p.Kind),
		Content:   p.Content,
		WordCount: int32(p.WordCount),
		CreatedAt: timestamppb.New(p.CreatedAt),
		UpdatedAt: timestamppb.New(p.UpdatedAt),
	}
}

