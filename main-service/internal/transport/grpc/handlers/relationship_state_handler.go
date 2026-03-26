package handlers

import (
	"context"

	"github.com/google/uuid"
	relationshipstateapp "github.com/story-engine/main-service/internal/application/world/relationship_state"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	pb "github.com/story-engine/main-service/proto/relationship_state"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RelationshipStateHandler struct {
	pb.UnimplementedRelationshipStateServiceServer
	createUseCase    *relationshipstateapp.CreateRelationshipStateUseCase
	getUseCase       *relationshipstateapp.GetRelationshipStateUseCase
	listUseCase      *relationshipstateapp.ListRelationshipStatesUseCase
	deleteUseCase    *relationshipstateapp.DeleteRelationshipStateUseCase
	logger           logger.Logger
}

func NewRelationshipStateHandler(
	createUseCase *relationshipstateapp.CreateRelationshipStateUseCase,
	getUseCase *relationshipstateapp.GetRelationshipStateUseCase,
	listUseCase *relationshipstateapp.ListRelationshipStatesUseCase,
	deleteUseCase *relationshipstateapp.DeleteRelationshipStateUseCase,
	logger logger.Logger,
) *RelationshipStateHandler {
	return &RelationshipStateHandler{
		createUseCase: createUseCase,
		getUseCase:    getUseCase,
		listUseCase:   listUseCase,
		deleteUseCase: deleteUseCase,
		logger:        logger,
	}
}

func (h *RelationshipStateHandler) CreateRelationshipState(ctx context.Context, req *pb.CreateRelationshipStateRequest) (*pb.CreateRelationshipStateResponse, error) {
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

	var timelineDay *int
	if req.HasTimelineDay {
		day := int(req.TimelineDay)
		timelineDay = &day
	}

	var timelineConfidence *string
	if req.TimelineConfidence != "" {
		timelineConfidence = &req.TimelineConfidence
	}

	entities := make([]relationshipstateapp.EntityLinkInput, len(req.Entities))
	for i, e := range req.Entities {
		entityID, err := uuid.Parse(e.EntityId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id at index %d: %v", i, err)
		}
		entities[i] = relationshipstateapp.EntityLinkInput{
			Role:       e.Role,
			EntityType: e.EntityType,
			EntityID:   entityID,
		}
	}

	output, err := h.createUseCase.Execute(ctx, relationshipstateapp.CreateRelationshipStateInput{
		TenantID:           tenantUUID,
		ChapterID:          chapterID,
		ChapterNumber:      int(req.ChapterNumber),
		TensionDescription: req.TensionDescription,
		TimelineDay:        timelineDay,
		TimelineConfidence: timelineConfidence,
		Entities:           entities,
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreateRelationshipStateResponse{
		RelationshipState: relationshipStateToProto(output.RelationshipState),
	}, nil
}

func (h *RelationshipStateHandler) GetRelationshipState(ctx context.Context, req *pb.GetRelationshipStateRequest) (*pb.GetRelationshipStateResponse, error) {
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

	output, err := h.getUseCase.Execute(ctx, relationshipstateapp.GetRelationshipStateInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &pb.GetRelationshipStateResponse{
		RelationshipState: relationshipStateToProto(output.RelationshipState),
	}, nil
}

func (h *RelationshipStateHandler) ListRelationshipStatesByChapter(ctx context.Context, req *pb.ListRelationshipStatesByChapterRequest) (*pb.ListRelationshipStatesByChapterResponse, error) {
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

	output, err := h.listUseCase.Execute(ctx, relationshipstateapp.ListRelationshipStatesInput{
		TenantID:  tenantUUID,
		ChapterID: chapterID,
	})
	if err != nil {
		return nil, err
	}

	states := make([]*pb.RelationshipState, len(output.RelationshipStates))
	for i, s := range output.RelationshipStates {
		states[i] = relationshipStateToProto(s)
	}

	return &pb.ListRelationshipStatesByChapterResponse{
		RelationshipStates: states,
	}, nil
}

func (h *RelationshipStateHandler) DeleteRelationshipState(ctx context.Context, req *pb.DeleteRelationshipStateRequest) (*pb.DeleteRelationshipStateResponse, error) {
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

	if err := h.deleteUseCase.Execute(ctx, relationshipstateapp.DeleteRelationshipStateInput{
		TenantID: tenantUUID,
		ID:       id,
	}); err != nil {
		return nil, err
	}

	return &pb.DeleteRelationshipStateResponse{}, nil
}

func relationshipStateToProto(rs *world.RelationshipState) *pb.RelationshipState {
	p := &pb.RelationshipState{
		Id:                 rs.ID.String(),
		ChapterId:          rs.ChapterID.String(),
		ChapterNumber:      int32(rs.ChapterNumber),
		TensionDescription: rs.TensionDescription,
		CreatedAt:          timestamppb.New(rs.CreatedAt),
		UpdatedAt:          timestamppb.New(rs.UpdatedAt),
	}
	if rs.TimelineDay != nil {
		p.HasTimelineDay = true
		p.TimelineDay = int32(*rs.TimelineDay)
	}
	if rs.TimelineConfidence != nil {
		p.TimelineConfidence = *rs.TimelineConfidence
	}

	entities := make([]*pb.EntityLink, len(rs.Entities))
	for i, e := range rs.Entities {
		entities[i] = &pb.EntityLink{
			Id:         e.ID.String(),
			Role:       e.Role,
			EntityType: e.EntityType,
			EntityId:   e.EntityID.String(),
		}
	}
	p.Entities = entities

	return p
}
