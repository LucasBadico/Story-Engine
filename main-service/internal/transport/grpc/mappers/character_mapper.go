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

// NOTE: CharacterRelationshipToProto was removed because world.CharacterRelationship no longer exists.
// Character relationships are now handled via entity_relations table.
// Use EntityRelationToProto from entity_relation_mapper.go instead.

