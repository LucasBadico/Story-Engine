package handlers

import (
	"context"

	"github.com/google/uuid"
	plotapp "github.com/story-engine/main-service/internal/application/story/plot"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	pb "github.com/story-engine/main-service/proto/plot"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PlotHandler struct {
	pb.UnimplementedPlotServiceServer
	createUseCase           *plotapp.CreatePlotUseCase
	getUseCase              *plotapp.GetPlotUseCase
	listUseCase             *plotapp.ListPlotsUseCase
	updateUseCase           *plotapp.UpdatePlotUseCase
	deleteUseCase           *plotapp.DeletePlotUseCase
	addEntityUseCase        *plotapp.AddPlotEntityUseCase
	createDevUseCase        *plotapp.CreatePlotDevelopmentUseCase
	listDevsUseCase         *plotapp.ListPlotDevelopmentsUseCase
	logger                  logger.Logger
}

func NewPlotHandler(
	createUseCase *plotapp.CreatePlotUseCase,
	getUseCase *plotapp.GetPlotUseCase,
	listUseCase *plotapp.ListPlotsUseCase,
	updateUseCase *plotapp.UpdatePlotUseCase,
	deleteUseCase *plotapp.DeletePlotUseCase,
	addEntityUseCase *plotapp.AddPlotEntityUseCase,
	createDevUseCase *plotapp.CreatePlotDevelopmentUseCase,
	listDevsUseCase *plotapp.ListPlotDevelopmentsUseCase,
	logger logger.Logger,
) *PlotHandler {
	return &PlotHandler{
		createUseCase:    createUseCase,
		getUseCase:       getUseCase,
		listUseCase:      listUseCase,
		updateUseCase:    updateUseCase,
		deleteUseCase:    deleteUseCase,
		addEntityUseCase: addEntityUseCase,
		createDevUseCase: createDevUseCase,
		listDevsUseCase:  listDevsUseCase,
		logger:           logger,
	}
}

func (h *PlotHandler) CreatePlot(ctx context.Context, req *pb.CreatePlotRequest) (*pb.CreatePlotResponse, error) {
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

	entities := make([]plotapp.EntityLinkInput, len(req.Entities))
	for i, e := range req.Entities {
		entityID, err := uuid.Parse(e.EntityId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id at index %d: %v", i, err)
		}
		entities[i] = plotapp.EntityLinkInput{
			Role:       e.Role,
			EntityType: e.EntityType,
			EntityID:   entityID,
		}
	}

	output, err := h.createUseCase.Execute(ctx, plotapp.CreatePlotInput{
		TenantID:    tenantUUID,
		StoryID:     storyID,
		Title:       req.Title,
		Type:        req.Type,
		Description: req.Description,
		Entities:    entities,
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreatePlotResponse{
		Plot: plotToProto(output.Plot),
	}, nil
}

func (h *PlotHandler) GetPlot(ctx context.Context, req *pb.GetPlotRequest) (*pb.GetPlotResponse, error) {
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

	output, err := h.getUseCase.Execute(ctx, plotapp.GetPlotInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &pb.GetPlotResponse{
		Plot: plotToProto(output.Plot),
	}, nil
}

func (h *PlotHandler) ListPlotsByStory(ctx context.Context, req *pb.ListPlotsByStoryRequest) (*pb.ListPlotsByStoryResponse, error) {
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

	output, err := h.listUseCase.Execute(ctx, plotapp.ListPlotsInput{
		TenantID: tenantUUID,
		StoryID:  storyID,
		Status:   req.Status,
	})
	if err != nil {
		return nil, err
	}

	plots := make([]*pb.Plot, len(output.Plots))
	for i, p := range output.Plots {
		plots[i] = plotToProto(p)
	}

	return &pb.ListPlotsByStoryResponse{
		Plots: plots,
	}, nil
}

func (h *PlotHandler) UpdatePlot(ctx context.Context, req *pb.UpdatePlotRequest) (*pb.UpdatePlotResponse, error) {
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

	input := plotapp.UpdatePlotInput{
		TenantID: tenantUUID,
		ID:       id,
	}
	if req.Title != "" {
		input.Title = &req.Title
	}
	if req.Type != "" {
		input.Type = &req.Type
	}
	if req.Description != "" {
		input.Description = &req.Description
	}
	if req.Status != "" {
		input.Status = &req.Status
	}

	output, err := h.updateUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &pb.UpdatePlotResponse{
		Plot: plotToProto(output.Plot),
	}, nil
}

func (h *PlotHandler) DeletePlot(ctx context.Context, req *pb.DeletePlotRequest) (*pb.DeletePlotResponse, error) {
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

	if err := h.deleteUseCase.Execute(ctx, plotapp.DeletePlotInput{
		TenantID: tenantUUID,
		ID:       id,
	}); err != nil {
		return nil, err
	}

	return &pb.DeletePlotResponse{}, nil
}

func (h *PlotHandler) AddPlotEntity(ctx context.Context, req *pb.AddPlotEntityRequest) (*pb.AddPlotEntityResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	plotID, err := uuid.Parse(req.PlotId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid plot_id: %v", err)
	}

	entityID, err := uuid.Parse(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id: %v", err)
	}

	output, err := h.addEntityUseCase.Execute(ctx, plotapp.AddPlotEntityInput{
		TenantID:   tenantUUID,
		PlotID:     plotID,
		Role:       req.Role,
		EntityType: req.EntityType,
		EntityID:   entityID,
	})
	if err != nil {
		return nil, err
	}

	return &pb.AddPlotEntityResponse{
		Entity: &pb.EntityLink{
			Id:         output.PlotEntity.ID.String(),
			Role:       output.PlotEntity.Role,
			EntityType: output.PlotEntity.EntityType,
			EntityId:   output.PlotEntity.EntityID.String(),
		},
	}, nil
}

func (h *PlotHandler) CreatePlotDevelopment(ctx context.Context, req *pb.CreatePlotDevelopmentRequest) (*pb.CreatePlotDevelopmentResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	plotID, err := uuid.Parse(req.PlotId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid plot_id: %v", err)
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

	var timelineEventID *uuid.UUID
	if req.TimelineEventId != "" {
		id, err := uuid.Parse(req.TimelineEventId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid timeline_event_id: %v", err)
		}
		timelineEventID = &id
	}

	output, err := h.createDevUseCase.Execute(ctx, plotapp.CreatePlotDevelopmentInput{
		TenantID:           tenantUUID,
		PlotID:             plotID,
		ChapterNumber:      int(req.ChapterNumber),
		Action:             req.Action,
		Description:        req.Description,
		Notes:              req.Notes,
		TimelineDay:        timelineDay,
		TimelineConfidence: timelineConfidence,
		TimelineEventID:    timelineEventID,
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreatePlotDevelopmentResponse{
		PlotDevelopment: plotDevelopmentToProto(output.PlotDevelopment),
	}, nil
}

func (h *PlotHandler) ListPlotDevelopments(ctx context.Context, req *pb.ListPlotDevelopmentsRequest) (*pb.ListPlotDevelopmentsResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	plotID, err := uuid.Parse(req.PlotId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid plot_id: %v", err)
	}

	var asOfChapter *int
	if req.AsOfChapter > 0 {
		ch := int(req.AsOfChapter)
		asOfChapter = &ch
	}

	output, err := h.listDevsUseCase.Execute(ctx, plotapp.ListPlotDevelopmentsInput{
		TenantID:    tenantUUID,
		PlotID:      plotID,
		AsOfChapter: asOfChapter,
	})
	if err != nil {
		return nil, err
	}

	devs := make([]*pb.PlotDevelopment, len(output.PlotDevelopments))
	for i, d := range output.PlotDevelopments {
		devs[i] = plotDevelopmentToProto(d)
	}

	return &pb.ListPlotDevelopmentsResponse{
		PlotDevelopments: devs,
	}, nil
}

func plotToProto(p *story.Plot) *pb.Plot {
	proto := &pb.Plot{
		Id:          p.ID.String(),
		StoryId:     p.StoryID.String(),
		Title:       p.Title,
		Type:        p.Type,
		Description: p.Description,
		Status:      p.Status,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}

	entities := make([]*pb.EntityLink, len(p.Entities))
	for i, e := range p.Entities {
		entities[i] = &pb.EntityLink{
			Id:         e.ID.String(),
			Role:       e.Role,
			EntityType: e.EntityType,
			EntityId:   e.EntityID.String(),
		}
	}
	proto.Entities = entities

	return proto
}

func plotDevelopmentToProto(d *story.PlotDevelopment) *pb.PlotDevelopment {
	p := &pb.PlotDevelopment{
		Id:            d.ID.String(),
		PlotId:        d.PlotID.String(),
		ChapterNumber: int32(d.ChapterNumber),
		Action:        d.Action,
		Description:   d.Description,
		Notes:         d.Notes,
		CreatedAt:     timestamppb.New(d.CreatedAt),
	}
	if d.TimelineDay != nil {
		p.HasTimelineDay = true
		p.TimelineDay = int32(*d.TimelineDay)
	}
	if d.TimelineConfidence != nil {
		p.TimelineConfidence = *d.TimelineConfidence
	}
	if d.TimelineEventID != nil {
		p.TimelineEventId = d.TimelineEventID.String()
	}

	return p
}
