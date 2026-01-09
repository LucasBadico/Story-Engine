package mappers

import (
	"time"

	factionapp "github.com/story-engine/main-service/internal/application/world/faction"
	"github.com/story-engine/main-service/internal/core/world"
	factionpb "github.com/story-engine/main-service/proto/faction"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FactionToProto converts a faction domain entity to a protobuf message
func FactionToProto(f *world.Faction) *factionpb.Faction {
	if f == nil {
		return nil
	}

	pb := &factionpb.Faction{
		Id:             f.ID.String(),
		WorldId:        f.WorldID.String(),
		Name:           f.Name,
		Description:    f.Description,
		Beliefs:        f.Beliefs,
		Structure:      f.Structure,
		Symbols:        f.Symbols,
		HierarchyLevel: int32(f.HierarchyLevel),
		CreatedAt:      timestamppb.New(f.CreatedAt),
		UpdatedAt:      timestamppb.New(f.UpdatedAt),
	}

	if f.ParentID != nil {
		parentIDStr := f.ParentID.String()
		pb.ParentId = &parentIDStr
	}
	if f.Type != nil {
		pb.Type = f.Type
	}

	return pb
}

// FactionReferenceDTOToProto converts a FactionReferenceDTO (compatibility layer) to a protobuf message
func FactionReferenceDTOToProto(dto *factionapp.FactionReferenceDTO) *factionpb.FactionReference {
	if dto == nil {
		return nil
	}

	pb := &factionpb.FactionReference{
		Id:         dto.ID.String(),
		FactionId:  dto.FactionID.String(),
		EntityType: dto.EntityType,
		EntityId:   dto.EntityID.String(),
		Notes:      dto.Notes,
	}

	if dto.Role != nil {
		pb.Role = dto.Role
	}

	// Parse CreatedAt string to timestamp
	if dto.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, dto.CreatedAt); err == nil {
			pb.CreatedAt = timestamppb.New(t)
		}
	}

	return pb
}

