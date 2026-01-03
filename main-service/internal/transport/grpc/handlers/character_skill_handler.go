package handlers

import (
	"context"

	"github.com/google/uuid"
	characterskillapp "github.com/story-engine/main-service/internal/application/rpg/character_skill"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	characterskillpb "github.com/story-engine/main-service/proto/character_skill"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CharacterSkillHandler implements the CharacterSkillService gRPC service
type CharacterSkillHandler struct {
	characterskillpb.UnimplementedCharacterSkillServiceServer
	learnSkillUseCase   *characterskillapp.LearnSkillUseCase
	updateSkillUseCase  *characterskillapp.UpdateCharacterSkillUseCase
	deleteSkillUseCase  *characterskillapp.DeleteCharacterSkillUseCase
	listSkillsUseCase   *characterskillapp.ListCharacterSkillsUseCase
	logger              logger.Logger
}

// NewCharacterSkillHandler creates a new CharacterSkillHandler
func NewCharacterSkillHandler(
	learnSkillUseCase *characterskillapp.LearnSkillUseCase,
	updateSkillUseCase *characterskillapp.UpdateCharacterSkillUseCase,
	deleteSkillUseCase *characterskillapp.DeleteCharacterSkillUseCase,
	listSkillsUseCase *characterskillapp.ListCharacterSkillsUseCase,
	logger logger.Logger,
) *CharacterSkillHandler {
	return &CharacterSkillHandler{
		learnSkillUseCase:  learnSkillUseCase,
		updateSkillUseCase: updateSkillUseCase,
		deleteSkillUseCase: deleteSkillUseCase,
		listSkillsUseCase:  listSkillsUseCase,
		logger:             logger,
	}
}

// AddCharacterSkill adds a skill to a character
func (h *CharacterSkillHandler) AddCharacterSkill(ctx context.Context, req *characterskillpb.AddCharacterSkillRequest) (*characterskillpb.AddCharacterSkillResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	skillID, err := uuid.Parse(req.SkillId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid skill_id: %v", err)
	}

	output, err := h.learnSkillUseCase.Execute(ctx, characterskillapp.LearnSkillInput{
		TenantID:    tenantUUID,
		CharacterID: characterID,
		SkillID:     skillID,
	})
	if err != nil {
		return nil, err
	}

	return &characterskillpb.AddCharacterSkillResponse{
		CharacterSkill: mappers.CharacterSkillToProto(output.CharacterSkill),
	}, nil
}

// UpdateCharacterSkill updates a character's skill
func (h *CharacterSkillHandler) UpdateCharacterSkill(ctx context.Context, req *characterskillpb.UpdateCharacterSkillRequest) (*characterskillpb.UpdateCharacterSkillResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_skill id: %v", err)
	}

	var rank *int
	if req.Rank != nil && *req.Rank > 0 {
		r := int(*req.Rank)
		rank = &r
	}

	var addXP *int
	if req.XpInSkill != nil {
		xp := int(*req.XpInSkill)
		addXP = &xp
	}

	var isActive *bool
	if req.IsActive != nil {
		isActive = req.IsActive
	}

	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	input := characterskillapp.UpdateCharacterSkillInput{
		TenantID: tenantUUID,
		ID:       id,
		Rank:     rank,
		AddXP:    addXP,
		IsActive: isActive,
	}

	output, err := h.updateSkillUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &characterskillpb.UpdateCharacterSkillResponse{
		CharacterSkill: mappers.CharacterSkillToProto(output.CharacterSkill),
	}, nil
}

// RemoveCharacterSkill removes a skill from a character
func (h *CharacterSkillHandler) RemoveCharacterSkill(ctx context.Context, req *characterskillpb.RemoveCharacterSkillRequest) (*characterskillpb.RemoveCharacterSkillResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_skill id: %v", err)
	}

	err = h.deleteSkillUseCase.Execute(ctx, characterskillapp.DeleteCharacterSkillInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &characterskillpb.RemoveCharacterSkillResponse{}, nil
}

// ListCharacterSkills lists skills for a character
func (h *CharacterSkillHandler) ListCharacterSkills(ctx context.Context, req *characterskillpb.ListCharacterSkillsRequest) (*characterskillpb.ListCharacterSkillsResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	output, err := h.listSkillsUseCase.Execute(ctx, characterskillapp.ListCharacterSkillsInput{
		TenantID:    tenantUUID,
		CharacterID: characterID,
		ActiveOnly:  false,
	})
	if err != nil {
		return nil, err
	}

	characterSkills := make([]*characterskillpb.CharacterSkill, len(output.Skills))
	for i, cs := range output.Skills {
		characterSkills[i] = mappers.CharacterSkillToProto(cs)
	}

	return &characterskillpb.ListCharacterSkillsResponse{
		CharacterSkills: characterSkills,
		TotalCount:      int32(len(output.Skills)),
	}, nil
}

