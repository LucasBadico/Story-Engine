package mappers

import (
	"encoding/json"

	"github.com/story-engine/main-service/internal/core/relation"
	entityrelationpb "github.com/story-engine/main-service/proto/entity_relation"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// EntityRelationToProto converts a relation domain entity to a protobuf message
func EntityRelationToProto(r *relation.EntityRelation) *entityrelationpb.EntityRelation {
	if r == nil {
		return nil
	}

	proto := &entityrelationpb.EntityRelation{
		Id:           r.ID.String(),
		TenantId:     r.TenantID.String(),
		WorldId:      r.WorldID.String(),
		SourceType:   r.SourceType,
		SourceId:     r.SourceID.String(),
		TargetType:   r.TargetType,
		TargetId:     r.TargetID.String(),
		RelationType: r.RelationType,
		Summary:      r.Summary,
		CreatedAt:    timestamppb.New(r.CreatedAt),
		UpdatedAt:    timestamppb.New(r.UpdatedAt),
	}

	// Optional fields
	if r.ContextType != nil {
		proto.ContextType = r.ContextType
	}
	if r.ContextID != nil {
		contextID := r.ContextID.String()
		proto.ContextId = &contextID
	}
	if r.MirrorID != nil {
		mirrorID := r.MirrorID.String()
		proto.MirrorId = &mirrorID
	}
	if r.CreatedByUserID != nil {
		createdByUserID := r.CreatedByUserID.String()
		proto.CreatedByUserId = &createdByUserID
	}

	// Attributes as JSON
	if r.Attributes != nil {
		attrsJSON, err := json.Marshal(r.Attributes)
		if err == nil {
			proto.AttributesJson = string(attrsJSON)
		}
	}

	return proto
}
