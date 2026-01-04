package handlers

import (
	"context"

	"github.com/google/uuid"
	loreapp "github.com/story-engine/main-service/internal/application/world/lore"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	lorepb "github.com/story-engine/main-service/proto/lore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoreHandler implements the LoreService gRPC service
type LoreHandler struct {
	lorepb.UnimplementedLoreServiceServer
	createLoreUseCase      *loreapp.CreateLoreUseCase
	getLoreUseCase         *loreapp.GetLoreUseCase
	listLoresUseCase       *loreapp.ListLoresUseCase
	updateLoreUseCase      *loreapp.UpdateLoreUseCase
	deleteLoreUseCase      *loreapp.DeleteLoreUseCase
	getChildrenUseCase     *loreapp.GetChildrenUseCase
	addReferenceUseCase    *loreapp.AddReferenceUseCase
	removeReferenceUseCase *loreapp.RemoveReferenceUseCase
	getReferencesUseCase   *loreapp.GetReferencesUseCase
	updateReferenceUseCase *loreapp.UpdateReferenceUseCase
	logger                 logger.Logger
}

// NewLoreHandler creates a new LoreHandler
func NewLoreHandler(
	createLoreUseCase *loreapp.CreateLoreUseCase,
	getLoreUseCase *loreapp.GetLoreUseCase,
	listLoresUseCase *loreapp.ListLoresUseCase,
	updateLoreUseCase *loreapp.UpdateLoreUseCase,
	deleteLoreUseCase *loreapp.DeleteLoreUseCase,
	getChildrenUseCase *loreapp.GetChildrenUseCase,
	addReferenceUseCase *loreapp.AddReferenceUseCase,
	removeReferenceUseCase *loreapp.RemoveReferenceUseCase,
	getReferencesUseCase *loreapp.GetReferencesUseCase,
	updateReferenceUseCase *loreapp.UpdateReferenceUseCase,
	logger logger.Logger,
) *LoreHandler {
	return &LoreHandler{
		createLoreUseCase:      createLoreUseCase,
		getLoreUseCase:         getLoreUseCase,
		listLoresUseCase:       listLoresUseCase,
		updateLoreUseCase:      updateLoreUseCase,
		deleteLoreUseCase:      deleteLoreUseCase,
		getChildrenUseCase:     getChildrenUseCase,
		addReferenceUseCase:    addReferenceUseCase,
		removeReferenceUseCase: removeReferenceUseCase,
		getReferencesUseCase:   getReferencesUseCase,
		updateReferenceUseCase: updateReferenceUseCase,
		logger:                 logger,
	}
}

// CreateLore creates a new lore
func (h *LoreHandler) CreateLore(ctx context.Context, req *lorepb.CreateLoreRequest) (*lorepb.CreateLoreResponse, error) {
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

	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}

	var parentID *uuid.UUID
	if req.ParentId != nil && *req.ParentId != "" {
		pid, err := uuid.Parse(*req.ParentId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid parent_id: %v", err)
		}
		parentID = &pid
	}

	input := loreapp.CreateLoreInput{
		TenantID:     tenantUUID,
		WorldID:      worldID,
		ParentID:     parentID,
		Name:         req.Name,
		Category:     req.Category,
		Description:  getStringValue(req.Description),
		Rules:        getStringValue(req.Rules),
		Limitations:  getStringValue(req.Limitations),
		Requirements: getStringValue(req.Requirements),
	}

	output, err := h.createLoreUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &lorepb.CreateLoreResponse{
		Lore: mappers.LoreToProto(output.Lore),
	}, nil
}

// GetLore retrieves a lore by ID
func (h *LoreHandler) GetLore(ctx context.Context, req *lorepb.GetLoreRequest) (*lorepb.GetLoreResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid lore id: %v", err)
	}

	output, err := h.getLoreUseCase.Execute(ctx, loreapp.GetLoreInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &lorepb.GetLoreResponse{
		Lore: mappers.LoreToProto(output.Lore),
	}, nil
}

// ListLores lists lores for a world
func (h *LoreHandler) ListLores(ctx context.Context, req *lorepb.ListLoresRequest) (*lorepb.ListLoresResponse, error) {
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

	output, err := h.listLoresUseCase.Execute(ctx, loreapp.ListLoresInput{
		TenantID: tenantUUID,
		WorldID:  worldID,
	})
	if err != nil {
		return nil, err
	}

	lores := make([]*lorepb.Lore, len(output.Lores))
	for i, l := range output.Lores {
		lores[i] = mappers.LoreToProto(l)
	}

	return &lorepb.ListLoresResponse{
		Lores:      lores,
		TotalCount: int32(len(lores)),
	}, nil
}

// UpdateLore updates an existing lore
func (h *LoreHandler) UpdateLore(ctx context.Context, req *lorepb.UpdateLoreRequest) (*lorepb.UpdateLoreResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid lore id: %v", err)
	}

	input := loreapp.UpdateLoreInput{
		TenantID:     tenantUUID,
		ID:           id,
		Name:         req.Name,
		Category:     req.Category,
		Description:  req.Description,
		Rules:        req.Rules,
		Limitations:  req.Limitations,
		Requirements: req.Requirements,
	}

	output, err := h.updateLoreUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &lorepb.UpdateLoreResponse{
		Lore: mappers.LoreToProto(output.Lore),
	}, nil
}

// DeleteLore deletes a lore
func (h *LoreHandler) DeleteLore(ctx context.Context, req *lorepb.DeleteLoreRequest) (*lorepb.DeleteLoreResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid lore id: %v", err)
	}

	err = h.deleteLoreUseCase.Execute(ctx, loreapp.DeleteLoreInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &lorepb.DeleteLoreResponse{}, nil
}

// AddReferenceToLore adds a reference to a lore
func (h *LoreHandler) AddReferenceToLore(ctx context.Context, req *lorepb.AddReferenceToLoreRequest) (*lorepb.AddReferenceToLoreResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	loreID, err := uuid.Parse(req.LoreId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid lore_id: %v", err)
	}

	entityID, err := uuid.Parse(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id: %v", err)
	}

	err = h.addReferenceUseCase.Execute(ctx, loreapp.AddReferenceInput{
		TenantID:         tenantUUID,
		LoreID:           loreID,
		EntityType:       req.EntityType,
		EntityID:         entityID,
		RelationshipType: req.RelationshipType,
		Notes:            getStringValue(req.Notes),
	})
	if err != nil {
		return nil, err
	}

	// Get the created reference
	output, err := h.getReferencesUseCase.Execute(ctx, loreapp.GetReferencesInput{
		TenantID: tenantUUID,
		LoreID:   loreID,
	})
	if err != nil {
		return nil, err
	}

	// Find the newly created reference
	var ref *world.LoreReference
	for _, r := range output.References {
		if r.EntityType == req.EntityType && r.EntityID == entityID {
			ref = r
			break
		}
	}

	if ref == nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve created reference")
	}

	return &lorepb.AddReferenceToLoreResponse{
		LoreReference: mappers.LoreReferenceToProto(ref),
	}, nil
}

// RemoveReferenceFromLore removes a reference from a lore
func (h *LoreHandler) RemoveReferenceFromLore(ctx context.Context, req *lorepb.RemoveReferenceFromLoreRequest) (*lorepb.RemoveReferenceFromLoreResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	loreID, err := uuid.Parse(req.LoreId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid lore_id: %v", err)
	}

	entityID, err := uuid.Parse(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id: %v", err)
	}

	err = h.removeReferenceUseCase.Execute(ctx, loreapp.RemoveReferenceInput{
		TenantID:   tenantUUID,
		LoreID:     loreID,
		EntityType: req.EntityType,
		EntityID:   entityID,
	})
	if err != nil {
		return nil, err
	}

	return &lorepb.RemoveReferenceFromLoreResponse{}, nil
}

// GetLoreReferences lists references for a lore
func (h *LoreHandler) GetLoreReferences(ctx context.Context, req *lorepb.GetLoreReferencesRequest) (*lorepb.GetLoreReferencesResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	loreID, err := uuid.Parse(req.LoreId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid lore_id: %v", err)
	}

	output, err := h.getReferencesUseCase.Execute(ctx, loreapp.GetReferencesInput{
		TenantID: tenantUUID,
		LoreID:   loreID,
	})
	if err != nil {
		return nil, err
	}

	references := make([]*lorepb.LoreReference, len(output.References))
	for i, ref := range output.References {
		references[i] = mappers.LoreReferenceToProto(ref)
	}

	return &lorepb.GetLoreReferencesResponse{
		References: references,
		TotalCount: int32(len(references)),
	}, nil
}

// UpdateLoreReference updates a lore reference
func (h *LoreHandler) UpdateLoreReference(ctx context.Context, req *lorepb.UpdateLoreReferenceRequest) (*lorepb.UpdateLoreReferenceResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid reference id: %v", err)
	}

	err = h.updateReferenceUseCase.Execute(ctx, loreapp.UpdateReferenceInput{
		TenantID:         tenantUUID,
		ID:               id,
		RelationshipType: req.RelationshipType,
		Notes:            req.Notes,
	})
	if err != nil {
		return nil, err
	}

	return &lorepb.UpdateLoreReferenceResponse{}, nil
}

// GetLoreChildren lists children of a lore
func (h *LoreHandler) GetLoreChildren(ctx context.Context, req *lorepb.GetLoreChildrenRequest) (*lorepb.GetLoreChildrenResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid lore id: %v", err)
	}

	output, err := h.getChildrenUseCase.Execute(ctx, loreapp.GetChildrenInput{
		TenantID: tenantUUID,
		LoreID:   id,
	})
	if err != nil {
		return nil, err
	}

	children := make([]*lorepb.Lore, len(output.Children))
	for i, l := range output.Children {
		children[i] = mappers.LoreToProto(l)
	}

	return &lorepb.GetLoreChildrenResponse{
		Children:   children,
		TotalCount: int32(len(children)),
	}, nil
}
