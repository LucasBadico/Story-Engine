package mappers

import (
	"time"

	loreapp "github.com/story-engine/main-service/internal/application/world/lore"
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

// LoreReferenceDTOToProto converts a LoreReferenceDTO (compatibility layer) to a protobuf message
func LoreReferenceDTOToProto(dto *loreapp.LoreReferenceDTO) *lorepb.LoreReference {
	if dto == nil {
		return nil
	}

	pb := &lorepb.LoreReference{
		Id:         dto.ID.String(),
		LoreId:     dto.LoreID.String(),
		EntityType: dto.EntityType,
		EntityId:   dto.EntityID.String(),
		Notes:      dto.Notes,
	}

	if dto.RelationshipType != nil {
		pb.RelationshipType = dto.RelationshipType
	}

	// Parse CreatedAt string to timestamp
	if dto.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, dto.CreatedAt); err == nil {
			pb.CreatedAt = timestamppb.New(t)
		}
	}

	return pb
}

