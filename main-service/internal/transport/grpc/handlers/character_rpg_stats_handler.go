package handlers

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	characterstatsapp "github.com/story-engine/main-service/internal/application/rpg/character_stats"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	"github.com/story-engine/main-service/internal/platform/logger"
	characterrpgstatspb "github.com/story-engine/main-service/proto/character_rpg_stats"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CharacterRPGStatsHandler implements the CharacterRPGStatsService gRPC service
type CharacterRPGStatsHandler struct {
	characterrpgstatspb.UnimplementedCharacterRPGStatsServiceServer
	createStatsUseCase      *characterstatsapp.CreateCharacterStatsUseCase
	getActiveStatsUseCase    *characterstatsapp.GetActiveCharacterStatsUseCase
	listStatsHistoryUseCase  *characterstatsapp.ListCharacterStatsHistoryUseCase
	activateVersionUseCase   *characterstatsapp.ActivateCharacterStatsVersionUseCase
	deleteAllStatsUseCase    *characterstatsapp.DeleteAllCharacterStatsUseCase
	logger                   logger.Logger
}

// NewCharacterRPGStatsHandler creates a new CharacterRPGStatsHandler
func NewCharacterRPGStatsHandler(
	createStatsUseCase *characterstatsapp.CreateCharacterStatsUseCase,
	getActiveStatsUseCase *characterstatsapp.GetActiveCharacterStatsUseCase,
	listStatsHistoryUseCase *characterstatsapp.ListCharacterStatsHistoryUseCase,
	activateVersionUseCase *characterstatsapp.ActivateCharacterStatsVersionUseCase,
	deleteAllStatsUseCase *characterstatsapp.DeleteAllCharacterStatsUseCase,
	logger logger.Logger,
) *CharacterRPGStatsHandler {
	return &CharacterRPGStatsHandler{
		createStatsUseCase:     createStatsUseCase,
		getActiveStatsUseCase:   getActiveStatsUseCase,
		listStatsHistoryUseCase: listStatsHistoryUseCase,
		activateVersionUseCase:  activateVersionUseCase,
		deleteAllStatsUseCase:   deleteAllStatsUseCase,
		logger:                  logger,
	}
}

// CreateCharacterRPGStats creates a new version of character RPG stats
func (h *CharacterRPGStatsHandler) CreateCharacterRPGStats(ctx context.Context, req *characterrpgstatspb.CreateCharacterRPGStatsRequest) (*characterrpgstatspb.CreateCharacterRPGStatsResponse, error) {
	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	if req.BaseStats == "" {
		return nil, status.Errorf(codes.InvalidArgument, "base_stats is required")
	}

	var eventID *uuid.UUID
	if req.EventId != nil && *req.EventId != "" {
		eid, err := uuid.Parse(*req.EventId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
		}
		eventID = &eid
	}

	baseStats := json.RawMessage(req.BaseStats)

	var derivedStats *json.RawMessage
	if req.DerivedStats != nil && *req.DerivedStats != "" {
		stats := json.RawMessage(*req.DerivedStats)
		derivedStats = &stats
	}

	var progression *json.RawMessage
	if req.Progression != nil && *req.Progression != "" {
		prog := json.RawMessage(*req.Progression)
		progression = &prog
	}

	input := characterstatsapp.CreateCharacterStatsInput{
		CharacterID:        characterID,
		EventID:            eventID,
		BaseStats:          baseStats,
		DerivedStats:       derivedStats,
		Progression:        progression,
		Reason:             req.Reason,
		Timeline:           req.Timeline,
		DeactivatePrevious: true, // Default: deactivate previous versions
	}

	output, err := h.createStatsUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &characterrpgstatspb.CreateCharacterRPGStatsResponse{
		CharacterRpgStats: mappers.CharacterRPGStatsToProto(output.Stats),
	}, nil
}

// GetCharacterRPGStats retrieves active character RPG stats
func (h *CharacterRPGStatsHandler) GetCharacterRPGStats(ctx context.Context, req *characterrpgstatspb.GetCharacterRPGStatsRequest) (*characterrpgstatspb.GetCharacterRPGStatsResponse, error) {
	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	output, err := h.getActiveStatsUseCase.Execute(ctx, characterstatsapp.GetActiveCharacterStatsInput{
		CharacterID: characterID,
	})
	if err != nil {
		return nil, err
	}

	return &characterrpgstatspb.GetCharacterRPGStatsResponse{
		CharacterRpgStats: mappers.CharacterRPGStatsToProto(output.Stats),
	}, nil
}

// ListCharacterRPGStatsHistory lists all versions of character RPG stats
func (h *CharacterRPGStatsHandler) ListCharacterRPGStatsHistory(ctx context.Context, req *characterrpgstatspb.ListCharacterRPGStatsHistoryRequest) (*characterrpgstatspb.ListCharacterRPGStatsHistoryResponse, error) {
	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	output, err := h.listStatsHistoryUseCase.Execute(ctx, characterstatsapp.ListCharacterStatsHistoryInput{
		CharacterID: characterID,
	})
	if err != nil {
		return nil, err
	}

	stats := make([]*characterrpgstatspb.CharacterRPGStats, len(output.Stats))
	for i, s := range output.Stats {
		stats[i] = mappers.CharacterRPGStatsToProto(s)
	}

	return &characterrpgstatspb.ListCharacterRPGStatsHistoryResponse{
		CharacterRpgStats: stats,
		TotalCount:        int32(len(output.Stats)),
	}, nil
}

// ActivateCharacterRPGStatsVersion activates a specific version (rollback)
func (h *CharacterRPGStatsHandler) ActivateCharacterRPGStatsVersion(ctx context.Context, req *characterrpgstatspb.ActivateCharacterRPGStatsVersionRequest) (*characterrpgstatspb.ActivateCharacterRPGStatsVersionResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid stats id: %v", err)
	}

	output, err := h.activateVersionUseCase.Execute(ctx, characterstatsapp.ActivateCharacterStatsVersionInput{
		StatsID: id,
	})
	if err != nil {
		return nil, err
	}

	return &characterrpgstatspb.ActivateCharacterRPGStatsVersionResponse{
		CharacterRpgStats: mappers.CharacterRPGStatsToProto(output.Stats),
	}, nil
}

// DeleteCharacterRPGStats deletes all stats for a character
func (h *CharacterRPGStatsHandler) DeleteCharacterRPGStats(ctx context.Context, req *characterrpgstatspb.DeleteCharacterRPGStatsRequest) (*characterrpgstatspb.DeleteCharacterRPGStatsResponse, error) {
	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	err = h.deleteAllStatsUseCase.Execute(ctx, characterstatsapp.DeleteAllCharacterStatsInput{
		CharacterID: characterID,
	})
	if err != nil {
		return nil, err
	}

	return &characterrpgstatspb.DeleteCharacterRPGStatsResponse{}, nil
}

