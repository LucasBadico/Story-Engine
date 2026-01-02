package mappers

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	rpgsystempb "github.com/story-engine/main-service/proto/rpg_system"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// RPGSystemToProto converts an RPG system domain entity to a protobuf message
func RPGSystemToProto(s *rpg.RPGSystem) *rpgsystempb.RPGSystem {
	if s == nil {
		return nil
	}

	var tenantID *string
	if s.TenantID != nil {
		id := s.TenantID.String()
		tenantID = &id
	}

	var description *string
	if s.Description != nil {
		desc := *s.Description
		description = &desc
	}

	baseStatsSchema := string(s.BaseStatsSchema)

	var derivedStatsSchema *string
	if s.DerivedStatsSchema != nil {
		schema := string(*s.DerivedStatsSchema)
		derivedStatsSchema = &schema
	}

	var progressionSchema *string
	if s.ProgressionSchema != nil {
		schema := string(*s.ProgressionSchema)
		progressionSchema = &schema
	}

	return &rpgsystempb.RPGSystem{
		Id:                s.ID.String(),
		TenantId:          tenantID,
		Name:              s.Name,
		Description:       description,
		BaseStatsSchema:   baseStatsSchema,
		DerivedStatsSchema: derivedStatsSchema,
		ProgressionSchema: progressionSchema,
		IsBuiltin:         s.IsBuiltin,
		CreatedAt:         timestamppb.New(s.CreatedAt),
		UpdatedAt:         timestamppb.New(s.UpdatedAt),
	}
}

// RPGSystemFromProto converts a protobuf message to an RPG system domain entity
func RPGSystemFromProto(pb *rpgsystempb.RPGSystem) (*rpg.RPGSystem, error) {
	if pb == nil {
		return nil, nil
	}

	id, err := uuid.Parse(pb.Id)
	if err != nil {
		return nil, err
	}

	var tenantID *uuid.UUID
	if pb.TenantId != nil && *pb.TenantId != "" {
		tid, err := uuid.Parse(*pb.TenantId)
		if err != nil {
			return nil, err
		}
		tenantID = &tid
	}

	var description *string
	if pb.Description != nil {
		desc := *pb.Description
		description = &desc
	}

	baseStatsSchema := json.RawMessage(pb.BaseStatsSchema)

	var derivedStatsSchema *json.RawMessage
	if pb.DerivedStatsSchema != nil {
		schema := json.RawMessage(*pb.DerivedStatsSchema)
		derivedStatsSchema = &schema
	}

	var progressionSchema *json.RawMessage
	if pb.ProgressionSchema != nil {
		schema := json.RawMessage(*pb.ProgressionSchema)
		progressionSchema = &schema
	}

	system, err := rpg.NewRPGSystem(tenantID, pb.Name, baseStatsSchema)
	if err != nil {
		return nil, err
	}

	system.ID = id
	system.Description = description
	system.DerivedStatsSchema = derivedStatsSchema
	system.ProgressionSchema = progressionSchema
	system.IsBuiltin = pb.IsBuiltin

	return system, nil
}

