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

// ProseBlockReferenceHandler implements the ProseBlockReferenceService gRPC service
type ProseBlockReferenceHandler struct {
	prosepb.UnimplementedProseBlockReferenceServiceServer
	refRepo        repositories.ProseBlockReferenceRepository
	proseBlockRepo repositories.ProseBlockRepository
	logger         logger.Logger
}

// NewProseBlockReferenceHandler creates a new ProseBlockReferenceHandler
func NewProseBlockReferenceHandler(
	refRepo repositories.ProseBlockReferenceRepository,
	proseBlockRepo repositories.ProseBlockRepository,
	logger logger.Logger,
) *ProseBlockReferenceHandler {
	return &ProseBlockReferenceHandler{
		refRepo:        refRepo,
		proseBlockRepo: proseBlockRepo,
		logger:         logger,
	}
}

// CreateProseBlockReference creates a new reference
func (h *ProseBlockReferenceHandler) CreateProseBlockReference(ctx context.Context, req *prosepb.CreateProseBlockReferenceRequest) (*prosepb.CreateProseBlockReferenceResponse, error) {
	proseBlockID, err := uuid.Parse(req.ProseBlockId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid prose_block_id: %v", err)
	}

	// Validate prose block exists
	_, err = h.proseBlockRepo.GetByID(ctx, proseBlockID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "prose block not found: %v", err)
	}

	if req.EntityType == "" {
		return nil, status.Errorf(codes.InvalidArgument, "entity_type is required")
	}

	entityID, err := uuid.Parse(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id: %v", err)
	}

	ref, err := story.NewProseBlockReference(proseBlockID, story.EntityType(req.EntityType), entityID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid reference: %v", err)
	}

	if err := h.refRepo.Create(ctx, ref); err != nil {
		h.logger.Error("failed to create prose block reference", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create reference: %v", err)
	}

	return &prosepb.CreateProseBlockReferenceResponse{
		Reference: proseBlockReferenceToProto(ref),
	}, nil
}

// ListProseBlockReferencesByProseBlock lists references for a prose block
func (h *ProseBlockReferenceHandler) ListProseBlockReferencesByProseBlock(ctx context.Context, req *prosepb.ListProseBlockReferencesByProseBlockRequest) (*prosepb.ListProseBlockReferencesByProseBlockResponse, error) {
	proseBlockID, err := uuid.Parse(req.ProseBlockId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid prose_block_id: %v", err)
	}

	references, err := h.refRepo.ListByProseBlock(ctx, proseBlockID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list references: %v", err)
	}

	protoRefs := make([]*prosepb.ProseBlockReference, len(references))
	for i, ref := range references {
		protoRefs[i] = proseBlockReferenceToProto(ref)
	}

	return &prosepb.ListProseBlockReferencesByProseBlockResponse{
		References: protoRefs,
		TotalCount: int32(len(references)),
	}, nil
}

// ListProseBlocksByEntity lists prose blocks associated with an entity
func (h *ProseBlockReferenceHandler) ListProseBlocksByEntity(ctx context.Context, req *prosepb.ListProseBlocksByEntityRequest) (*prosepb.ListProseBlocksByEntityResponse, error) {
	if req.EntityType == "" {
		return nil, status.Errorf(codes.InvalidArgument, "entity_type is required")
	}

	entityID, err := uuid.Parse(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id: %v", err)
	}

	references, err := h.refRepo.ListByEntity(ctx, story.EntityType(req.EntityType), entityID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list references: %v", err)
	}

	// Get prose blocks for each reference
	protoProseBlocks := make([]*prosepb.ProseBlock, 0, len(references))
	for _, ref := range references {
		proseBlock, err := h.proseBlockRepo.GetByID(ctx, ref.ProseBlockID)
		if err != nil {
			h.logger.Error("failed to get prose block", "prose_block_id", ref.ProseBlockID, "error", err)
			continue
		}
		protoProseBlocks = append(protoProseBlocks, proseBlockToProto(proseBlock))
	}

	return &prosepb.ListProseBlocksByEntityResponse{
		ProseBlocks: protoProseBlocks,
		TotalCount:  int32(len(protoProseBlocks)),
	}, nil
}

// DeleteProseBlockReference deletes a reference
func (h *ProseBlockReferenceHandler) DeleteProseBlockReference(ctx context.Context, req *prosepb.DeleteProseBlockReferenceRequest) (*prosepb.DeleteProseBlockReferenceResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	// Check if reference exists
	_, err = h.refRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "reference not found: %v", err)
	}

	if err := h.refRepo.Delete(ctx, id); err != nil {
		h.logger.Error("failed to delete prose block reference", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to delete reference: %v", err)
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


