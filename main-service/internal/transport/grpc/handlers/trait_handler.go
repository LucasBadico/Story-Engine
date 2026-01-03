package handlers

import (
	"context"

	"github.com/google/uuid"
	traitapp "github.com/story-engine/main-service/internal/application/world/trait"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	"github.com/story-engine/main-service/internal/platform/logger"
	traitpb "github.com/story-engine/main-service/proto/trait"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TraitHandler implements the TraitService gRPC service
type TraitHandler struct {
	traitpb.UnimplementedTraitServiceServer
	createTraitUseCase *traitapp.CreateTraitUseCase
	getTraitUseCase    *traitapp.GetTraitUseCase
	listTraitsUseCase  *traitapp.ListTraitsUseCase
	updateTraitUseCase *traitapp.UpdateTraitUseCase
	deleteTraitUseCase *traitapp.DeleteTraitUseCase
	logger             logger.Logger
}

// NewTraitHandler creates a new TraitHandler
func NewTraitHandler(
	createTraitUseCase *traitapp.CreateTraitUseCase,
	getTraitUseCase *traitapp.GetTraitUseCase,
	listTraitsUseCase *traitapp.ListTraitsUseCase,
	updateTraitUseCase *traitapp.UpdateTraitUseCase,
	deleteTraitUseCase *traitapp.DeleteTraitUseCase,
	logger logger.Logger,
) *TraitHandler {
	return &TraitHandler{
		createTraitUseCase: createTraitUseCase,
		getTraitUseCase:    getTraitUseCase,
		listTraitsUseCase:  listTraitsUseCase,
		updateTraitUseCase: updateTraitUseCase,
		deleteTraitUseCase: deleteTraitUseCase,
		logger:             logger,
	}
}

// CreateTrait creates a new trait
func (h *TraitHandler) CreateTrait(ctx context.Context, req *traitpb.CreateTraitRequest) (*traitpb.CreateTraitResponse, error) {
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

	input := traitapp.CreateTraitInput{
		TenantID:    tenantUUID,
		Name:        req.Name,
		Category:    req.Category,
		Description: req.Description,
	}

	output, err := h.createTraitUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &traitpb.CreateTraitResponse{
		Trait: mappers.TraitToProto(output.Trait),
	}, nil
}

// GetTrait retrieves a trait by ID
func (h *TraitHandler) GetTrait(ctx context.Context, req *traitpb.GetTraitRequest) (*traitpb.GetTraitResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid trait id: %v", err)
	}

	output, err := h.getTraitUseCase.Execute(ctx, traitapp.GetTraitInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &traitpb.GetTraitResponse{
		Trait: mappers.TraitToProto(output.Trait),
	}, nil
}

// ListTraits lists traits for a tenant
func (h *TraitHandler) ListTraits(ctx context.Context, req *traitpb.ListTraitsRequest) (*traitpb.ListTraitsResponse, error) {
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

	output, err := h.listTraitsUseCase.Execute(ctx, traitapp.ListTraitsInput{
		TenantID: tenantUUID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}

	traits := make([]*traitpb.Trait, len(output.Traits))
	for i, t := range output.Traits {
		traits[i] = mappers.TraitToProto(t)
	}

	return &traitpb.ListTraitsResponse{
		Traits:     traits,
		TotalCount: int32(output.Total),
	}, nil
}

// UpdateTrait updates an existing trait
func (h *TraitHandler) UpdateTrait(ctx context.Context, req *traitpb.UpdateTraitRequest) (*traitpb.UpdateTraitResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid trait id: %v", err)
	}

	var name *string
	if req.Name != nil {
		name = req.Name
	}

	var category *string
	if req.Category != nil {
		category = req.Category
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

	input := traitapp.UpdateTraitInput{
		TenantID:    tenantUUID,
		ID:          id,
		Name:        name,
		Category:    category,
		Description: description,
	}

	output, err := h.updateTraitUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &traitpb.UpdateTraitResponse{
		Trait: mappers.TraitToProto(output.Trait),
	}, nil
}

// DeleteTrait deletes a trait
func (h *TraitHandler) DeleteTrait(ctx context.Context, req *traitpb.DeleteTraitRequest) (*traitpb.DeleteTraitResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid trait id: %v", err)
	}

	err = h.deleteTraitUseCase.Execute(ctx, traitapp.DeleteTraitInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &traitpb.DeleteTraitResponse{}, nil
}


