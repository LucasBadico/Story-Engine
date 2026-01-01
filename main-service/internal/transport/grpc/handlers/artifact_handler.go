package handlers

import (
	"context"

	"github.com/google/uuid"
	artifactapp "github.com/story-engine/main-service/internal/application/world/artifact"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	"github.com/story-engine/main-service/internal/platform/logger"
	artifactpb "github.com/story-engine/main-service/proto/artifact"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ArtifactHandler implements the ArtifactService gRPC service
type ArtifactHandler struct {
	artifactpb.UnimplementedArtifactServiceServer
	createArtifactUseCase *artifactapp.CreateArtifactUseCase
	getArtifactUseCase    *artifactapp.GetArtifactUseCase
	listArtifactsUseCase  *artifactapp.ListArtifactsUseCase
	updateArtifactUseCase *artifactapp.UpdateArtifactUseCase
	deleteArtifactUseCase *artifactapp.DeleteArtifactUseCase
	logger                logger.Logger
}

// NewArtifactHandler creates a new ArtifactHandler
func NewArtifactHandler(
	createArtifactUseCase *artifactapp.CreateArtifactUseCase,
	getArtifactUseCase *artifactapp.GetArtifactUseCase,
	listArtifactsUseCase *artifactapp.ListArtifactsUseCase,
	updateArtifactUseCase *artifactapp.UpdateArtifactUseCase,
	deleteArtifactUseCase *artifactapp.DeleteArtifactUseCase,
	logger logger.Logger,
) *ArtifactHandler {
	return &ArtifactHandler{
		createArtifactUseCase: createArtifactUseCase,
		getArtifactUseCase:    getArtifactUseCase,
		listArtifactsUseCase:  listArtifactsUseCase,
		updateArtifactUseCase: updateArtifactUseCase,
		deleteArtifactUseCase: deleteArtifactUseCase,
		logger:                logger,
	}
}

// CreateArtifact creates a new artifact
func (h *ArtifactHandler) CreateArtifact(ctx context.Context, req *artifactpb.CreateArtifactRequest) (*artifactpb.CreateArtifactResponse, error) {
	worldID, err := uuid.Parse(req.WorldId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid world_id: %v", err)
	}

	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}

	var characterID *uuid.UUID
	if req.CharacterId != "" {
		cid, err := uuid.Parse(req.CharacterId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
		}
		characterID = &cid
	}

	var locationID *uuid.UUID
	if req.LocationId != "" {
		lid, err := uuid.Parse(req.LocationId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid location_id: %v", err)
		}
		locationID = &lid
	}

	input := artifactapp.CreateArtifactInput{
		WorldID:     worldID,
		CharacterID: characterID,
		LocationID:  locationID,
		Name:        req.Name,
		Description: req.Description,
		Rarity:      req.Rarity,
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
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid artifact id: %v", err)
	}

	output, err := h.getArtifactUseCase.Execute(ctx, artifactapp.GetArtifactInput{
		ID: id,
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
		WorldID: worldID,
		Limit:   limit,
		Offset:  offset,
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

	var characterID *uuid.UUID
	if req.CharacterId != nil {
		if *req.CharacterId == "" {
			characterID = nil
		} else {
			cid, err := uuid.Parse(*req.CharacterId)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
			}
			characterID = &cid
		}
	}

	var locationID *uuid.UUID
	if req.LocationId != nil {
		if *req.LocationId == "" {
			locationID = nil
		} else {
			lid, err := uuid.Parse(*req.LocationId)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid location_id: %v", err)
			}
			locationID = &lid
		}
	}

	input := artifactapp.UpdateArtifactInput{
		ID:          id,
		Name:        name,
		Description: description,
		Rarity:      rarity,
		CharacterID: characterID,
		LocationID:  locationID,
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
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid artifact id: %v", err)
	}

	err = h.deleteArtifactUseCase.Execute(ctx, artifactapp.DeleteArtifactInput{
		ID: id,
	})
	if err != nil {
		return nil, err
	}

	return &artifactpb.DeleteArtifactResponse{}, nil
}

