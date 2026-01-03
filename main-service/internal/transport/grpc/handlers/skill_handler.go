package handlers

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	skillapp "github.com/story-engine/main-service/internal/application/rpg/skill"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	skillpb "github.com/story-engine/main-service/proto/skill"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SkillHandler implements the SkillService gRPC service
type SkillHandler struct {
	skillpb.UnimplementedSkillServiceServer
	createSkillUseCase *skillapp.CreateSkillUseCase
	getSkillUseCase    *skillapp.GetSkillUseCase
	listSkillsUseCase  *skillapp.ListSkillsUseCase
	updateSkillUseCase *skillapp.UpdateSkillUseCase
	deleteSkillUseCase *skillapp.DeleteSkillUseCase
	logger             logger.Logger
}

// NewSkillHandler creates a new SkillHandler
func NewSkillHandler(
	createSkillUseCase *skillapp.CreateSkillUseCase,
	getSkillUseCase *skillapp.GetSkillUseCase,
	listSkillsUseCase *skillapp.ListSkillsUseCase,
	updateSkillUseCase *skillapp.UpdateSkillUseCase,
	deleteSkillUseCase *skillapp.DeleteSkillUseCase,
	logger logger.Logger,
) *SkillHandler {
	return &SkillHandler{
		createSkillUseCase: createSkillUseCase,
		getSkillUseCase:    getSkillUseCase,
		listSkillsUseCase:  listSkillsUseCase,
		updateSkillUseCase: updateSkillUseCase,
		deleteSkillUseCase: deleteSkillUseCase,
		logger:             logger,
	}
}

// CreateSkill creates a new skill
func (h *SkillHandler) CreateSkill(ctx context.Context, req *skillpb.CreateSkillRequest) (*skillpb.CreateSkillResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	rpgSystemID, err := uuid.Parse(req.RpgSystemId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid rpg_system_id: %v", err)
	}

	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}

	var category *rpg.SkillCategory
	if req.Category != nil && *req.Category != "" {
		cat := rpg.SkillCategory(*req.Category)
		category = &cat
	}

	var skillType *rpg.SkillType
	if req.Type != nil && *req.Type != "" {
		st := rpg.SkillType(*req.Type)
		skillType = &st
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	var prerequisites *json.RawMessage
	if req.Prerequisites != nil && *req.Prerequisites != "" {
		prereq := json.RawMessage(*req.Prerequisites)
		prerequisites = &prereq
	}

	var maxRank *int
	if req.MaxRank != nil && *req.MaxRank > 0 {
		mr := int(*req.MaxRank)
		maxRank = &mr
	}

	var effectsSchema *json.RawMessage
	if req.EffectsSchema != nil && *req.EffectsSchema != "" {
		effects := json.RawMessage(*req.EffectsSchema)
		effectsSchema = &effects
	}

	input := skillapp.CreateSkillInput{
		TenantID:      tenantUUID,
		RPGSystemID:   rpgSystemID,
		Name:          req.Name,
		Category:      category,
		Type:          skillType,
		Description:   description,
		Prerequisites: prerequisites,
		MaxRank:       maxRank,
		EffectsSchema: effectsSchema,
	}

	output, err := h.createSkillUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &skillpb.CreateSkillResponse{
		Skill: mappers.SkillToProto(output.Skill),
	}, nil
}

// GetSkill retrieves a skill by ID
func (h *SkillHandler) GetSkill(ctx context.Context, req *skillpb.GetSkillRequest) (*skillpb.GetSkillResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid skill id: %v", err)
	}

	output, err := h.getSkillUseCase.Execute(ctx, skillapp.GetSkillInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &skillpb.GetSkillResponse{
		Skill: mappers.SkillToProto(output.Skill),
	}, nil
}

// ListSkills lists skills for an RPG system
func (h *SkillHandler) ListSkills(ctx context.Context, req *skillpb.ListSkillsRequest) (*skillpb.ListSkillsResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	rpgSystemID, err := uuid.Parse(req.RpgSystemId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid rpg_system_id: %v", err)
	}

	output, err := h.listSkillsUseCase.Execute(ctx, skillapp.ListSkillsInput{
		TenantID:    tenantUUID,
		RPGSystemID: rpgSystemID,
	})
	if err != nil {
		return nil, err
	}

	skills := make([]*skillpb.Skill, len(output.Skills))
	for i, s := range output.Skills {
		skills[i] = mappers.SkillToProto(s)
	}

	return &skillpb.ListSkillsResponse{
		Skills:     skills,
		TotalCount: int32(len(output.Skills)),
	}, nil
}

// UpdateSkill updates an existing skill
func (h *SkillHandler) UpdateSkill(ctx context.Context, req *skillpb.UpdateSkillRequest) (*skillpb.UpdateSkillResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid skill id: %v", err)
	}

	var name *string
	if req.Name != nil {
		name = req.Name
	}

	var category *rpg.SkillCategory
	if req.Category != nil && *req.Category != "" {
		cat := rpg.SkillCategory(*req.Category)
		category = &cat
	}

	var skillType *rpg.SkillType
	if req.Type != nil && *req.Type != "" {
		st := rpg.SkillType(*req.Type)
		skillType = &st
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	var prerequisites *json.RawMessage
	if req.Prerequisites != nil && *req.Prerequisites != "" {
		prereq := json.RawMessage(*req.Prerequisites)
		prerequisites = &prereq
	}

	var maxRank *int
	if req.MaxRank != nil && *req.MaxRank > 0 {
		mr := int(*req.MaxRank)
		maxRank = &mr
	}

	var effectsSchema *json.RawMessage
	if req.EffectsSchema != nil && *req.EffectsSchema != "" {
		effects := json.RawMessage(*req.EffectsSchema)
		effectsSchema = &effects
	}

	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	input := skillapp.UpdateSkillInput{
		TenantID:      tenantUUID,
		ID:            id,
		Name:          name,
		Category:      category,
		Type:          skillType,
		Description:   description,
		Prerequisites: prerequisites,
		MaxRank:       maxRank,
		EffectsSchema: effectsSchema,
	}

	output, err := h.updateSkillUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &skillpb.UpdateSkillResponse{
		Skill: mappers.SkillToProto(output.Skill),
	}, nil
}

// DeleteSkill deletes a skill
func (h *SkillHandler) DeleteSkill(ctx context.Context, req *skillpb.DeleteSkillRequest) (*skillpb.DeleteSkillResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid skill id: %v", err)
	}

	err = h.deleteSkillUseCase.Execute(ctx, skillapp.DeleteSkillInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &skillpb.DeleteSkillResponse{}, nil
}

