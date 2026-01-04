package handlers

import (
	"context"

	"github.com/google/uuid"
	factionapp "github.com/story-engine/main-service/internal/application/world/faction"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	factionpb "github.com/story-engine/main-service/proto/faction"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FactionHandler implements the FactionService gRPC service
type FactionHandler struct {
	factionpb.UnimplementedFactionServiceServer
	createFactionUseCase   *factionapp.CreateFactionUseCase
	getFactionUseCase      *factionapp.GetFactionUseCase
	listFactionsUseCase    *factionapp.ListFactionsUseCase
	updateFactionUseCase   *factionapp.UpdateFactionUseCase
	deleteFactionUseCase   *factionapp.DeleteFactionUseCase
	getChildrenUseCase     *factionapp.GetChildrenUseCase
	addReferenceUseCase    *factionapp.AddReferenceUseCase
	removeReferenceUseCase *factionapp.RemoveReferenceUseCase
	getReferencesUseCase   *factionapp.GetReferencesUseCase
	updateReferenceUseCase *factionapp.UpdateReferenceUseCase
	logger                 logger.Logger
}

// NewFactionHandler creates a new FactionHandler
func NewFactionHandler(
	createFactionUseCase *factionapp.CreateFactionUseCase,
	getFactionUseCase *factionapp.GetFactionUseCase,
	listFactionsUseCase *factionapp.ListFactionsUseCase,
	updateFactionUseCase *factionapp.UpdateFactionUseCase,
	deleteFactionUseCase *factionapp.DeleteFactionUseCase,
	getChildrenUseCase *factionapp.GetChildrenUseCase,
	addReferenceUseCase *factionapp.AddReferenceUseCase,
	removeReferenceUseCase *factionapp.RemoveReferenceUseCase,
	getReferencesUseCase *factionapp.GetReferencesUseCase,
	updateReferenceUseCase *factionapp.UpdateReferenceUseCase,
	logger logger.Logger,
) *FactionHandler {
	return &FactionHandler{
		createFactionUseCase:   createFactionUseCase,
		getFactionUseCase:      getFactionUseCase,
		listFactionsUseCase:    listFactionsUseCase,
		updateFactionUseCase:   updateFactionUseCase,
		deleteFactionUseCase:   deleteFactionUseCase,
		getChildrenUseCase:     getChildrenUseCase,
		addReferenceUseCase:    addReferenceUseCase,
		removeReferenceUseCase: removeReferenceUseCase,
		getReferencesUseCase:   getReferencesUseCase,
		updateReferenceUseCase: updateReferenceUseCase,
		logger:                 logger,
	}
}

// CreateFaction creates a new faction
func (h *FactionHandler) CreateFaction(ctx context.Context, req *factionpb.CreateFactionRequest) (*factionpb.CreateFactionResponse, error) {
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

	input := factionapp.CreateFactionInput{
		TenantID:    tenantUUID,
		WorldID:     worldID,
		ParentID:    parentID,
		Name:        req.Name,
		Type:        req.Type,
		Description: getStringValue(req.Description),
		Beliefs:     getStringValue(req.Beliefs),
		Structure:   getStringValue(req.Structure),
		Symbols:     getStringValue(req.Symbols),
	}

	output, err := h.createFactionUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &factionpb.CreateFactionResponse{
		Faction: mappers.FactionToProto(output.Faction),
	}, nil
}

// GetFaction retrieves a faction by ID
func (h *FactionHandler) GetFaction(ctx context.Context, req *factionpb.GetFactionRequest) (*factionpb.GetFactionResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid faction id: %v", err)
	}

	output, err := h.getFactionUseCase.Execute(ctx, factionapp.GetFactionInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &factionpb.GetFactionResponse{
		Faction: mappers.FactionToProto(output.Faction),
	}, nil
}

// ListFactions lists factions for a world
func (h *FactionHandler) ListFactions(ctx context.Context, req *factionpb.ListFactionsRequest) (*factionpb.ListFactionsResponse, error) {
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

	output, err := h.listFactionsUseCase.Execute(ctx, factionapp.ListFactionsInput{
		TenantID: tenantUUID,
		WorldID:  worldID,
	})
	if err != nil {
		return nil, err
	}

	factions := make([]*factionpb.Faction, len(output.Factions))
	for i, f := range output.Factions {
		factions[i] = mappers.FactionToProto(f)
	}

	return &factionpb.ListFactionsResponse{
		Factions:   factions,
		TotalCount: int32(len(factions)),
	}, nil
}

// UpdateFaction updates an existing faction
func (h *FactionHandler) UpdateFaction(ctx context.Context, req *factionpb.UpdateFactionRequest) (*factionpb.UpdateFactionResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid faction id: %v", err)
	}

	input := factionapp.UpdateFactionInput{
		TenantID:    tenantUUID,
		ID:          id,
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Beliefs:     req.Beliefs,
		Structure:   req.Structure,
		Symbols:     req.Symbols,
	}

	output, err := h.updateFactionUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &factionpb.UpdateFactionResponse{
		Faction: mappers.FactionToProto(output.Faction),
	}, nil
}

// DeleteFaction deletes a faction
func (h *FactionHandler) DeleteFaction(ctx context.Context, req *factionpb.DeleteFactionRequest) (*factionpb.DeleteFactionResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid faction id: %v", err)
	}

	err = h.deleteFactionUseCase.Execute(ctx, factionapp.DeleteFactionInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &factionpb.DeleteFactionResponse{}, nil
}

// AddReferenceToFaction adds a reference to a faction
func (h *FactionHandler) AddReferenceToFaction(ctx context.Context, req *factionpb.AddReferenceToFactionRequest) (*factionpb.AddReferenceToFactionResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	factionID, err := uuid.Parse(req.FactionId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid faction_id: %v", err)
	}

	entityID, err := uuid.Parse(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id: %v", err)
	}

	err = h.addReferenceUseCase.Execute(ctx, factionapp.AddReferenceInput{
		TenantID:   tenantUUID,
		FactionID:  factionID,
		EntityType: req.EntityType,
		EntityID:   entityID,
		Role:       req.Role,
		Notes:      getStringValue(req.Notes),
	})
	if err != nil {
		return nil, err
	}

	// Get the created reference
	output, err := h.getReferencesUseCase.Execute(ctx, factionapp.GetReferencesInput{
		TenantID:  tenantUUID,
		FactionID: factionID,
	})
	if err != nil {
		return nil, err
	}

	// Find the newly created reference
	var ref *world.FactionReference
	for _, r := range output.References {
		if r.EntityType == req.EntityType && r.EntityID == entityID {
			ref = r
			break
		}
	}

	if ref == nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve created reference")
	}

	return &factionpb.AddReferenceToFactionResponse{
		FactionReference: mappers.FactionReferenceToProto(ref),
	}, nil
}

// RemoveReferenceFromFaction removes a reference from a faction
func (h *FactionHandler) RemoveReferenceFromFaction(ctx context.Context, req *factionpb.RemoveReferenceFromFactionRequest) (*factionpb.RemoveReferenceFromFactionResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	factionID, err := uuid.Parse(req.FactionId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid faction_id: %v", err)
	}

	entityID, err := uuid.Parse(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id: %v", err)
	}

	err = h.removeReferenceUseCase.Execute(ctx, factionapp.RemoveReferenceInput{
		TenantID:   tenantUUID,
		FactionID:  factionID,
		EntityType: req.EntityType,
		EntityID:   entityID,
	})
	if err != nil {
		return nil, err
	}

	return &factionpb.RemoveReferenceFromFactionResponse{}, nil
}

// GetFactionReferences lists references for a faction
func (h *FactionHandler) GetFactionReferences(ctx context.Context, req *factionpb.GetFactionReferencesRequest) (*factionpb.GetFactionReferencesResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	factionID, err := uuid.Parse(req.FactionId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid faction_id: %v", err)
	}

	output, err := h.getReferencesUseCase.Execute(ctx, factionapp.GetReferencesInput{
		TenantID:  tenantUUID,
		FactionID: factionID,
	})
	if err != nil {
		return nil, err
	}

	references := make([]*factionpb.FactionReference, len(output.References))
	for i, ref := range output.References {
		references[i] = mappers.FactionReferenceToProto(ref)
	}

	return &factionpb.GetFactionReferencesResponse{
		References: references,
		TotalCount: int32(len(references)),
	}, nil
}

// UpdateFactionReference updates a faction reference
func (h *FactionHandler) UpdateFactionReference(ctx context.Context, req *factionpb.UpdateFactionReferenceRequest) (*factionpb.UpdateFactionReferenceResponse, error) {
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

	err = h.updateReferenceUseCase.Execute(ctx, factionapp.UpdateReferenceInput{
		TenantID: tenantUUID,
		ID:       id,
		Role:     req.Role,
		Notes:    req.Notes,
	})
	if err != nil {
		return nil, err
	}

	// Get the updated reference - we need to find it from the faction
	// For simplicity, we'll just return success. The client can fetch if needed.
	return &factionpb.UpdateFactionReferenceResponse{}, nil
}

// GetFactionChildren lists children of a faction
func (h *FactionHandler) GetFactionChildren(ctx context.Context, req *factionpb.GetFactionChildrenRequest) (*factionpb.GetFactionChildrenResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid faction id: %v", err)
	}

	output, err := h.getChildrenUseCase.Execute(ctx, factionapp.GetChildrenInput{
		TenantID:  tenantUUID,
		FactionID: id,
	})
	if err != nil {
		return nil, err
	}

	children := make([]*factionpb.Faction, len(output.Children))
	for i, f := range output.Children {
		children[i] = mappers.FactionToProto(f)
	}

	return &factionpb.GetFactionChildrenResponse{
		Children:   children,
		TotalCount: int32(len(children)),
	}, nil
}

// Helper function to get string value from optional string
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

