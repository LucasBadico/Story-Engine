package mappers

import (
	"github.com/story-engine/main-service/internal/core/world"
	archetypepb "github.com/story-engine/main-service/proto/archetype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ArchetypeToProto converts an archetype domain entity to a protobuf message
func ArchetypeToProto(a *world.Archetype) *archetypepb.Archetype {
	if a == nil {
		return nil
	}

	return &archetypepb.Archetype{
		Id:          a.ID.String(),
		TenantId:    a.TenantID.String(),
		Name:        a.Name,
		Description: a.Description,
		CreatedAt:   timestamppb.New(a.CreatedAt),
		UpdatedAt:   timestamppb.New(a.UpdatedAt),
	}
}

// ArchetypeTraitToProto converts an archetype trait domain entity to a protobuf message
func ArchetypeTraitToProto(at *world.ArchetypeTrait, trait *world.Trait) *archetypepb.ArchetypeTrait {
	if at == nil {
		return nil
	}

	pb := &archetypepb.ArchetypeTrait{
		Id:           at.ID.String(),
		ArchetypeId:  at.ArchetypeID.String(),
		TraitId:      at.TraitID.String(),
		DefaultValue: at.DefaultValue,
		CreatedAt:    timestamppb.New(at.CreatedAt),
	}

	// Include trait information if provided
	if trait != nil {
		pb.TraitName = trait.Name
		pb.TraitCategory = trait.Category
		pb.TraitDescription = trait.Description
	}

	return pb
}


