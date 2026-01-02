package handlers

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	rpgsystemapp "github.com/story-engine/main-service/internal/application/rpg/rpg_system"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	"github.com/story-engine/main-service/internal/platform/logger"
	rpgsystempb "github.com/story-engine/main-service/proto/rpg_system"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RPGSystemHandler implements the RPGSystemService gRPC service
type RPGSystemHandler struct {
	rpgsystempb.UnimplementedRPGSystemServiceServer
	createRPGSystemUseCase *rpgsystemapp.CreateRPGSystemUseCase
	getRPGSystemUseCase    *rpgsystemapp.GetRPGSystemUseCase
	listRPGSystemsUseCase  *rpgsystemapp.ListRPGSystemsUseCase
	updateRPGSystemUseCase *rpgsystemapp.UpdateRPGSystemUseCase
	deleteRPGSystemUseCase *rpgsystemapp.DeleteRPGSystemUseCase
	logger                 logger.Logger
}

// NewRPGSystemHandler creates a new RPGSystemHandler
func NewRPGSystemHandler(
	createRPGSystemUseCase *rpgsystemapp.CreateRPGSystemUseCase,
	getRPGSystemUseCase *rpgsystemapp.GetRPGSystemUseCase,
	listRPGSystemsUseCase *rpgsystemapp.ListRPGSystemsUseCase,
	updateRPGSystemUseCase *rpgsystemapp.UpdateRPGSystemUseCase,
	deleteRPGSystemUseCase *rpgsystemapp.DeleteRPGSystemUseCase,
	logger logger.Logger,
) *RPGSystemHandler {
	return &RPGSystemHandler{
		createRPGSystemUseCase: createRPGSystemUseCase,
		getRPGSystemUseCase:    getRPGSystemUseCase,
		listRPGSystemsUseCase:  listRPGSystemsUseCase,
		updateRPGSystemUseCase: updateRPGSystemUseCase,
		deleteRPGSystemUseCase: deleteRPGSystemUseCase,
		logger:                 logger,
	}
}

// CreateRPGSystem creates a new RPG system
func (h *RPGSystemHandler) CreateRPGSystem(ctx context.Context, req *rpgsystempb.CreateRPGSystemRequest) (*rpgsystempb.CreateRPGSystemResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	var tenantUUID *uuid.UUID
	if ok {
		parsedTenantID, err := uuid.Parse(tenantID)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
		}
		tenantUUID = &parsedTenantID
	} else if req.TenantId != nil && *req.TenantId != "" {
		parsedTenantID, err := uuid.Parse(*req.TenantId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
		}
		tenantUUID = &parsedTenantID
	}

	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}

	if req.BaseStatsSchema == "" {
		return nil, status.Errorf(codes.InvalidArgument, "base_stats_schema is required")
	}

	baseStatsSchema := json.RawMessage(req.BaseStatsSchema)

	var derivedStatsSchema *json.RawMessage
	if req.DerivedStatsSchema != nil && *req.DerivedStatsSchema != "" {
		schema := json.RawMessage(*req.DerivedStatsSchema)
		derivedStatsSchema = &schema
	}

	var progressionSchema *json.RawMessage
	if req.ProgressionSchema != nil && *req.ProgressionSchema != "" {
		schema := json.RawMessage(*req.ProgressionSchema)
		progressionSchema = &schema
	}

	input := rpgsystemapp.CreateRPGSystemInput{
		TenantID:          tenantUUID,
		Name:              req.Name,
		Description:       req.Description,
		BaseStatsSchema:   baseStatsSchema,
		DerivedStatsSchema: derivedStatsSchema,
		ProgressionSchema: progressionSchema,
	}

	output, err := h.createRPGSystemUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &rpgsystempb.CreateRPGSystemResponse{
		RpgSystem: mappers.RPGSystemToProto(output.RPGSystem),
	}, nil
}

// GetRPGSystem retrieves an RPG system by ID
func (h *RPGSystemHandler) GetRPGSystem(ctx context.Context, req *rpgsystempb.GetRPGSystemRequest) (*rpgsystempb.GetRPGSystemResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid rpg_system id: %v", err)
	}

	output, err := h.getRPGSystemUseCase.Execute(ctx, rpgsystemapp.GetRPGSystemInput{
		ID: id,
	})
	if err != nil {
		return nil, err
	}

	return &rpgsystempb.GetRPGSystemResponse{
		RpgSystem: mappers.RPGSystemToProto(output.RPGSystem),
	}, nil
}

// ListRPGSystems lists RPG systems
func (h *RPGSystemHandler) ListRPGSystems(ctx context.Context, req *rpgsystempb.ListRPGSystemsRequest) (*rpgsystempb.ListRPGSystemsResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	var tenantUUID *uuid.UUID
	if ok {
		parsedTenantID, err := uuid.Parse(tenantID)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
		}
		tenantUUID = &parsedTenantID
	} else if req.TenantId != nil && *req.TenantId != "" {
		parsedTenantID, err := uuid.Parse(*req.TenantId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
		}
		tenantUUID = &parsedTenantID
	}

	output, err := h.listRPGSystemsUseCase.Execute(ctx, rpgsystemapp.ListRPGSystemsInput{
		TenantID: tenantUUID,
	})
	if err != nil {
		return nil, err
	}

	rpgSystems := make([]*rpgsystempb.RPGSystem, len(output.RPGSystems))
	for i, s := range output.RPGSystems {
		rpgSystems[i] = mappers.RPGSystemToProto(s)
	}

	return &rpgsystempb.ListRPGSystemsResponse{
		RpgSystems: rpgSystems,
		TotalCount: int32(len(output.RPGSystems)),
	}, nil
}

// UpdateRPGSystem updates an existing RPG system
func (h *RPGSystemHandler) UpdateRPGSystem(ctx context.Context, req *rpgsystempb.UpdateRPGSystemRequest) (*rpgsystempb.UpdateRPGSystemResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid rpg_system id: %v", err)
	}

	var name *string
	if req.Name != nil {
		name = req.Name
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	var baseStatsSchema *json.RawMessage
	if req.BaseStatsSchema != nil && *req.BaseStatsSchema != "" {
		schema := json.RawMessage(*req.BaseStatsSchema)
		baseStatsSchema = &schema
	}

	var derivedStatsSchema *json.RawMessage
	if req.DerivedStatsSchema != nil && *req.DerivedStatsSchema != "" {
		schema := json.RawMessage(*req.DerivedStatsSchema)
		derivedStatsSchema = &schema
	}

	var progressionSchema *json.RawMessage
	if req.ProgressionSchema != nil && *req.ProgressionSchema != "" {
		schema := json.RawMessage(*req.ProgressionSchema)
		progressionSchema = &schema
	}

	input := rpgsystemapp.UpdateRPGSystemInput{
		ID:                id,
		Name:              name,
		Description:       description,
		BaseStatsSchema:   baseStatsSchema,
		DerivedStatsSchema: derivedStatsSchema,
		ProgressionSchema: progressionSchema,
	}

	output, err := h.updateRPGSystemUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &rpgsystempb.UpdateRPGSystemResponse{
		RpgSystem: mappers.RPGSystemToProto(output.RPGSystem),
	}, nil
}

// DeleteRPGSystem deletes an RPG system
func (h *RPGSystemHandler) DeleteRPGSystem(ctx context.Context, req *rpgsystempb.DeleteRPGSystemRequest) (*rpgsystempb.DeleteRPGSystemResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid rpg_system id: %v", err)
	}

	err = h.deleteRPGSystemUseCase.Execute(ctx, rpgsystemapp.DeleteRPGSystemInput{
		ID: id,
	})
	if err != nil {
		return nil, err
	}

	return &rpgsystempb.DeleteRPGSystemResponse{}, nil
}

