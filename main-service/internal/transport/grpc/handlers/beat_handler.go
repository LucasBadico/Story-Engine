package handlers

import (
	"context"

	"github.com/google/uuid"
	beatapp "github.com/story-engine/main-service/internal/application/story/beat"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	beatpb "github.com/story-engine/main-service/proto/beat"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// BeatHandler implements the BeatService gRPC service
type BeatHandler struct {
	beatpb.UnimplementedBeatServiceServer
	createBeatUseCase *beatapp.CreateBeatUseCase
	getBeatUseCase    *beatapp.GetBeatUseCase
	updateBeatUseCase *beatapp.UpdateBeatUseCase
	deleteBeatUseCase *beatapp.DeleteBeatUseCase
	listBeatsUseCase  *beatapp.ListBeatsUseCase
	moveBeatUseCase   *beatapp.MoveBeatUseCase
	logger            logger.Logger
}

// NewBeatHandler creates a new BeatHandler
func NewBeatHandler(
	createBeatUseCase *beatapp.CreateBeatUseCase,
	getBeatUseCase *beatapp.GetBeatUseCase,
	updateBeatUseCase *beatapp.UpdateBeatUseCase,
	deleteBeatUseCase *beatapp.DeleteBeatUseCase,
	listBeatsUseCase *beatapp.ListBeatsUseCase,
	moveBeatUseCase *beatapp.MoveBeatUseCase,
	logger logger.Logger,
) *BeatHandler {
	return &BeatHandler{
		createBeatUseCase: createBeatUseCase,
		getBeatUseCase:    getBeatUseCase,
		updateBeatUseCase: updateBeatUseCase,
		deleteBeatUseCase: deleteBeatUseCase,
		listBeatsUseCase:  listBeatsUseCase,
		moveBeatUseCase:   moveBeatUseCase,
		logger:            logger,
	}
}

// CreateBeat creates a new beat
func (h *BeatHandler) CreateBeat(ctx context.Context, req *beatpb.CreateBeatRequest) (*beatpb.CreateBeatResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	sceneID, err := uuid.Parse(req.SceneId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scene_id: %v", err)
	}

	output, err := h.createBeatUseCase.Execute(ctx, beatapp.CreateBeatInput{
		TenantID: tenantUUID,
		SceneID:  sceneID,
		OrderNum: int(req.OrderNum),
		Type:     story.BeatType(req.Type),
		Intent:   req.Intent,
		Outcome:  req.Outcome,
	})
	if err != nil {
		return nil, err
	}

	return &beatpb.CreateBeatResponse{
		Beat: beatToProto(output.Beat),
	}, nil
}

// GetBeat retrieves a beat by ID
func (h *BeatHandler) GetBeat(ctx context.Context, req *beatpb.GetBeatRequest) (*beatpb.GetBeatResponse, error) {
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

	output, err := h.getBeatUseCase.Execute(ctx, beatapp.GetBeatInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &beatpb.GetBeatResponse{
		Beat: beatToProto(output.Beat),
	}, nil
}

// UpdateBeat updates an existing beat
func (h *BeatHandler) UpdateBeat(ctx context.Context, req *beatpb.UpdateBeatRequest) (*beatpb.UpdateBeatResponse, error) {
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
		n := int(*req.OrderNum)
		orderNum = &n
	}

	input := beatapp.UpdateBeatInput{
		TenantID: tenantUUID,
		ID:       id,
		OrderNum: orderNum,
		Intent:   req.Intent,
		Outcome:  req.Outcome,
	}
	if req.Type != nil {
		beatType := story.BeatType(*req.Type)
		input.Type = &beatType
	}

	output, err := h.updateBeatUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &beatpb.UpdateBeatResponse{
		Beat: beatToProto(output.Beat),
	}, nil
}

// DeleteBeat deletes a beat
func (h *BeatHandler) DeleteBeat(ctx context.Context, req *beatpb.DeleteBeatRequest) (*beatpb.DeleteBeatResponse, error) {
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

	if err := h.deleteBeatUseCase.Execute(ctx, beatapp.DeleteBeatInput{
		TenantID: tenantUUID,
		ID:       id,
	}); err != nil {
		return nil, err
	}

	return &beatpb.DeleteBeatResponse{
		Success: true,
	}, nil
}

// MoveBeat moves a beat to a different scene
func (h *BeatHandler) MoveBeat(ctx context.Context, req *beatpb.MoveBeatRequest) (*beatpb.MoveBeatResponse, error) {
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

	newSceneID, err := uuid.Parse(req.SceneId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scene_id: %v", err)
	}

	output, err := h.moveBeatUseCase.Execute(ctx, beatapp.MoveBeatInput{
		TenantID:   tenantUUID,
		BeatID:     id,
		NewSceneID: newSceneID,
	})
	if err != nil {
		return nil, err
	}

	return &beatpb.MoveBeatResponse{
		Beat: beatToProto(output.Beat),
	}, nil
}

// ListBeatsByScene lists beats for a specific scene
func (h *BeatHandler) ListBeatsByScene(ctx context.Context, req *beatpb.ListBeatsBySceneRequest) (*beatpb.ListBeatsBySceneResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	sceneID, err := uuid.Parse(req.SceneId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scene_id: %v", err)
	}

	output, err := h.listBeatsUseCase.Execute(ctx, beatapp.ListBeatsInput{
		TenantID: tenantUUID,
		SceneID:  sceneID,
	})
	if err != nil {
		return nil, err
	}

	protoBeats := make([]*beatpb.Beat, len(output.Beats))
	for i, b := range output.Beats {
		protoBeats[i] = beatToProto(b)
	}

	return &beatpb.ListBeatsBySceneResponse{
		Beats:      protoBeats,
		TotalCount: int32(output.Total),
	}, nil
}

// ListBeatsByStory lists all beats for a story
func (h *BeatHandler) ListBeatsByStory(ctx context.Context, req *beatpb.ListBeatsByStoryRequest) (*beatpb.ListBeatsByStoryResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	storyID, err := uuid.Parse(req.StoryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid story_id: %v", err)
	}

	output, err := h.listBeatsUseCase.Execute(ctx, beatapp.ListBeatsInput{
		TenantID: tenantUUID,
		StoryID:  &storyID,
	})
	if err != nil {
		return nil, err
	}

	protoBeats := make([]*beatpb.Beat, len(output.Beats))
	for i, b := range output.Beats {
		protoBeats[i] = beatToProto(b)
	}

	return &beatpb.ListBeatsByStoryResponse{
		Beats:      protoBeats,
		TotalCount: int32(output.Total),
	}, nil
}

// beatToProto converts a domain Beat to a proto Beat message
func beatToProto(b *story.Beat) *beatpb.Beat {
	return &beatpb.Beat{
		Id:        b.ID.String(),
		SceneId:   b.SceneID.String(),
		OrderNum:  int32(b.OrderNum),
		Type:      string(b.Type),
		Intent:    b.Intent,
		Outcome:   b.Outcome,
		CreatedAt: timestamppb.New(b.CreatedAt),
		UpdatedAt: timestamppb.New(b.UpdatedAt),
	}
}


