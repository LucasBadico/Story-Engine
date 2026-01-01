package handlers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
	beatpb "github.com/story-engine/main-service/proto/beat"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// BeatHandler implements the BeatService gRPC service
type BeatHandler struct {
	beatpb.UnimplementedBeatServiceServer
	beatRepo  repositories.BeatRepository
	sceneRepo repositories.SceneRepository
	storyRepo repositories.StoryRepository
	logger    logger.Logger
}

// NewBeatHandler creates a new BeatHandler
func NewBeatHandler(
	beatRepo repositories.BeatRepository,
	sceneRepo repositories.SceneRepository,
	storyRepo repositories.StoryRepository,
	logger logger.Logger,
) *BeatHandler {
	return &BeatHandler{
		beatRepo:  beatRepo,
		sceneRepo: sceneRepo,
		storyRepo: storyRepo,
		logger:    logger,
	}
}

// CreateBeat creates a new beat
func (h *BeatHandler) CreateBeat(ctx context.Context, req *beatpb.CreateBeatRequest) (*beatpb.CreateBeatResponse, error) {
	sceneID, err := uuid.Parse(req.SceneId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scene_id: %v", err)
	}

	// Validate scene exists
	_, err = h.sceneRepo.GetByID(ctx, sceneID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "scene not found: %v", err)
	}

	if req.OrderNum < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "order_num must be greater than 0")
	}

	if req.Type == "" {
		return nil, status.Errorf(codes.InvalidArgument, "type is required")
	}

	beat, err := story.NewBeat(sceneID, int(req.OrderNum), story.BeatType(req.Type))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid beat: %v", err)
	}

	if req.Intent != "" {
		beat.UpdateIntent(req.Intent)
	}
	if req.Outcome != "" {
		beat.UpdateOutcome(req.Outcome)
	}

	if err := h.beatRepo.Create(ctx, beat); err != nil {
		h.logger.Error("failed to create beat", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create beat: %v", err)
	}

	return &beatpb.CreateBeatResponse{
		Beat: beatToProto(beat),
	}, nil
}

// GetBeat retrieves a beat by ID
func (h *BeatHandler) GetBeat(ctx context.Context, req *beatpb.GetBeatRequest) (*beatpb.GetBeatResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	beat, err := h.beatRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "beat not found: %v", err)
	}

	return &beatpb.GetBeatResponse{
		Beat: beatToProto(beat),
	}, nil
}

// UpdateBeat updates an existing beat
func (h *BeatHandler) UpdateBeat(ctx context.Context, req *beatpb.UpdateBeatRequest) (*beatpb.UpdateBeatResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	beat, err := h.beatRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "beat not found: %v", err)
	}

	// Update fields if provided
	if req.OrderNum != nil {
		if *req.OrderNum < 1 {
			return nil, status.Errorf(codes.InvalidArgument, "order_num must be greater than 0")
		}
		beat.OrderNum = int(*req.OrderNum)
	}

	if req.Type != nil {
		beat.Type = story.BeatType(*req.Type)
		if err := beat.Validate(); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid beat type: %v", err)
		}
	}

	if req.Intent != nil {
		beat.UpdateIntent(*req.Intent)
	}

	if req.Outcome != nil {
		beat.UpdateOutcome(*req.Outcome)
	}

	if err := h.beatRepo.Update(ctx, beat); err != nil {
		h.logger.Error("failed to update beat", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to update beat: %v", err)
	}

	return &beatpb.UpdateBeatResponse{
		Beat: beatToProto(beat),
	}, nil
}

// DeleteBeat deletes a beat
func (h *BeatHandler) DeleteBeat(ctx context.Context, req *beatpb.DeleteBeatRequest) (*beatpb.DeleteBeatResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	// Check if beat exists
	_, err = h.beatRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "beat not found: %v", err)
	}

	if err := h.beatRepo.Delete(ctx, id); err != nil {
		h.logger.Error("failed to delete beat", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to delete beat: %v", err)
	}

	return &beatpb.DeleteBeatResponse{
		Success: true,
	}, nil
}

// MoveBeat moves a beat to a different scene
func (h *BeatHandler) MoveBeat(ctx context.Context, req *beatpb.MoveBeatRequest) (*beatpb.MoveBeatResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	beat, err := h.beatRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "beat not found: %v", err)
	}

	sceneID, err := uuid.Parse(req.SceneId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scene_id: %v", err)
	}

	// Validate scene exists
	_, err = h.sceneRepo.GetByID(ctx, sceneID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "scene not found: %v", err)
	}

	beat.SceneID = sceneID
	beat.UpdatedAt = time.Now()

	if err := h.beatRepo.Update(ctx, beat); err != nil {
		h.logger.Error("failed to move beat", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to move beat: %v", err)
	}

	return &beatpb.MoveBeatResponse{
		Beat: beatToProto(beat),
	}, nil
}

// ListBeatsByScene lists beats for a specific scene
func (h *BeatHandler) ListBeatsByScene(ctx context.Context, req *beatpb.ListBeatsBySceneRequest) (*beatpb.ListBeatsBySceneResponse, error) {
	sceneID, err := uuid.Parse(req.SceneId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scene_id: %v", err)
	}

	beats, err := h.beatRepo.ListByScene(ctx, sceneID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list beats: %v", err)
	}

	protoBeats := make([]*beatpb.Beat, len(beats))
	for i, b := range beats {
		protoBeats[i] = beatToProto(b)
	}

	return &beatpb.ListBeatsBySceneResponse{
		Beats:      protoBeats,
		TotalCount: int32(len(beats)),
	}, nil
}

// ListBeatsByStory lists all beats for a story
func (h *BeatHandler) ListBeatsByStory(ctx context.Context, req *beatpb.ListBeatsByStoryRequest) (*beatpb.ListBeatsByStoryResponse, error) {
	storyID, err := uuid.Parse(req.StoryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid story_id: %v", err)
	}

	beats, err := h.beatRepo.ListByStory(ctx, storyID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list beats: %v", err)
	}

	protoBeats := make([]*beatpb.Beat, len(beats))
	for i, b := range beats {
		protoBeats[i] = beatToProto(b)
	}

	return &beatpb.ListBeatsByStoryResponse{
		Beats:      protoBeats,
		TotalCount: int32(len(beats)),
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

