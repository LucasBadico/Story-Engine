package handlers

import (
	"context"

	"github.com/google/uuid"
	timelineeventapp "github.com/story-engine/main-service/internal/application/story/timeline_event"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	pb "github.com/story-engine/main-service/proto/timeline_event"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TimelineEventHandler struct {
	pb.UnimplementedTimelineEventServiceServer
	createUseCase        *timelineeventapp.CreateTimelineEventUseCase
	getUseCase           *timelineeventapp.GetTimelineEventUseCase
	listByStoryUseCase   *timelineeventapp.ListTimelineEventsByStoryUseCase
	listByChapterUseCase *timelineeventapp.ListTimelineEventsByChapterUseCase
	deleteUseCase        *timelineeventapp.DeleteTimelineEventUseCase
	logger               logger.Logger
}

func NewTimelineEventHandler(
	createUseCase *timelineeventapp.CreateTimelineEventUseCase,
	getUseCase *timelineeventapp.GetTimelineEventUseCase,
	listByStoryUseCase *timelineeventapp.ListTimelineEventsByStoryUseCase,
	listByChapterUseCase *timelineeventapp.ListTimelineEventsByChapterUseCase,
	deleteUseCase *timelineeventapp.DeleteTimelineEventUseCase,
	logger logger.Logger,
) *TimelineEventHandler {
	return &TimelineEventHandler{
		createUseCase:        createUseCase,
		getUseCase:           getUseCase,
		listByStoryUseCase:   listByStoryUseCase,
		listByChapterUseCase: listByChapterUseCase,
		deleteUseCase:        deleteUseCase,
		logger:               logger,
	}
}

func (h *TimelineEventHandler) CreateTimelineEvent(ctx context.Context, req *pb.CreateTimelineEventRequest) (*pb.CreateTimelineEventResponse, error) {
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

	entities := make([]timelineeventapp.EntityLinkInput, len(req.Entities))
	for i, e := range req.Entities {
		entityID, err := uuid.Parse(e.EntityId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id at index %d: %v", i, err)
		}
		entities[i] = timelineeventapp.EntityLinkInput{
			Role:       e.Role,
			EntityType: e.EntityType,
			EntityID:   entityID,
		}
	}

	output, err := h.createUseCase.Execute(ctx, timelineeventapp.CreateTimelineEventInput{
		TenantID:           tenantUUID,
		StoryID:            storyID,
		ChapterID:          chapterID,
		ChapterNumber:      int(req.ChapterNumber),
		TimelineDay:        timelineDay,
		TimelineConfidence: timelineConfidence,
		Description:        req.Description,
		Entities:           entities,
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreateTimelineEventResponse{
		TimelineEvent: timelineEventToProto(output.TimelineEvent),
	}, nil
}

func (h *TimelineEventHandler) GetTimelineEvent(ctx context.Context, req *pb.GetTimelineEventRequest) (*pb.GetTimelineEventResponse, error) {
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

	output, err := h.getUseCase.Execute(ctx, timelineeventapp.GetTimelineEventInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &pb.GetTimelineEventResponse{
		TimelineEvent: timelineEventToProto(output.TimelineEvent),
	}, nil
}

func (h *TimelineEventHandler) ListTimelineEventsByStory(ctx context.Context, req *pb.ListTimelineEventsByStoryRequest) (*pb.ListTimelineEventsByStoryResponse, error) {
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

	output, err := h.listByStoryUseCase.Execute(ctx, timelineeventapp.ListTimelineEventsByStoryInput{
		TenantID:    tenantUUID,
		StoryID:     storyID,
		AsOfChapter: int(req.AsOfChapter),
		Limit:       int(req.Limit),
	})
	if err != nil {
		return nil, err
	}

	events := make([]*pb.TimelineEvent, len(output.TimelineEvents))
	for i, e := range output.TimelineEvents {
		events[i] = timelineEventToProto(e)
	}

	return &pb.ListTimelineEventsByStoryResponse{
		TimelineEvents: events,
	}, nil
}

func (h *TimelineEventHandler) ListTimelineEventsByChapter(ctx context.Context, req *pb.ListTimelineEventsByChapterRequest) (*pb.ListTimelineEventsByChapterResponse, error) {
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

	output, err := h.listByChapterUseCase.Execute(ctx, timelineeventapp.ListTimelineEventsByChapterInput{
		TenantID:  tenantUUID,
		ChapterID: chapterID,
	})
	if err != nil {
		return nil, err
	}

	events := make([]*pb.TimelineEvent, len(output.TimelineEvents))
	for i, e := range output.TimelineEvents {
		events[i] = timelineEventToProto(e)
	}

	return &pb.ListTimelineEventsByChapterResponse{
		TimelineEvents: events,
	}, nil
}

func (h *TimelineEventHandler) DeleteTimelineEvent(ctx context.Context, req *pb.DeleteTimelineEventRequest) (*pb.DeleteTimelineEventResponse, error) {
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

	if err := h.deleteUseCase.Execute(ctx, timelineeventapp.DeleteTimelineEventInput{
		TenantID: tenantUUID,
		ID:       id,
	}); err != nil {
		return nil, err
	}

	return &pb.DeleteTimelineEventResponse{}, nil
}

func timelineEventToProto(te *story.TimelineEvent) *pb.TimelineEvent {
	p := &pb.TimelineEvent{
		Id:            te.ID.String(),
		StoryId:       te.StoryID.String(),
		ChapterId:     te.ChapterID.String(),
		ChapterNumber: int32(te.ChapterNumber),
		Description:   te.Description,
		CreatedAt:     timestamppb.New(te.CreatedAt),
	}
	if te.TimelineDay != nil {
		p.HasTimelineDay = true
		p.TimelineDay = int32(*te.TimelineDay)
	}
	if te.TimelineConfidence != nil {
		p.TimelineConfidence = *te.TimelineConfidence
	}

	entities := make([]*pb.EntityLink, len(te.Entities))
	for i, e := range te.Entities {
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
