package mappers

import (
	"github.com/story-engine/main-service/internal/core/world"
	characterpb "github.com/story-engine/main-service/proto/character"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CharacterToProto converts a character domain entity to a protobuf message
func CharacterToProto(c *world.Character) *characterpb.Character {
	if c == nil {
		return nil
	}

	pb := &characterpb.Character{
		Id:          c.ID.String(),
		WorldId:     c.WorldID.String(),
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   timestamppb.New(c.CreatedAt),
		UpdatedAt:   timestamppb.New(c.UpdatedAt),
	}

	if c.ArchetypeID != nil {
		pb.ArchetypeId = c.ArchetypeID.String()
	}

	return pb
}

// CharacterTraitToProto converts a character trait domain entity to a protobuf message
func CharacterTraitToProto(ct *world.CharacterTrait) *characterpb.CharacterTrait {
	if ct == nil {
		return nil
	}

	return &characterpb.CharacterTrait{
		Id:               ct.ID.String(),
		CharacterId:      ct.CharacterID.String(),
		TraitId:          ct.TraitID.String(),
		TraitName:        ct.TraitName,
		TraitCategory:    ct.TraitCategory,
		TraitDescription: ct.TraitDescription,
		Value:            ct.Value,
		Notes:            ct.Notes,
		CreatedAt:        timestamppb.New(ct.CreatedAt),
		UpdatedAt:        timestamppb.New(ct.UpdatedAt),
	}
}

// CharacterRelationshipToProto converts a character relationship domain entity to a protobuf message
func CharacterRelationshipToProto(cr *world.CharacterRelationship) *characterpb.CharacterRelationship {
	if cr == nil {
		return nil
	}

	return &characterpb.CharacterRelationship{
		Id:              cr.ID.String(),
		Character1Id:    cr.Character1ID.String(),
		Character2Id:    cr.Character2ID.String(),
		RelationshipType: cr.RelationshipType,
		Description:     cr.Description,
		Bidirectional:    cr.Bidirectional,
		CreatedAt:        timestamppb.New(cr.CreatedAt),
		UpdatedAt:        timestamppb.New(cr.UpdatedAt),
	}
}


