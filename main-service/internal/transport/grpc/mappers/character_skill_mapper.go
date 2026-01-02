package mappers

import (
	"github.com/story-engine/main-service/internal/core/rpg"
	characterskillpb "github.com/story-engine/main-service/proto/character_skill"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CharacterSkillToProto converts a character skill domain entity to a protobuf message
func CharacterSkillToProto(cs *rpg.CharacterSkill) *characterskillpb.CharacterSkill {
	if cs == nil {
		return nil
	}

	return &characterskillpb.CharacterSkill{
		Id:          cs.ID.String(),
		CharacterId: cs.CharacterID.String(),
		SkillId:     cs.SkillID.String(),
		Rank:        int32(cs.Rank),
		XpInSkill:   int32(cs.XPInSkill),
		IsActive:    cs.IsActive,
		AcquiredAt:  timestamppb.New(cs.AcquiredAt),
	}
}

