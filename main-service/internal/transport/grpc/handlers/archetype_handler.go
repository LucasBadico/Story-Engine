package handlers

import (
	"context"

	"github.com/google/uuid"
	archetypeapp "github.com/story-engine/main-service/internal/application/world/archetype"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	archetypepb "github.com/story-engine/main-service/proto/archetype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ArchetypeHandler implements the ArchetypeService gRPC service
type ArchetypeHandler struct {
	archetypepb.UnimplementedArchetypeServiceServer
	createArchetypeUseCase     *archetypeapp.CreateArchetypeUseCase
	getArchetypeUseCase        *archetypeapp.GetArchetypeUseCase
	listArchetypesUseCase      *archetypeapp.ListArchetypesUseCase
	updateArchetypeUseCase     *archetypeapp.UpdateArchetypeUseCase
	deleteArchetypeUseCase     *archetypeapp.DeleteArchetypeUseCase
	addTraitUseCase            *archetypeapp.AddTraitToArchetypeUseCase
	removeTraitUseCase         *archetypeapp.RemoveTraitFromArchetypeUseCase
	getTraitsUseCase          *archetypeapp.GetArchetypeTraitsUseCase
	logger                     logger.Logger
}

// NewArchetypeHandler creates a new ArchetypeHandler
func NewArchetypeHandler(
	createArchetypeUseCase *archetypeapp.CreateArchetypeUseCase,
	getArchetypeUseCase *archetypeapp.GetArchetypeUseCase,
	listArchetypesUseCase *archetypeapp.ListArchetypesUseCase,
	updateArchetypeUseCase *archetypeapp.UpdateArchetypeUseCase,
	deleteArchetypeUseCase *archetypeapp.DeleteArchetypeUseCase,
	addTraitUseCase *archetypeapp.AddTraitToArchetypeUseCase,
	removeTraitUseCase *archetypeapp.RemoveTraitFromArchetypeUseCase,
	getTraitsUseCase *archetypeapp.GetArchetypeTraitsUseCase,
	logger logger.Logger,
) *ArchetypeHandler {
	return &ArchetypeHandler{
		createArchetypeUseCase: createArchetypeUseCase,
		getArchetypeUseCase:    getArchetypeUseCase,
		listArchetypesUseCase:  listArchetypesUseCase,
		updateArchetypeUseCase: updateArchetypeUseCase,
		deleteArchetypeUseCase: deleteArchetypeUseCase,
		addTraitUseCase:        addTraitUseCase,
		removeTraitUseCase:     removeTraitUseCase,
		getTraitsUseCase:       getTraitsUseCase,
		logger:                 logger,
	}
}

// CreateArchetype creates a new archetype
func (h *ArchetypeHandler) CreateArchetype(ctx context.Context, req *archetypepb.CreateArchetypeRequest) (*archetypepb.CreateArchetypeResponse, error) {
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

	input := archetypeapp.CreateArchetypeInput{
		TenantID:    tenantUUID,
		Name:        req.Name,
		Description: req.Description,
	}

	output, err := h.createArchetypeUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &archetypepb.CreateArchetypeResponse{
		Archetype: mappers.ArchetypeToProto(output.Archetype),
	}, nil
}

// GetArchetype retrieves an archetype by ID
func (h *ArchetypeHandler) GetArchetype(ctx context.Context, req *archetypepb.GetArchetypeRequest) (*archetypepb.GetArchetypeResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid archetype id: %v", err)
	}

	output, err := h.getArchetypeUseCase.Execute(ctx, archetypeapp.GetArchetypeInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &archetypepb.GetArchetypeResponse{
		Archetype: mappers.ArchetypeToProto(output.Archetype),
	}, nil
}

// ListArchetypes lists archetypes for a tenant
func (h *ArchetypeHandler) ListArchetypes(ctx context.Context, req *archetypepb.ListArchetypesRequest) (*archetypepb.ListArchetypesResponse, error) {
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

	output, err := h.listArchetypesUseCase.Execute(ctx, archetypeapp.ListArchetypesInput{
		TenantID: tenantUUID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}

	archetypes := make([]*archetypepb.Archetype, len(output.Archetypes))
	for i, a := range output.Archetypes {
		archetypes[i] = mappers.ArchetypeToProto(a)
	}

	return &archetypepb.ListArchetypesResponse{
		Archetypes: archetypes,
		TotalCount: int32(output.Total),
	}, nil
}

// UpdateArchetype updates an existing archetype
func (h *ArchetypeHandler) UpdateArchetype(ctx context.Context, req *archetypepb.UpdateArchetypeRequest) (*archetypepb.UpdateArchetypeResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid archetype id: %v", err)
	}

	var name *string
	if req.Name != nil {
		name = req.Name
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	input := archetypeapp.UpdateArchetypeInput{
		TenantID:    tenantUUID,
		ID:          id,
		Name:        name,
		Description: description,
	}

	output, err := h.updateArchetypeUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &archetypepb.UpdateArchetypeResponse{
		Archetype: mappers.ArchetypeToProto(output.Archetype),
	}, nil
}

// DeleteArchetype deletes an archetype
func (h *ArchetypeHandler) DeleteArchetype(ctx context.Context, req *archetypepb.DeleteArchetypeRequest) (*archetypepb.DeleteArchetypeResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid archetype id: %v", err)
	}

	err = h.deleteArchetypeUseCase.Execute(ctx, archetypeapp.DeleteArchetypeInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &archetypepb.DeleteArchetypeResponse{}, nil
}

// AddTraitToArchetype adds a trait to an archetype
func (h *ArchetypeHandler) AddTraitToArchetype(ctx context.Context, req *archetypepb.AddTraitToArchetypeRequest) (*archetypepb.AddTraitToArchetypeResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	archetypeID, err := uuid.Parse(req.ArchetypeId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid archetype id: %v", err)
	}

	traitID, err := uuid.Parse(req.TraitId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid trait id: %v", err)
	}

	err = h.addTraitUseCase.Execute(ctx, archetypeapp.AddTraitToArchetypeInput{
		TenantID:     tenantUUID,
		ArchetypeID:  archetypeID,
		TraitID:      traitID,
		DefaultValue: req.DefaultValue,
	})
	if err != nil {
		return nil, err
	}

	return &archetypepb.AddTraitToArchetypeResponse{}, nil
}

// RemoveTraitFromArchetype removes a trait from an archetype
func (h *ArchetypeHandler) RemoveTraitFromArchetype(ctx context.Context, req *archetypepb.RemoveTraitFromArchetypeRequest) (*archetypepb.RemoveTraitFromArchetypeResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	archetypeID, err := uuid.Parse(req.ArchetypeId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid archetype id: %v", err)
	}

	traitID, err := uuid.Parse(req.TraitId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid trait id: %v", err)
	}

	err = h.removeTraitUseCase.Execute(ctx, archetypeapp.RemoveTraitFromArchetypeInput{
		TenantID:    tenantUUID,
		ArchetypeID: archetypeID,
		TraitID:     traitID,
	})
	if err != nil {
		return nil, err
	}

	return &archetypepb.RemoveTraitFromArchetypeResponse{}, nil
}

// GetArchetypeTraits retrieves all traits for an archetype
func (h *ArchetypeHandler) GetArchetypeTraits(ctx context.Context, req *archetypepb.GetArchetypeTraitsRequest) (*archetypepb.GetArchetypeTraitsResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	archetypeID, err := uuid.Parse(req.ArchetypeId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid archetype_id: %v", err)
	}

	output, err := h.getTraitsUseCase.Execute(ctx, archetypeapp.GetArchetypeTraitsInput{
		TenantID:    tenantUUID,
		ArchetypeID: archetypeID,
	})
	if err != nil {
		return nil, err
	}

	traits := make([]*archetypepb.ArchetypeTrait, len(output.Traits))
	for i, at := range output.Traits {
		traits[i] = mappers.ArchetypeTraitToProto(at, nil)
	}

	return &archetypepb.GetArchetypeTraitsResponse{
		Traits: traits,
	}, nil
}


