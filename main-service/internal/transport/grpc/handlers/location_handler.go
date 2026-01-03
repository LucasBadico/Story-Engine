package handlers

import (
	"context"

	"github.com/google/uuid"
	locationapp "github.com/story-engine/main-service/internal/application/world/location"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	locationpb "github.com/story-engine/main-service/proto/location"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LocationHandler implements the LocationService gRPC service
type LocationHandler struct {
	locationpb.UnimplementedLocationServiceServer
	createLocationUseCase   *locationapp.CreateLocationUseCase
	getLocationUseCase       *locationapp.GetLocationUseCase
	listLocationsUseCase     *locationapp.ListLocationsUseCase
	updateLocationUseCase    *locationapp.UpdateLocationUseCase
	deleteLocationUseCase    *locationapp.DeleteLocationUseCase
	getChildrenUseCase       *locationapp.GetChildrenUseCase
	getAncestorsUseCase      *locationapp.GetAncestorsUseCase
	getDescendantsUseCase    *locationapp.GetDescendantsUseCase
	moveLocationUseCase      *locationapp.MoveLocationUseCase
	logger                   logger.Logger
}

// NewLocationHandler creates a new LocationHandler
func NewLocationHandler(
	createLocationUseCase *locationapp.CreateLocationUseCase,
	getLocationUseCase *locationapp.GetLocationUseCase,
	listLocationsUseCase *locationapp.ListLocationsUseCase,
	updateLocationUseCase *locationapp.UpdateLocationUseCase,
	deleteLocationUseCase *locationapp.DeleteLocationUseCase,
	getChildrenUseCase *locationapp.GetChildrenUseCase,
	getAncestorsUseCase *locationapp.GetAncestorsUseCase,
	getDescendantsUseCase *locationapp.GetDescendantsUseCase,
	moveLocationUseCase *locationapp.MoveLocationUseCase,
	logger logger.Logger,
) *LocationHandler {
	return &LocationHandler{
		createLocationUseCase: createLocationUseCase,
		getLocationUseCase:    getLocationUseCase,
		listLocationsUseCase:  listLocationsUseCase,
		updateLocationUseCase: updateLocationUseCase,
		deleteLocationUseCase: deleteLocationUseCase,
		getChildrenUseCase:    getChildrenUseCase,
		getAncestorsUseCase:   getAncestorsUseCase,
		getDescendantsUseCase: getDescendantsUseCase,
		moveLocationUseCase:   moveLocationUseCase,
		logger:                logger,
	}
}

// CreateLocation creates a new location
func (h *LocationHandler) CreateLocation(ctx context.Context, req *locationpb.CreateLocationRequest) (*locationpb.CreateLocationResponse, error) {
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
	if req.ParentId != "" {
		pid, err := uuid.Parse(req.ParentId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid parent_id: %v", err)
		}
		parentID = &pid
	}

	input := locationapp.CreateLocationInput{
		TenantID:    tenantUUID,
		WorldID:     worldID,
		ParentID:    parentID,
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
	}

	output, err := h.createLocationUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &locationpb.CreateLocationResponse{
		Location: mappers.LocationToProto(output.Location),
	}, nil
}

// GetLocation retrieves a location by ID
func (h *LocationHandler) GetLocation(ctx context.Context, req *locationpb.GetLocationRequest) (*locationpb.GetLocationResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid location id: %v", err)
	}

	output, err := h.getLocationUseCase.Execute(ctx, locationapp.GetLocationInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &locationpb.GetLocationResponse{
		Location: mappers.LocationToProto(output.Location),
	}, nil
}

// ListLocations lists locations for a world
func (h *LocationHandler) ListLocations(ctx context.Context, req *locationpb.ListLocationsRequest) (*locationpb.ListLocationsResponse, error) {
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

	output, err := h.listLocationsUseCase.Execute(ctx, locationapp.ListLocationsInput{
		TenantID: tenantUUID,
		WorldID:  worldID,
		Format:   req.Format,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}

	locations := make([]*locationpb.Location, len(output.Locations))
	for i, l := range output.Locations {
		locations[i] = mappers.LocationToProto(l)
	}

	return &locationpb.ListLocationsResponse{
		Locations:  locations,
		TotalCount: int32(output.Total),
	}, nil
}

// UpdateLocation updates an existing location
func (h *LocationHandler) UpdateLocation(ctx context.Context, req *locationpb.UpdateLocationRequest) (*locationpb.UpdateLocationResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid location id: %v", err)
	}

	var name *string
	if req.Name != nil {
		name = req.Name
	}

	var locationType *string
	if req.Type != nil {
		locationType = req.Type
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

	input := locationapp.UpdateLocationInput{
		TenantID:    tenantUUID,
		ID:          id,
		Name:        name,
		Type:        locationType,
		Description: description,
	}

	output, err := h.updateLocationUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &locationpb.UpdateLocationResponse{
		Location: mappers.LocationToProto(output.Location),
	}, nil
}

// DeleteLocation deletes a location
func (h *LocationHandler) DeleteLocation(ctx context.Context, req *locationpb.DeleteLocationRequest) (*locationpb.DeleteLocationResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid location id: %v", err)
	}

	err = h.deleteLocationUseCase.Execute(ctx, locationapp.DeleteLocationInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &locationpb.DeleteLocationResponse{}, nil
}

// GetChildren retrieves direct children of a location
func (h *LocationHandler) GetChildren(ctx context.Context, req *locationpb.GetChildrenRequest) (*locationpb.GetChildrenResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	locationID, err := uuid.Parse(req.LocationId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid location_id: %v", err)
	}

	output, err := h.getChildrenUseCase.Execute(ctx, locationapp.GetChildrenInput{
		TenantID:   tenantUUID,
		LocationID: locationID,
	})
	if err != nil {
		return nil, err
	}

	children := make([]*locationpb.Location, len(output.Children))
	for i, l := range output.Children {
		children[i] = mappers.LocationToProto(l)
	}

	return &locationpb.GetChildrenResponse{
		Children: children,
	}, nil
}

// GetAncestors retrieves all ancestors of a location
func (h *LocationHandler) GetAncestors(ctx context.Context, req *locationpb.GetAncestorsRequest) (*locationpb.GetAncestorsResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	locationID, err := uuid.Parse(req.LocationId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid location_id: %v", err)
	}

	output, err := h.getAncestorsUseCase.Execute(ctx, locationapp.GetAncestorsInput{
		TenantID:   tenantUUID,
		LocationID: locationID,
	})
	if err != nil {
		return nil, err
	}

	ancestors := make([]*locationpb.Location, len(output.Ancestors))
	for i, l := range output.Ancestors {
		ancestors[i] = mappers.LocationToProto(l)
	}

	return &locationpb.GetAncestorsResponse{
		Ancestors: ancestors,
	}, nil
}

// GetDescendants retrieves all descendants of a location
func (h *LocationHandler) GetDescendants(ctx context.Context, req *locationpb.GetDescendantsRequest) (*locationpb.GetDescendantsResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	locationID, err := uuid.Parse(req.LocationId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid location_id: %v", err)
	}

	output, err := h.getDescendantsUseCase.Execute(ctx, locationapp.GetDescendantsInput{
		TenantID:   tenantUUID,
		LocationID: locationID,
	})
	if err != nil {
		return nil, err
	}

	descendants := make([]*locationpb.Location, len(output.Descendants))
	for i, l := range output.Descendants {
		descendants[i] = mappers.LocationToProto(l)
	}

	return &locationpb.GetDescendantsResponse{
		Descendants: descendants,
	}, nil
}

// MoveLocation moves a location to a different parent
func (h *LocationHandler) MoveLocation(ctx context.Context, req *locationpb.MoveLocationRequest) (*locationpb.MoveLocationResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	locationID, err := uuid.Parse(req.LocationId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid location_id: %v", err)
	}

	var parentID *uuid.UUID
	if req.ParentId != "" {
		pid, err := uuid.Parse(req.ParentId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid parent_id: %v", err)
		}
		parentID = &pid
	}

	output, err := h.moveLocationUseCase.Execute(ctx, locationapp.MoveLocationInput{
		TenantID:    tenantUUID,
		LocationID:  locationID,
		NewParentID: parentID,
	})
	if err != nil {
		return nil, err
	}

	return &locationpb.MoveLocationResponse{
		Location: mappers.LocationToProto(output.Location),
	}, nil
}


