package handlers

import (
	"context"

	"github.com/google/uuid"
	characterstateapp "github.com/story-engine/main-service/internal/application/world/character_state"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	pb "github.com/story-engine/main-service/proto/character_state"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CharacterStateHandler struct {
	pb.UnimplementedCharacterStateServiceServer
	createUseCase    *characterstateapp.CreateCharacterStateUseCase
	getUseCase       *characterstateapp.GetCharacterStateUseCase
	getLatestUseCase *characterstateapp.GetLatestCharacterStateUseCase
	listUseCase      *characterstateapp.ListCharacterStatesUseCase
	deleteUseCase    *characterstateapp.DeleteCharacterStateUseCase
	logger           logger.Logger
}

func NewCharacterStateHandler(
	createUseCase *characterstateapp.CreateCharacterStateUseCase,
	getUseCase *characterstateapp.GetCharacterStateUseCase,
	getLatestUseCase *characterstateapp.GetLatestCharacterStateUseCase,
	listUseCase *characterstateapp.ListCharacterStatesUseCase,
	deleteUseCase *characterstateapp.DeleteCharacterStateUseCase,
	logger logger.Logger,
) *CharacterStateHandler {
	return &CharacterStateHandler{
		createUseCase:    createUseCase,
		getUseCase:       getUseCase,
		getLatestUseCase: getLatestUseCase,
		listUseCase:      listUseCase,
		deleteUseCase:    deleteUseCase,
		logger:           logger,
	}
}

func (h *CharacterStateHandler) CreateCharacterState(ctx context.Context, req *pb.CreateCharacterStateRequest) (*pb.CreateCharacterStateResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	var chapterID *uuid.UUID
	if req.ChapterId != "" {
		id, err := uuid.Parse(req.ChapterId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid chapter_id: %v", err)
		}
		chapterID = &id
	}

	var timelineDay *int
	if req.HasTimelineDay {
		day := int(req.TimelineDay)
		timelineDay = &day
	}

	var timelineConfidence *string
	if req.TimelineConfidence != "" {
		timelineConfidence = &req.TimelineConfidence
	}

	output, err := h.createUseCase.Execute(ctx, characterstateapp.CreateCharacterStateInput{
		TenantID:           tenantUUID,
		CharacterID:        characterID,
		ChapterID:          chapterID,
		ChapterNumber:      int(req.ChapterNumber),
		Emotion:            req.Emotion,
		Goal:               req.Goal,
		Secrets:            req.Secrets,
		Clothes:            req.Clothes,
		PhysicalState:      req.PhysicalState,
		TimelineDay:        timelineDay,
		TimelineConfidence: timelineConfidence,
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreateCharacterStateResponse{
		CharacterState: characterStateToProto(output.CharacterState),
	}, nil
}

func (h *CharacterStateHandler) GetCharacterState(ctx context.Context, req *pb.GetCharacterStateRequest) (*pb.GetCharacterStateResponse, error) {
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

	output, err := h.getUseCase.Execute(ctx, characterstateapp.GetCharacterStateInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &pb.GetCharacterStateResponse{
		CharacterState: characterStateToProto(output.CharacterState),
	}, nil
}

func (h *CharacterStateHandler) GetLatestCharacterState(ctx context.Context, req *pb.GetLatestCharacterStateRequest) (*pb.GetLatestCharacterStateResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	output, err := h.getLatestUseCase.Execute(ctx, characterstateapp.GetLatestCharacterStateInput{
		TenantID:    tenantUUID,
		CharacterID: characterID,
		AsOfChapter: int(req.AsOfChapter),
	})
	if err != nil {
		return nil, err
	}

	return &pb.GetLatestCharacterStateResponse{
		CharacterState: characterStateToProto(output.CharacterState),
	}, nil
}

func (h *CharacterStateHandler) ListCharacterStatesByChapter(ctx context.Context, req *pb.ListCharacterStatesByChapterRequest) (*pb.ListCharacterStatesByChapterResponse, error) {
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

	output, err := h.listUseCase.Execute(ctx, characterstateapp.ListCharacterStatesInput{
		TenantID:  tenantUUID,
		ChapterID: chapterID,
	})
	if err != nil {
		return nil, err
	}

	states := make([]*pb.CharacterState, len(output.CharacterStates))
	for i, s := range output.CharacterStates {
		states[i] = characterStateToProto(s)
	}

	return &pb.ListCharacterStatesByChapterResponse{
		CharacterStates: states,
	}, nil
}

func (h *CharacterStateHandler) DeleteCharacterState(ctx context.Context, req *pb.DeleteCharacterStateRequest) (*pb.DeleteCharacterStateResponse, error) {
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

	if err := h.deleteUseCase.Execute(ctx, characterstateapp.DeleteCharacterStateInput{
		TenantID: tenantUUID,
		ID:       id,
	}); err != nil {
		return nil, err
	}

	return &pb.DeleteCharacterStateResponse{}, nil
}

func characterStateToProto(cs *world.CharacterState) *pb.CharacterState {
	p := &pb.CharacterState{
		Id:            cs.ID.String(),
		CharacterId:   cs.CharacterID.String(),
		ChapterNumber: int32(cs.ChapterNumber),
		Emotion:       cs.Emotion,
		Goal:          cs.Goal,
		Secrets:       cs.Secrets,
		Clothes:       cs.Clothes,
		PhysicalState: cs.PhysicalState,
		CreatedAt:     timestamppb.New(cs.CreatedAt),
		UpdatedAt:     timestamppb.New(cs.UpdatedAt),
	}
	if cs.ChapterID != nil {
		p.ChapterId = cs.ChapterID.String()
	}
	if cs.TimelineDay != nil {
		p.HasTimelineDay = true
		p.TimelineDay = int32(*cs.TimelineDay)
	}
	if cs.TimelineConfidence != nil {
		p.TimelineConfidence = *cs.TimelineConfidence
	}
	return p
}
