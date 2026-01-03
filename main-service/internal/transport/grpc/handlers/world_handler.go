package handlers

import (
	"context"

	"github.com/google/uuid"
	worldapp "github.com/story-engine/main-service/internal/application/world"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	"github.com/story-engine/main-service/internal/platform/logger"
	worldpb "github.com/story-engine/main-service/proto/world"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WorldHandler implements the WorldService gRPC service
type WorldHandler struct {
	worldpb.UnimplementedWorldServiceServer
	createWorldUseCase *worldapp.CreateWorldUseCase
	getWorldUseCase    *worldapp.GetWorldUseCase
	listWorldsUseCase  *worldapp.ListWorldsUseCase
	updateWorldUseCase *worldapp.UpdateWorldUseCase
	deleteWorldUseCase *worldapp.DeleteWorldUseCase
	logger             logger.Logger
}

// NewWorldHandler creates a new WorldHandler
func NewWorldHandler(
	createWorldUseCase *worldapp.CreateWorldUseCase,
	getWorldUseCase *worldapp.GetWorldUseCase,
	listWorldsUseCase *worldapp.ListWorldsUseCase,
	updateWorldUseCase *worldapp.UpdateWorldUseCase,
	deleteWorldUseCase *worldapp.DeleteWorldUseCase,
	logger logger.Logger,
) *WorldHandler {
	return &WorldHandler{
		createWorldUseCase: createWorldUseCase,
		getWorldUseCase:    getWorldUseCase,
		listWorldsUseCase:  listWorldsUseCase,
		updateWorldUseCase: updateWorldUseCase,
		deleteWorldUseCase: deleteWorldUseCase,
		logger:             logger,
	}
}

// CreateWorld creates a new world
func (h *WorldHandler) CreateWorld(ctx context.Context, req *worldpb.CreateWorldRequest) (*worldpb.CreateWorldResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		if req.TenantId != "" {
			tenantID = req.TenantId
		} else {
			return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
		}
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}

	input := worldapp.CreateWorldInput{
		TenantID:    tenantUUID,
		Name:        req.Name,
		Description: req.Description,
		Genre:       req.Genre,
		IsImplicit:  req.IsImplicit,
	}

	output, err := h.createWorldUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &worldpb.CreateWorldResponse{
		World: mappers.WorldToProto(output.World),
	}, nil
}

// GetWorld retrieves a world by ID
func (h *WorldHandler) GetWorld(ctx context.Context, req *worldpb.GetWorldRequest) (*worldpb.GetWorldResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid world id: %v", err)
	}

	output, err := h.getWorldUseCase.Execute(ctx, worldapp.GetWorldInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &worldpb.GetWorldResponse{
		World: mappers.WorldToProto(output.World),
	}, nil
}

// ListWorlds lists worlds for a tenant
func (h *WorldHandler) ListWorlds(ctx context.Context, req *worldpb.ListWorldsRequest) (*worldpb.ListWorldsResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		if req.TenantId != "" {
			tenantID = req.TenantId
		} else {
			return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
		}
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	limit := 50
	offset := 0
	if req.Pagination != nil {
		if req.Pagination.Limit > 0 {
			limit = int(req.Pagination.Limit)
		}
		if req.Pagination.Offset > 0 {
			offset = int(req.Pagination.Offset)
		}
	}

	output, err := h.listWorldsUseCase.Execute(ctx, worldapp.ListWorldsInput{
		TenantID: tenantUUID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}

	worlds := make([]*worldpb.World, len(output.Worlds))
	for i, w := range output.Worlds {
		worlds[i] = mappers.WorldToProto(w)
	}

	return &worldpb.ListWorldsResponse{
		Worlds:     worlds,
		TotalCount: int32(output.Total),
	}, nil
}

// UpdateWorld updates an existing world
func (h *WorldHandler) UpdateWorld(ctx context.Context, req *worldpb.UpdateWorldRequest) (*worldpb.UpdateWorldResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid world id: %v", err)
	}

	var name *string
	if req.Name != nil {
		name = req.Name
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	var genre *string
	if req.Genre != nil {
		genre = req.Genre
	}

	var isImplicit *bool
	if req.IsImplicit != nil {
		isImplicit = req.IsImplicit
	}

	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	input := worldapp.UpdateWorldInput{
		TenantID:    tenantUUID,
		ID:          id,
		Name:        name,
		Description: description,
		Genre:       genre,
		IsImplicit:  isImplicit,
	}

	output, err := h.updateWorldUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &worldpb.UpdateWorldResponse{
		World: mappers.WorldToProto(output.World),
	}, nil
}

// DeleteWorld deletes a world
func (h *WorldHandler) DeleteWorld(ctx context.Context, req *worldpb.DeleteWorldRequest) (*worldpb.DeleteWorldResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid world id: %v", err)
	}

	err = h.deleteWorldUseCase.Execute(ctx, worldapp.DeleteWorldInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &worldpb.DeleteWorldResponse{}, nil
}


