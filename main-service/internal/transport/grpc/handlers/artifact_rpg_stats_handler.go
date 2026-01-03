package handlers

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	artifactstatsapp "github.com/story-engine/main-service/internal/application/rpg/artifact_stats"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	artifactrpgstatspb "github.com/story-engine/main-service/proto/artifact_rpg_stats"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ArtifactRPGStatsHandler implements the ArtifactRPGStatsService gRPC service
type ArtifactRPGStatsHandler struct {
	artifactrpgstatspb.UnimplementedArtifactRPGStatsServiceServer
	createStatsUseCase     *artifactstatsapp.CreateArtifactStatsUseCase
	getActiveStatsUseCase  *artifactstatsapp.GetActiveArtifactStatsUseCase
	listStatsHistoryUseCase *artifactstatsapp.ListArtifactStatsHistoryUseCase
	activateVersionUseCase *artifactstatsapp.ActivateArtifactStatsVersionUseCase
	logger                 logger.Logger
}

// NewArtifactRPGStatsHandler creates a new ArtifactRPGStatsHandler
func NewArtifactRPGStatsHandler(
	createStatsUseCase *artifactstatsapp.CreateArtifactStatsUseCase,
	getActiveStatsUseCase *artifactstatsapp.GetActiveArtifactStatsUseCase,
	listStatsHistoryUseCase *artifactstatsapp.ListArtifactStatsHistoryUseCase,
	activateVersionUseCase *artifactstatsapp.ActivateArtifactStatsVersionUseCase,
	logger logger.Logger,
) *ArtifactRPGStatsHandler {
	return &ArtifactRPGStatsHandler{
		createStatsUseCase:      createStatsUseCase,
		getActiveStatsUseCase:   getActiveStatsUseCase,
		listStatsHistoryUseCase: listStatsHistoryUseCase,
		activateVersionUseCase:  activateVersionUseCase,
		logger:                  logger,
	}
}

// CreateArtifactRPGStats creates a new version of artifact RPG stats
func (h *ArtifactRPGStatsHandler) CreateArtifactRPGStats(ctx context.Context, req *artifactrpgstatspb.CreateArtifactRPGStatsRequest) (*artifactrpgstatspb.CreateArtifactRPGStatsResponse, error) {
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

	if req.Stats == "" {
		return nil, status.Errorf(codes.InvalidArgument, "stats is required")
	}

	var eventID *uuid.UUID
	if req.EventId != nil && *req.EventId != "" {
		eid, err := uuid.Parse(*req.EventId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
		}
		eventID = &eid
	}

	stats := json.RawMessage(req.Stats)

	input := artifactstatsapp.CreateArtifactStatsInput{
		TenantID:           tenantUUID,
		ArtifactID:         artifactID,
		EventID:            eventID,
		Stats:              stats,
		Reason:             req.Reason,
		Timeline:           req.Timeline,
		DeactivatePrevious: true, // Default: deactivate previous versions
	}

	output, err := h.createStatsUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &artifactrpgstatspb.CreateArtifactRPGStatsResponse{
		ArtifactRpgStats: mappers.ArtifactRPGStatsToProto(output.Stats),
	}, nil
}

// GetArtifactRPGStats retrieves active artifact RPG stats
func (h *ArtifactRPGStatsHandler) GetArtifactRPGStats(ctx context.Context, req *artifactrpgstatspb.GetArtifactRPGStatsRequest) (*artifactrpgstatspb.GetArtifactRPGStatsResponse, error) {
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

	output, err := h.getActiveStatsUseCase.Execute(ctx, artifactstatsapp.GetActiveArtifactStatsInput{
		TenantID:   tenantUUID,
		ArtifactID: artifactID,
	})
	if err != nil {
		return nil, err
	}

	return &artifactrpgstatspb.GetArtifactRPGStatsResponse{
		ArtifactRpgStats: mappers.ArtifactRPGStatsToProto(output.Stats),
	}, nil
}

// ListArtifactRPGStatsHistory lists all versions of artifact RPG stats
func (h *ArtifactRPGStatsHandler) ListArtifactRPGStatsHistory(ctx context.Context, req *artifactrpgstatspb.ListArtifactRPGStatsHistoryRequest) (*artifactrpgstatspb.ListArtifactRPGStatsHistoryResponse, error) {
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

	output, err := h.listStatsHistoryUseCase.Execute(ctx, artifactstatsapp.ListArtifactStatsHistoryInput{
		TenantID:   tenantUUID,
		ArtifactID: artifactID,
	})
	if err != nil {
		return nil, err
	}

	stats := make([]*artifactrpgstatspb.ArtifactRPGStats, len(output.Stats))
	for i, s := range output.Stats {
		stats[i] = mappers.ArtifactRPGStatsToProto(s)
	}

	return &artifactrpgstatspb.ListArtifactRPGStatsHistoryResponse{
		ArtifactRpgStats: stats,
		TotalCount:       int32(len(output.Stats)),
	}, nil
}

// ActivateArtifactRPGStatsVersion activates a specific version (rollback)
func (h *ArtifactRPGStatsHandler) ActivateArtifactRPGStatsVersion(ctx context.Context, req *artifactrpgstatspb.ActivateArtifactRPGStatsVersionRequest) (*artifactrpgstatspb.ActivateArtifactRPGStatsVersionResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid stats id: %v", err)
	}

	output, err := h.activateVersionUseCase.Execute(ctx, artifactstatsapp.ActivateArtifactStatsVersionInput{
		TenantID: tenantUUID,
		StatsID:  id,
	})
	if err != nil {
		return nil, err
	}

	return &artifactrpgstatspb.ActivateArtifactRPGStatsVersionResponse{
		ArtifactRpgStats: mappers.ArtifactRPGStatsToProto(output.Stats),
	}, nil
}

