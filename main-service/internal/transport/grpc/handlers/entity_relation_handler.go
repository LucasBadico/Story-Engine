package handlers

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	commonpb "github.com/story-engine/main-service/proto/common"
	entityrelationpb "github.com/story-engine/main-service/proto/entity_relation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// EntityRelationHandler implements the EntityRelationService gRPC service
type EntityRelationHandler struct {
	entityrelationpb.UnimplementedEntityRelationServiceServer
	createRelationUseCase        *relationapp.CreateRelationUseCase
	getRelationUseCase           *relationapp.GetRelationUseCase
	listRelationsBySourceUseCase *relationapp.ListRelationsBySourceUseCase
	listRelationsByTargetUseCase *relationapp.ListRelationsByTargetUseCase
	listRelationsByWorldUseCase  *relationapp.ListRelationsByWorldUseCase
	updateRelationUseCase        *relationapp.UpdateRelationUseCase
	deleteRelationUseCase        *relationapp.DeleteRelationUseCase
	logger                       logger.Logger
}

// NewEntityRelationHandler creates a new EntityRelationHandler
func NewEntityRelationHandler(
	createRelationUseCase *relationapp.CreateRelationUseCase,
	getRelationUseCase *relationapp.GetRelationUseCase,
	listRelationsBySourceUseCase *relationapp.ListRelationsBySourceUseCase,
	listRelationsByTargetUseCase *relationapp.ListRelationsByTargetUseCase,
	listRelationsByWorldUseCase *relationapp.ListRelationsByWorldUseCase,
	updateRelationUseCase *relationapp.UpdateRelationUseCase,
	deleteRelationUseCase *relationapp.DeleteRelationUseCase,
	logger logger.Logger,
) *EntityRelationHandler {
	return &EntityRelationHandler{
		createRelationUseCase:        createRelationUseCase,
		getRelationUseCase:           getRelationUseCase,
		listRelationsBySourceUseCase: listRelationsBySourceUseCase,
		listRelationsByTargetUseCase: listRelationsByTargetUseCase,
		listRelationsByWorldUseCase:  listRelationsByWorldUseCase,
		updateRelationUseCase:        updateRelationUseCase,
		deleteRelationUseCase:        deleteRelationUseCase,
		logger:                       logger,
	}
}

// CreateRelation creates a new relation
// Both source_id and target_id are required - entities must exist
func (h *EntityRelationHandler) CreateRelation(ctx context.Context, req *entityrelationpb.CreateRelationRequest) (*entityrelationpb.CreateRelationResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	worldID, err := uuid.Parse(req.WorldId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid world_id: %v", err)
	}

	// source_id is required
	if req.SourceId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "source_id is required")
	}
	sourceID, err := uuid.Parse(req.SourceId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid source_id: %v", err)
	}

	// target_id is required
	if req.TargetId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "target_id is required")
	}
	targetID, err := uuid.Parse(req.TargetId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid target_id: %v", err)
	}

	var contextID *uuid.UUID
	if req.ContextId != nil && *req.ContextId != "" {
		id, err := uuid.Parse(*req.ContextId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid context_id: %v", err)
		}
		contextID = &id
	}

	var createdByUserID *uuid.UUID
	if req.CreatedByUserId != nil && *req.CreatedByUserId != "" {
		id, err := uuid.Parse(*req.CreatedByUserId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid created_by_user_id: %v", err)
		}
		createdByUserID = &id
	}

	// Parse attributes JSON
	var attributes map[string]interface{}
	if req.AttributesJson != "" {
		if err := json.Unmarshal([]byte(req.AttributesJson), &attributes); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid attributes_json: %v", err)
		}
	}

	input := relationapp.CreateRelationInput{
		TenantID:        tenantUUID,
		WorldID:         worldID,
		SourceType:      req.SourceType,
		SourceID:        sourceID,
		TargetType:      req.TargetType,
		TargetID:        targetID,
		RelationType:    req.RelationType,
		ContextType:     req.ContextType,
		ContextID:       contextID,
		Attributes:      attributes,
		Summary:         req.Summary,
		CreatedByUserID: createdByUserID,
		CreateMirror:    req.CreateMirror,
	}

	output, err := h.createRelationUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	response := &entityrelationpb.CreateRelationResponse{
		Relation: mappers.EntityRelationToProto(output.Relation),
	}

	if output.Mirror != nil {
		mirrorProto := mappers.EntityRelationToProto(output.Mirror)
		response.Mirror = mirrorProto
	}

	return response, nil
}

// GetRelation retrieves a relation by ID
func (h *EntityRelationHandler) GetRelation(ctx context.Context, req *entityrelationpb.GetRelationRequest) (*entityrelationpb.GetRelationResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid relation id: %v", err)
	}

	output, err := h.getRelationUseCase.Execute(ctx, relationapp.GetRelationInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &entityrelationpb.GetRelationResponse{
		Relation: mappers.EntityRelationToProto(output.Relation),
	}, nil
}

// ListRelationsBySource lists relations by source
func (h *EntityRelationHandler) ListRelationsBySource(ctx context.Context, req *entityrelationpb.ListRelationsBySourceRequest) (*entityrelationpb.ListRelationsResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	sourceID, err := uuid.Parse(req.SourceId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid source_id: %v", err)
	}

	opts := h.parseListOptions(req.Pagination, req.RelationType, req.ExcludeMirrors)

	output, err := h.listRelationsBySourceUseCase.Execute(ctx, relationapp.ListRelationsBySourceInput{
		TenantID:   tenantUUID,
		SourceType: req.SourceType,
		SourceID:   sourceID,
		Options:    opts,
	})
	if err != nil {
		return nil, err
	}

	return h.buildListResponse(output.Relations), nil
}

// ListRelationsByTarget lists relations by target
func (h *EntityRelationHandler) ListRelationsByTarget(ctx context.Context, req *entityrelationpb.ListRelationsByTargetRequest) (*entityrelationpb.ListRelationsResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	targetID, err := uuid.Parse(req.TargetId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid target_id: %v", err)
	}

	opts := h.parseListOptions(req.Pagination, req.RelationType, req.ExcludeMirrors)

	output, err := h.listRelationsByTargetUseCase.Execute(ctx, relationapp.ListRelationsByTargetInput{
		TenantID:   tenantUUID,
		TargetType: req.TargetType,
		TargetID:   targetID,
		Options:    opts,
	})
	if err != nil {
		return nil, err
	}

	return h.buildListResponse(output.Relations), nil
}

// ListRelationsByWorld lists relations by world
func (h *EntityRelationHandler) ListRelationsByWorld(ctx context.Context, req *entityrelationpb.ListRelationsByWorldRequest) (*entityrelationpb.ListRelationsResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	worldID, err := uuid.Parse(req.WorldId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid world_id: %v", err)
	}

	opts := h.parseListOptions(req.Pagination, req.RelationType, req.ExcludeMirrors)

	output, err := h.listRelationsByWorldUseCase.Execute(ctx, relationapp.ListRelationsByWorldInput{
		TenantID: tenantUUID,
		WorldID:  worldID,
		Options:  opts,
	})
	if err != nil {
		return nil, err
	}

	return h.buildListResponse(output.Relations), nil
}

// UpdateRelation updates an existing relation
func (h *EntityRelationHandler) UpdateRelation(ctx context.Context, req *entityrelationpb.UpdateRelationRequest) (*entityrelationpb.UpdateRelationResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid relation id: %v", err)
	}

	var relationType *string
	if req.RelationType != nil {
		relationType = req.RelationType
	}

	var attributes *map[string]interface{}
	if req.AttributesJson != nil && *req.AttributesJson != "" {
		var attrs map[string]interface{}
		if err := json.Unmarshal([]byte(*req.AttributesJson), &attrs); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid attributes_json: %v", err)
		}
		attributes = &attrs
	}

	input := relationapp.UpdateRelationInput{
		TenantID:     tenantUUID,
		ID:           id,
		RelationType: relationType,
		Attributes:   attributes,
		Summary:      req.Summary,
	}

	output, err := h.updateRelationUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &entityrelationpb.UpdateRelationResponse{
		Relation: mappers.EntityRelationToProto(output.Relation),
	}, nil
}

// DeleteRelation deletes a relation
func (h *EntityRelationHandler) DeleteRelation(ctx context.Context, req *entityrelationpb.DeleteRelationRequest) (*emptypb.Empty, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid relation id: %v", err)
	}

	err = h.deleteRelationUseCase.Execute(ctx, relationapp.DeleteRelationInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// Helper methods

func (h *EntityRelationHandler) parseListOptions(
	pagination *commonpb.PaginationRequest,
	relationType *string,
	excludeMirrors bool,
) repositories.ListOptions {
	opts := repositories.ListOptions{
		ExcludeMirrors: excludeMirrors,
	}

	if relationType != nil {
		opts.RelationType = relationType
	}

	if pagination != nil {
		if pagination.Limit > 0 {
			opts.Limit = int(pagination.Limit)
		}
	}

	return opts
}

func (h *EntityRelationHandler) buildListResponse(result *repositories.ListResult) *entityrelationpb.ListRelationsResponse {
	relations := make([]*entityrelationpb.EntityRelation, len(result.Items))
	for i, r := range result.Items {
		relations[i] = mappers.EntityRelationToProto(r)
	}

	return &entityrelationpb.ListRelationsResponse{
		Relations:  relations,
		TotalCount: int32(len(relations)),
	}
}
