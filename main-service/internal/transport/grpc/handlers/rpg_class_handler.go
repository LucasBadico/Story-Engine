package handlers

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	rpgclassapp "github.com/story-engine/main-service/internal/application/rpg/rpg_class"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	"github.com/story-engine/main-service/internal/platform/logger"
	rpgclasspb "github.com/story-engine/main-service/proto/rpg_class"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RPGClassHandler implements the RPGClassService gRPC service
type RPGClassHandler struct {
	rpgclasspb.UnimplementedRPGClassServiceServer
	createRPGClassUseCase     *rpgclassapp.CreateRPGClassUseCase
	getRPGClassUseCase        *rpgclassapp.GetRPGClassUseCase
	listRPGClassesUseCase     *rpgclassapp.ListRPGClassesUseCase
	updateRPGClassUseCase     *rpgclassapp.UpdateRPGClassUseCase
	deleteRPGClassUseCase     *rpgclassapp.DeleteRPGClassUseCase
	addSkillToClassUseCase    *rpgclassapp.AddSkillToClassUseCase
	listClassSkillsUseCase    *rpgclassapp.ListClassSkillsUseCase
	removeSkillFromClassUseCase *rpgclassapp.RemoveSkillFromClassUseCase
	logger                    logger.Logger
}

// NewRPGClassHandler creates a new RPGClassHandler
func NewRPGClassHandler(
	createRPGClassUseCase *rpgclassapp.CreateRPGClassUseCase,
	getRPGClassUseCase *rpgclassapp.GetRPGClassUseCase,
	listRPGClassesUseCase *rpgclassapp.ListRPGClassesUseCase,
	updateRPGClassUseCase *rpgclassapp.UpdateRPGClassUseCase,
	deleteRPGClassUseCase *rpgclassapp.DeleteRPGClassUseCase,
	addSkillToClassUseCase *rpgclassapp.AddSkillToClassUseCase,
	listClassSkillsUseCase *rpgclassapp.ListClassSkillsUseCase,
	removeSkillFromClassUseCase *rpgclassapp.RemoveSkillFromClassUseCase,
	logger logger.Logger,
) *RPGClassHandler {
	return &RPGClassHandler{
		createRPGClassUseCase:      createRPGClassUseCase,
		getRPGClassUseCase:         getRPGClassUseCase,
		listRPGClassesUseCase:      listRPGClassesUseCase,
		updateRPGClassUseCase:      updateRPGClassUseCase,
		deleteRPGClassUseCase:      deleteRPGClassUseCase,
		addSkillToClassUseCase:      addSkillToClassUseCase,
		listClassSkillsUseCase:      listClassSkillsUseCase,
		removeSkillFromClassUseCase: removeSkillFromClassUseCase,
		logger:                     logger,
	}
}

// CreateRPGClass creates a new RPG class
func (h *RPGClassHandler) CreateRPGClass(ctx context.Context, req *rpgclasspb.CreateRPGClassRequest) (*rpgclasspb.CreateRPGClassResponse, error) {
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

	var parentClassID *uuid.UUID
	if req.ParentClassId != nil && *req.ParentClassId != "" {
		pid, err := uuid.Parse(*req.ParentClassId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid parent_class_id: %v", err)
		}
		parentClassID = &pid
	}

	var tier *int
	if req.Tier != nil && *req.Tier > 0 {
		t := int(*req.Tier)
		tier = &t
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	var requirements *json.RawMessage
	if req.Requirements != nil && *req.Requirements != "" {
		reqJSON := json.RawMessage(*req.Requirements)
		requirements = &reqJSON
	}

	var statBonuses *json.RawMessage
	if req.StatBonuses != nil && *req.StatBonuses != "" {
		bonuses := json.RawMessage(*req.StatBonuses)
		statBonuses = &bonuses
	}

	input := rpgclassapp.CreateRPGClassInput{
		TenantID:      tenantUUID,
		RPGSystemID:   rpgSystemID,
		ParentClassID: parentClassID,
		Name:          req.Name,
		Tier:          tier,
		Description:   description,
		Requirements:  requirements,
		StatBonuses:   statBonuses,
	}

	output, err := h.createRPGClassUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &rpgclasspb.CreateRPGClassResponse{
		RpgClass: mappers.RPGClassToProto(output.Class),
	}, nil
}

// GetRPGClass retrieves an RPG class by ID
func (h *RPGClassHandler) GetRPGClass(ctx context.Context, req *rpgclasspb.GetRPGClassRequest) (*rpgclasspb.GetRPGClassResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid rpg_class id: %v", err)
	}

	output, err := h.getRPGClassUseCase.Execute(ctx, rpgclassapp.GetRPGClassInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &rpgclasspb.GetRPGClassResponse{
		RpgClass: mappers.RPGClassToProto(output.Class),
	}, nil
}

// ListRPGClasses lists RPG classes for an RPG system
func (h *RPGClassHandler) ListRPGClasses(ctx context.Context, req *rpgclasspb.ListRPGClassesRequest) (*rpgclasspb.ListRPGClassesResponse, error) {
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

	output, err := h.listRPGClassesUseCase.Execute(ctx, rpgclassapp.ListRPGClassesInput{
		TenantID:    tenantUUID,
		RPGSystemID: rpgSystemID,
	})
	if err != nil {
		return nil, err
	}

	rpgClasses := make([]*rpgclasspb.RPGClass, len(output.Classes))
	for i, c := range output.Classes {
		rpgClasses[i] = mappers.RPGClassToProto(c)
	}

	return &rpgclasspb.ListRPGClassesResponse{
		RpgClasses: rpgClasses,
		TotalCount: int32(len(output.Classes)),
	}, nil
}

// UpdateRPGClass updates an existing RPG class
func (h *RPGClassHandler) UpdateRPGClass(ctx context.Context, req *rpgclasspb.UpdateRPGClassRequest) (*rpgclasspb.UpdateRPGClassResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid rpg_class id: %v", err)
	}

	var name *string
	if req.Name != nil {
		name = req.Name
	}

	var parentClassID *uuid.UUID
	if req.ParentClassId != nil && *req.ParentClassId != "" {
		pid, err := uuid.Parse(*req.ParentClassId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid parent_class_id: %v", err)
		}
		parentClassID = &pid
	}

	var tier *int
	if req.Tier != nil && *req.Tier > 0 {
		t := int(*req.Tier)
		tier = &t
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	var requirements *json.RawMessage
	if req.Requirements != nil && *req.Requirements != "" {
		reqJSON := json.RawMessage(*req.Requirements)
		requirements = &reqJSON
	}

	var statBonuses *json.RawMessage
	if req.StatBonuses != nil && *req.StatBonuses != "" {
		bonuses := json.RawMessage(*req.StatBonuses)
		statBonuses = &bonuses
	}

	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	input := rpgclassapp.UpdateRPGClassInput{
		TenantID:      tenantUUID,
		ID:            id,
		Name:          name,
		ParentClassID: parentClassID,
		Tier:          tier,
		Description:   description,
		Requirements:  requirements,
		StatBonuses:   statBonuses,
	}

	output, err := h.updateRPGClassUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &rpgclasspb.UpdateRPGClassResponse{
		RpgClass: mappers.RPGClassToProto(output.Class),
	}, nil
}

// DeleteRPGClass deletes an RPG class
func (h *RPGClassHandler) DeleteRPGClass(ctx context.Context, req *rpgclasspb.DeleteRPGClassRequest) (*rpgclasspb.DeleteRPGClassResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid rpg_class id: %v", err)
	}

	err = h.deleteRPGClassUseCase.Execute(ctx, rpgclassapp.DeleteRPGClassInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &rpgclasspb.DeleteRPGClassResponse{}, nil
}

// AddSkillToRPGClass adds a skill to an RPG class
func (h *RPGClassHandler) AddSkillToRPGClass(ctx context.Context, req *rpgclasspb.AddSkillToRPGClassRequest) (*rpgclasspb.AddSkillToRPGClassResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	classID, err := uuid.Parse(req.RpgClassId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid rpg_class_id: %v", err)
	}

	skillID, err := uuid.Parse(req.SkillId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid skill_id: %v", err)
	}

	_, err = h.addSkillToClassUseCase.Execute(ctx, rpgclassapp.AddSkillToClassInput{
		TenantID:    tenantUUID,
		ClassID:     classID,
		SkillID:     skillID,
		UnlockLevel: 1, // Default unlock level
	})
	if err != nil {
		return nil, err
	}

	return &rpgclasspb.AddSkillToRPGClassResponse{}, nil
}

// RemoveSkillFromRPGClass removes a skill from an RPG class
func (h *RPGClassHandler) RemoveSkillFromRPGClass(ctx context.Context, req *rpgclasspb.RemoveSkillFromRPGClassRequest) (*rpgclasspb.RemoveSkillFromRPGClassResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	classID, err := uuid.Parse(req.RpgClassId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid rpg_class_id: %v", err)
	}

	skillID, err := uuid.Parse(req.SkillId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid skill_id: %v", err)
	}

	_, err = h.removeSkillFromClassUseCase.Execute(ctx, rpgclassapp.RemoveSkillFromClassInput{
		TenantID: tenantUUID,
		ClassID:  classID,
		SkillID:  skillID,
	})
	if err != nil {
		return nil, err
	}

	return &rpgclasspb.RemoveSkillFromRPGClassResponse{}, nil
}

// ListRPGClassSkills lists skills for an RPG class
func (h *RPGClassHandler) ListRPGClassSkills(ctx context.Context, req *rpgclasspb.ListRPGClassSkillsRequest) (*rpgclasspb.ListRPGClassSkillsResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	classID, err := uuid.Parse(req.RpgClassId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid rpg_class_id: %v", err)
	}

	output, err := h.listClassSkillsUseCase.Execute(ctx, rpgclassapp.ListClassSkillsInput{
		TenantID: tenantUUID,
		ClassID:  classID,
	})
	if err != nil {
		return nil, err
	}

	skillIDs := make([]string, len(output.Skills))
	for i, s := range output.Skills {
		skillIDs[i] = s.SkillID.String()
	}

	return &rpgclasspb.ListRPGClassSkillsResponse{
		SkillIds:   skillIDs,
		TotalCount: int32(len(output.Skills)),
	}, nil
}

