package handlers

import (
	"context"

	"github.com/google/uuid"
	artifactapp "github.com/story-engine/main-service/internal/application/world/artifact"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	artifactpb "github.com/story-engine/main-service/proto/artifact"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ArtifactHandler implements the ArtifactService gRPC service
type ArtifactHandler struct {
	artifactpb.UnimplementedArtifactServiceServer
	createArtifactUseCase      *artifactapp.CreateArtifactUseCase
	getArtifactUseCase         *artifactapp.GetArtifactUseCase
	listArtifactsUseCase       *artifactapp.ListArtifactsUseCase
	updateArtifactUseCase      *artifactapp.UpdateArtifactUseCase
	deleteArtifactUseCase      *artifactapp.DeleteArtifactUseCase
	getReferencesUseCase       *artifactapp.GetArtifactReferencesUseCase
	addReferenceUseCase         *artifactapp.AddArtifactReferenceUseCase
	removeReferenceUseCase      *artifactapp.RemoveArtifactReferenceUseCase
	logger                      logger.Logger
}

// NewArtifactHandler creates a new ArtifactHandler
func NewArtifactHandler(
	createArtifactUseCase *artifactapp.CreateArtifactUseCase,
	getArtifactUseCase *artifactapp.GetArtifactUseCase,
	listArtifactsUseCase *artifactapp.ListArtifactsUseCase,
	updateArtifactUseCase *artifactapp.UpdateArtifactUseCase,
	deleteArtifactUseCase *artifactapp.DeleteArtifactUseCase,
	getReferencesUseCase *artifactapp.GetArtifactReferencesUseCase,
	addReferenceUseCase *artifactapp.AddArtifactReferenceUseCase,
	removeReferenceUseCase *artifactapp.RemoveArtifactReferenceUseCase,
	logger logger.Logger,
) *ArtifactHandler {
	return &ArtifactHandler{
		createArtifactUseCase: createArtifactUseCase,
		getArtifactUseCase:    getArtifactUseCase,
		listArtifactsUseCase:  listArtifactsUseCase,
		updateArtifactUseCase: updateArtifactUseCase,
		deleteArtifactUseCase: deleteArtifactUseCase,
		getReferencesUseCase:  getReferencesUseCase,
		addReferenceUseCase:   addReferenceUseCase,
		removeReferenceUseCase: removeReferenceUseCase,
		logger:                logger,
	}
}

// CreateArtifact creates a new artifact
func (h *ArtifactHandler) CreateArtifact(ctx context.Context, req *artifactpb.CreateArtifactRequest) (*artifactpb.CreateArtifactResponse, error) {
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

	characterIDs := make([]uuid.UUID, 0, len(req.CharacterIds))
	for _, cidStr := range req.CharacterIds {
		cid, err := uuid.Parse(cidStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
		}
		characterIDs = append(characterIDs, cid)
	}

	locationIDs := make([]uuid.UUID, 0, len(req.LocationIds))
	for _, lidStr := range req.LocationIds {
		lid, err := uuid.Parse(lidStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid location_id: %v", err)
		}
		locationIDs = append(locationIDs, lid)
	}

	input := artifactapp.CreateArtifactInput{
		TenantID:     tenantUUID,
		WorldID:      worldID,
		CharacterIDs: characterIDs,
		LocationIDs:  locationIDs,
		Name:         req.Name,
		Description:  req.Description,
		Rarity:       req.Rarity,
	}

	output, err := h.createArtifactUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &artifactpb.CreateArtifactResponse{
		Artifact: mappers.ArtifactToProto(output.Artifact),
	}, nil
}

// GetArtifact retrieves an artifact by ID
func (h *ArtifactHandler) GetArtifact(ctx context.Context, req *artifactpb.GetArtifactRequest) (*artifactpb.GetArtifactResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid artifact id: %v", err)
	}

	output, err := h.getArtifactUseCase.Execute(ctx, artifactapp.GetArtifactInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &artifactpb.GetArtifactResponse{
		Artifact: mappers.ArtifactToProto(output.Artifact),
	}, nil
}

// ListArtifacts lists artifacts for a world
func (h *ArtifactHandler) ListArtifacts(ctx context.Context, req *artifactpb.ListArtifactsRequest) (*artifactpb.ListArtifactsResponse, error) {
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

	output, err := h.listArtifactsUseCase.Execute(ctx, artifactapp.ListArtifactsInput{
		TenantID: tenantUUID,
		WorldID:  worldID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}

	artifacts := make([]*artifactpb.Artifact, len(output.Artifacts))
	for i, a := range output.Artifacts {
		artifacts[i] = mappers.ArtifactToProto(a)
	}

	return &artifactpb.ListArtifactsResponse{
		Artifacts:  artifacts,
		TotalCount: int32(output.Total),
	}, nil
}

// UpdateArtifact updates an existing artifact
func (h *ArtifactHandler) UpdateArtifact(ctx context.Context, req *artifactpb.UpdateArtifactRequest) (*artifactpb.UpdateArtifactResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid artifact id: %v", err)
	}

	var name *string
	if req.Name != nil {
		name = req.Name
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	var rarity *string
	if req.Rarity != nil {
		rarity = req.Rarity
	}

	var characterIDs *[]uuid.UUID
	if req.CharacterIds != nil {
		ids := make([]uuid.UUID, 0, len(req.CharacterIds.Ids))
		for _, cidStr := range req.CharacterIds.Ids {
			cid, err := uuid.Parse(cidStr)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
			}
			ids = append(ids, cid)
		}
		characterIDs = &ids
	}

	var locationIDs *[]uuid.UUID
	if req.LocationIds != nil {
		ids := make([]uuid.UUID, 0, len(req.LocationIds.Ids))
		for _, lidStr := range req.LocationIds.Ids {
			lid, err := uuid.Parse(lidStr)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid location_id: %v", err)
			}
			ids = append(ids, lid)
		}
		locationIDs = &ids
	}

	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	input := artifactapp.UpdateArtifactInput{
		TenantID:     tenantUUID,
		ID:           id,
		Name:         name,
		Description:  description,
		Rarity:       rarity,
		CharacterIDs: characterIDs,
		LocationIDs:  locationIDs,
	}

	output, err := h.updateArtifactUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &artifactpb.UpdateArtifactResponse{
		Artifact: mappers.ArtifactToProto(output.Artifact),
	}, nil
}

// DeleteArtifact deletes an artifact
func (h *ArtifactHandler) DeleteArtifact(ctx context.Context, req *artifactpb.DeleteArtifactRequest) (*artifactpb.DeleteArtifactResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid artifact id: %v", err)
	}

	err = h.deleteArtifactUseCase.Execute(ctx, artifactapp.DeleteArtifactInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &artifactpb.DeleteArtifactResponse{}, nil
}

// GetArtifactReferences retrieves all references for an artifact
func (h *ArtifactHandler) GetArtifactReferences(ctx context.Context, req *artifactpb.GetArtifactReferencesRequest) (*artifactpb.GetArtifactReferencesResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	artifactID, err := uuid.Parse(req.ArtifactId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid artifact_id: %v", err)
	}

	output, err := h.getReferencesUseCase.Execute(ctx, artifactapp.GetArtifactReferencesInput{
		TenantID:   tenantUUID,
		ArtifactID: artifactID,
	})
	if err != nil {
		return nil, err
	}

	references := make([]*artifactpb.ArtifactReference, len(output.References))
	for i, ref := range output.References {
		references[i] = mappers.ArtifactReferenceToProto(ref)
	}

	return &artifactpb.GetArtifactReferencesResponse{
		References: references,
	}, nil
}

// AddArtifactReference adds a reference to an artifact
func (h *ArtifactHandler) AddArtifactReference(ctx context.Context, req *artifactpb.AddArtifactReferenceRequest) (*artifactpb.AddArtifactReferenceResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	artifactID, err := uuid.Parse(req.ArtifactId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid artifact_id: %v", err)
	}

	entityID, err := uuid.Parse(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id: %v", err)
	}

	entityType := world.ArtifactReferenceEntityType(req.EntityType)
	if entityType != world.ArtifactReferenceEntityTypeCharacter && entityType != world.ArtifactReferenceEntityTypeLocation {
		return nil, status.Errorf(codes.InvalidArgument, "entity_type must be 'character' or 'location'")
	}

	err = h.addReferenceUseCase.Execute(ctx, artifactapp.AddArtifactReferenceInput{
		TenantID:   tenantUUID,
		ArtifactID: artifactID,
		EntityType: entityType,
		EntityID:   entityID,
	})
	if err != nil {
		return nil, err
	}

	return &artifactpb.AddArtifactReferenceResponse{}, nil
}

// RemoveArtifactReference removes a reference from an artifact
func (h *ArtifactHandler) RemoveArtifactReference(ctx context.Context, req *artifactpb.RemoveArtifactReferenceRequest) (*artifactpb.RemoveArtifactReferenceResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	artifactID, err := uuid.Parse(req.ArtifactId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid artifact_id: %v", err)
	}

	entityID, err := uuid.Parse(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id: %v", err)
	}

	entityType := world.ArtifactReferenceEntityType(req.EntityType)
	if entityType != world.ArtifactReferenceEntityTypeCharacter && entityType != world.ArtifactReferenceEntityTypeLocation {
		return nil, status.Errorf(codes.InvalidArgument, "entity_type must be 'character' or 'location'")
	}

	err = h.removeReferenceUseCase.Execute(ctx, artifactapp.RemoveArtifactReferenceInput{
		TenantID:   tenantUUID,
		ArtifactID: artifactID,
		EntityType: entityType,
		EntityID:   entityID,
	})
	if err != nil {
		return nil, err
	}

	return &artifactpb.RemoveArtifactReferenceResponse{}, nil
}

