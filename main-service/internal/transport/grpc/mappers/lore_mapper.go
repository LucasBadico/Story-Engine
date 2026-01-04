package mappers

import (
	"github.com/story-engine/main-service/internal/core/world"
	lorepb "github.com/story-engine/main-service/proto/lore"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// LoreToProto converts a lore domain entity to a protobuf message
func LoreToProto(l *world.Lore) *lorepb.Lore {
	if l == nil {
		return nil
	}

	pb := &lorepb.Lore{
		Id:             l.ID.String(),
		WorldId:        l.WorldID.String(),
		Name:           l.Name,
		Description:    l.Description,
		Rules:          l.Rules,
		Limitations:    l.Limitations,
		Requirements:   l.Requirements,
		HierarchyLevel: int32(l.HierarchyLevel),
		CreatedAt:      timestamppb.New(l.CreatedAt),
		UpdatedAt:      timestamppb.New(l.UpdatedAt),
	}

	if l.ParentID != nil {
		parentIDStr := l.ParentID.String()
		pb.ParentId = &parentIDStr
	}
	if l.Category != nil {
		pb.Category = l.Category
	}

	return pb
}

// LoreReferenceToProto converts a lore-reference relationship to a protobuf message
func LoreReferenceToProto(lr *world.LoreReference) *lorepb.LoreReference {
	if lr == nil {
		return nil
	}

	pb := &lorepb.LoreReference{
		Id:         lr.ID.String(),
		LoreId:     lr.LoreID.String(),
		EntityType: lr.EntityType,
		EntityId:   lr.EntityID.String(),
		Notes:      lr.Notes,
		CreatedAt:  timestamppb.New(lr.CreatedAt),
	}

	if lr.RelationshipType != nil {
		pb.RelationshipType = lr.RelationshipType
	}

	return pb
}

