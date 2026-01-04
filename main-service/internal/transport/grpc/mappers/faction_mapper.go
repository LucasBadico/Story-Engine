package mappers

import (
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

// FactionReferenceToProto converts a faction-reference relationship to a protobuf message
func FactionReferenceToProto(fr *world.FactionReference) *factionpb.FactionReference {
	if fr == nil {
		return nil
	}

	pb := &factionpb.FactionReference{
		Id:         fr.ID.String(),
		FactionId:  fr.FactionID.String(),
		EntityType: fr.EntityType,
		EntityId:   fr.EntityID.String(),
		Notes:      fr.Notes,
		CreatedAt:  timestamppb.New(fr.CreatedAt),
	}

	if fr.Role != nil {
		pb.Role = fr.Role
	}

	return pb
}

